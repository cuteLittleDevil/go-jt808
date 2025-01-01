package model

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"math"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	type Handler interface {
		Parse(*jt808.JTMessage) error
		String() string
		Encode() []byte
	}
	type args struct {
		msg string
		Handler
		bodyLens []int // 用于覆盖率100测试 强制替换了解析正确的body
	}
	tests := []struct {
		name   string
		args   args
		fields Handler
	}{
		{
			name: "T0x0001 终端-通用应答",
			args: args{
				msg:      "7e000100050123456789017fff007b01c803bd7e",
				Handler:  &T0x0001{},
				bodyLens: []int{4},
			},
			fields: &T0x0001{
				SerialNumber: 123,
				ID:           456,
				Result:       3,
			},
		},
		{
			name: "P0x8001 平台-通用应答",
			args: args{
				msg:      "7e8001000501234567890100007fff0002008e7e",
				Handler:  &P0x8001{},
				bodyLens: []int{4},
			},
			fields: &P0x8001{
				RespondSerialNumber: 32767,
				RespondID:           2,
				Result:              0,
			},
		},
		{
			name: "P0x8100 平台-注册消息应答",
			args: args{
				msg:      "7e8100000e01234567890100000000003132333435363738393031377e",
				Handler:  &P0x8100{},
				bodyLens: []int{2},
			},
			fields: &P0x8100{
				RespondSerialNumber: 0,
				Result:              0,
				AuthCode:            "12345678901",
			},
		},
		{
			name: "T0x0002 终端-心跳",
			args: args{
				msg:     "7e0002000001234567890100008a7e",
				Handler: &T0x0002{},
			},
			fields: &T0x0002{},
		},
		{
			name: "T0x0102 注册-鉴权 2013版本",
			args: args{
				msg:     "7e0102000b01234567890100003137323939383431373338b57e",
				Handler: &T0x0102{},
			},
			fields: &T0x0102{
				AuthCodeLen:     0,
				AuthCode:        "17299841738",
				TerminalIMEI:    "",
				SoftwareVersion: "",
				Version:         consts.JT808Protocol2013,
			},
		},
		{
			name: "T0x0102 注册-鉴权 2019版本",
			args: args{
				msg:      "7e0102402f010000000001729984173800000b3137323939383431373338313233343536373839303132333435332e372e31350000000000000000000000000000227e",
				Handler:  &T0x0102{},
				bodyLens: []int{35, 37},
			},
			fields: &T0x0102{
				AuthCodeLen:     uint8(len("17299841738")),
				AuthCode:        "17299841738",
				TerminalIMEI:    "123456789012345",
				SoftwareVersion: "3.7.15",
				Version:         consts.JT808Protocol2019,
			},
		},
		{
			name: "T0x0100 终端注册 2011版本",
			args: args{
				msg:      "7e010000200123456789010000001f007363640000007777772e3830382e3736353433323101b2e24131323334a17e",
				Handler:  &T0x0100{},
				bodyLens: []int{24},
			},
			fields: &T0x0100{
				ProvinceID:         31,
				CityID:             115,
				ManufacturerID:     "cd",
				TerminalModel:      "www.808.",
				TerminalID:         "7654321",
				PlateColor:         1,
				LicensePlateNumber: "测A1234",
				Version:            consts.JT808Protocol2011,
			},
		},
		{
			name: "T0x0100 终端注册 2013版本",
			args: args{
				msg:      "7e0100002c0123456789010000001f007363640000007777772e3830382e636f6d0000000000000000003736353433323101b2e24131323334cc7e",
				Handler:  &T0x0100{},
				bodyLens: []int{36},
			},
			fields: &T0x0100{
				ProvinceID:         31,
				CityID:             115,
				ManufacturerID:     "cd",
				TerminalModel:      "www.808.com",
				TerminalID:         "7654321",
				PlateColor:         1,
				LicensePlateNumber: "测A1234",
				Version:            consts.JT808Protocol2013,
			},
		},
		{
			name: "T0x0100 终端注册 2019版本",
			args: args{
				msg:      "7e0100405301000000000172998417380000001f007363640000000000000000007777772e3830382e636f6d0000000000000000000000000000000000000037363534333231000000000000000000000000000000000000000000000001b2e241313233343b7e",
				Handler:  &T0x0100{},
				bodyLens: []int{75},
			},
			fields: &T0x0100{
				ProvinceID:         31,
				CityID:             115,
				ManufacturerID:     "cd",
				TerminalModel:      "www.808.com",
				TerminalID:         "7654321",
				PlateColor:         1,
				LicensePlateNumber: "测A1234",
				Version:            consts.JT808Protocol2019,
			},
		},
		{
			name: "T0x0200 终端-位置上报",
			args: args{
				msg:      "7e0200001c0123456789010000000004000000080007203b7d0202633df70138000300632410012359591c7e",
				Handler:  &T0x0200{},
				bodyLens: []int{27},
			},
			fields: &T0x0200{
				T0x0200LocationItem: T0x0200LocationItem{
					AlarmSign:  1024,
					StatusSign: 2048,
					Latitude:   119552894,
					Longitude:  40058359,
					Altitude:   312,
					Speed:      3,
					Direction:  99,
					DateTime:   "2024-10-01 23:59:59",
				},
			},
		},
		{
			name: "T0x0704 终端-位置批量上传",
			args: args{
				msg:      "7e0704003f0123456789010000000200001c000004000000080007203b7d0202633df7013800030063241001235959001c000004000000080007203b7d0202633df7013800030063241001235959b67e",
				Handler:  &T0x0704{},
				bodyLens: []int{30, 60, 68},
			},
			fields: &T0x0704{
				Num:          2,
				LocationType: 0,
				Items: []T0x0704LocationItem{
					{
						Len: 28,
						T0x0200LocationItem: T0x0200LocationItem{
							AlarmSign:  1024,
							StatusSign: 2048,
							Latitude:   119552894,
							Longitude:  40058359,
							Altitude:   312,
							Speed:      3,
							Direction:  99,
							DateTime:   "2024-10-01 23:59:59",
						},
					},
					{
						Len: 28,
						T0x0200LocationItem: T0x0200LocationItem{
							AlarmSign:  1024,
							StatusSign: 2048,
							Latitude:   119552894,
							Longitude:  40058359,
							Altitude:   312,
							Speed:      3,
							Direction:  99,
							DateTime:   "2024-10-01 23:59:59",
						},
					},
				},
			},
		},
		{
			name: "P0x8104 平台-查询终端参数",
			args: args{
				msg:      "7e8104400001000000000144199999990003027e",
				Handler:  &P0x8104{},
				bodyLens: nil,
			},
			fields: &P0x8104{},
		},
		{
			name: "T0x0104 终端-查询参数",
			args: args{
				msg:      "7E010443A20100000000014419999999000500045B00000001040000000A00000002040000003C00000003040000000200000004040000003C00000005040000000200000006040000003C000000070400000002000000100B31333031323334353637300000001105313233343500000012053132333435000000130E3132372E302E302E313A37303030000000140531323334350000001505313233343500000016053132333435000000170531323334350000001A093132372E302E302E310000001B04000004570000001C04000004580000001D093132372E302E302E310000002004000000000000002104000000000000002204000000000000002301300000002401300000002501300000002601300000002704000000000000002804000000000000002904000000000000002C04000003E80000002D04000003E80000002E04000003E80000002F04000003E800000030040000000A0000003102003C000000320416320A1E000000400B3133303132333435363731000000410B3133303132333435363732000000420B3133303132333435363733000000430B3133303132333435363734000000440B3133303132333435363735000000450400000001000000460400000000000000470400000000000000480B3133303132333435363738000000490B313330313233343536373900000050040000000000000051040000000000000052040000000000000053040000000000000054040000000000000055040000003C000000560400000014000000570400003840000000580400000708000000590400001C200000005A040000012C0000005B0200500000005C0200050000005D02000A0000005E02001E00000064040000000100000065040000000100000070040000000100000071040000006F000000720400000070000000730400000071000000740400000072000000751500030190320000002800030190320000002800050100000076130400000101000002020000030300000404000000000077160101000301F43200000028000301F43200000028000500000079032808010000007A04000000230000007B0232320000007C1405000000000000000000000000000000000000000000008004000000240000008102000B000000820200660000008308BEA9415830303031000000840101000000900102000000910101000000920101000000930400000001000000940100000000950400000001000001000400000064000001010213880000010204000000640000010302138800000110080000000000000101F07E",
				Handler:  &T0x0104{},
				bodyLens: []int{2},
			},
			fields: &T0x0104{
				RespondSerialNumber: 4,
				RespondParamCount:   91,
				TerminalParamDetails: TerminalParamDetails{
					T0x001HeartbeatInterval:                     ParamContent[uint32]{ID: 0x001, Len: 4, Value: 10},
					T0x002TCPRespondOverTime:                    ParamContent[uint32]{ID: 0x002, Len: 4, Value: 60},
					T0x003TCPRetransmissionCount:                ParamContent[uint32]{ID: 0x003, Len: 4, Value: 2},
					T0x004UDPRespondOverTime:                    ParamContent[uint32]{ID: 0x004, Len: 4, Value: 60},
					T0x005UDPRetransmissionCount:                ParamContent[uint32]{ID: 0x005, Len: 4, Value: 2},
					T0x006SMSRetransmissionCount:                ParamContent[uint32]{ID: 0x006, Len: 4, Value: 60},
					T0x007SMSRetransmissionCount:                ParamContent[uint32]{ID: 0x007, Len: 4, Value: 2},
					T0x010APN:                                   ParamContent[string]{ID: 0x010, Len: 11, Value: "13012345670"},
					T0x011WIFIUsername:                          ParamContent[string]{ID: 0x011, Len: 5, Value: "12345"},
					T0x012WIFIPassword:                          ParamContent[string]{ID: 0x012, Len: 5, Value: "12345"},
					T0x013Address:                               ParamContent[string]{ID: 0x013, Len: 14, Value: "127.0.0.1:7000"},
					T0x014BackupServerAPN:                       ParamContent[string]{ID: 0x014, Len: 5, Value: "12345"},
					T0x015BackupServerWIFIUsername:              ParamContent[string]{ID: 0x015, Len: 5, Value: "12345"},
					T0x016BackupServerWIFIPassword:              ParamContent[string]{ID: 0x016, Len: 5, Value: "12345"},
					T0x017BackupServerAddress:                   ParamContent[string]{ID: 0x017, Len: 5, Value: "12345"},
					T0x018TCPPort:                               ParamContent[uint32]{}, // 2019版本不存在
					T0x019UDPPort:                               ParamContent[uint32]{}, // 2019版本不存在
					T0x01AICCardAddress:                         ParamContent[string]{ID: 0x01a, Len: 9, Value: "127.0.0.1"},
					T0x01BICCardTCPPort:                         ParamContent[uint32]{ID: 0x01b, Len: 4, Value: 1111},
					T0x01CICCardUDPPort:                         ParamContent[uint32]{ID: 0x01c, Len: 4, Value: 1112},
					T0x01DICCardAddress:                         ParamContent[string]{ID: 0x01d, Len: 9, Value: "127.0.0.1"},
					T0x020PositionReportingStrategy:             ParamContent[uint32]{ID: 0x020, Len: 4, Value: 0},
					T0x021PositionReportingPlan:                 ParamContent[uint32]{},
					T0x022DriverReportingInterval:               ParamContent[uint32]{ID: 0x022, Len: 4, Value: 0},
					T0x023FromServerAPN:                         ParamContent[string]{ID: 0x023, Len: 1, Value: "0"},
					T0x024FromServerAPNWIFIUsername:             ParamContent[string]{ID: 0x024, Len: 1, Value: "0"},
					T0x025FromServerAPNWIFIPassword:             ParamContent[string]{ID: 0x025, Len: 1, Value: "0"},
					T0x026FromServerAPNWIFIAddress:              ParamContent[string]{ID: 0x026, Len: 1, Value: "0"},
					T0x027ReportingTimeInterval:                 ParamContent[uint32]{ID: 0x027, Len: 4, Value: 0},
					T0x028EmergencyReportingTimeInterval:        ParamContent[uint32]{ID: 0x028, Len: 4, Value: 0},
					T0x029DefaultReportingTimeInterval:          ParamContent[uint32]{ID: 0x029, Len: 4, Value: 0},
					T0x02CDefaultDistanceReportingTimeInterval:  ParamContent[uint32]{ID: 0x02C, Len: 4, Value: 1000},
					T0x02DDrivingReportingDistanceInterval:      ParamContent[uint32]{ID: 0x02D, Len: 4, Value: 1000},
					T0x02ESleepReportingDistanceInterval:        ParamContent[uint32]{ID: 0x02E, Len: 4, Value: 1000},
					T0x02FAlarmReportingDistanceInterval:        ParamContent[uint32]{ID: 0x02F, Len: 4, Value: 1000},
					T0x030InflectionPointSupplementaryPassAngle: ParamContent[uint32]{ID: 0x030, Len: 4, Value: 10},
					T0x031GeofenceRadius:                        ParamContent[uint16]{ID: 0x031, Len: 2, Value: 60},
					T0x032IllegalDrivingTime:                    ParamContent[[4]byte]{ID: 0x032, Len: 4, Value: [4]byte{22, 50, 10, 30}},
					T0x040MonitoringPlatformPhone:               ParamContent[string]{ID: 0x040, Len: 11, Value: "13012345671"},
					T0x041ResetPhone:                            ParamContent[string]{ID: 0x041, Len: 11, Value: "13012345672"},
					T0x042RestoreFactoryPhone:                   ParamContent[string]{ID: 0x042, Len: 11, Value: "13012345673"},
					T0x043SMSPhone:                              ParamContent[string]{ID: 0x043, Len: 11, Value: "13012345674"},
					T0x044SMSTxtPhone:                           ParamContent[string]{ID: 0x044, Len: 11, Value: "13012345675"},
					T0x045TerminalTelephoneStrategy:             ParamContent[uint32]{ID: 0x045, Len: 4, Value: 1},
					T0x046MaximumCallTime:                       ParamContent[uint32]{ID: 0x046, Len: 4, Value: 0},
					T0x047MonthMaximumCallTime:                  ParamContent[uint32]{ID: 0x047, Len: 4, Value: 0},
					T0x048MonitorPhone:                          ParamContent[string]{ID: 0x048, Len: 11, Value: "13012345678"},
					T0x049MonitorPrivilegedSMS:                  ParamContent[string]{ID: 0x049, Len: 11, Value: "13012345679"},
					T0x050AlarmBlockingWords:                    ParamContent[uint32]{ID: 0x050, Len: 4, Value: 0},
					T0x051AlarmSendTextSMSSwitch:                ParamContent[uint32]{ID: 0x051, Len: 4, Value: 0},
					T0x052AlarmShootingSwitch:                   ParamContent[uint32]{ID: 0x052, Len: 4, Value: 0},
					T0x053AlarmShootingStorageSign:              ParamContent[uint32]{ID: 0x053, Len: 4, Value: 0},
					T0x054KeySign:                               ParamContent[uint32]{ID: 0x054, Len: 4, Value: 0},
					T0x055MaxSpeed:                              ParamContent[uint32]{ID: 0x055, Len: 4, Value: 60},
					T0x056DurationOverSpeed:                     ParamContent[uint32]{ID: 0x056, Len: 4, Value: 20},
					T0x057ContinuousDrivingTimeLimit:            ParamContent[uint32]{ID: 0x057, Len: 4, Value: 14400},
					T0x058CumulativeDayDrivingTime:              ParamContent[uint32]{ID: 0x058, Len: 4, Value: 1800},
					T0x059MinimumRestTime:                       ParamContent[uint32]{ID: 0x059, Len: 4, Value: 7200},
					T0x05AMaximumParkingTime:                    ParamContent[uint32]{ID: 0x05a, Len: 4, Value: 300},
					T0x05BSpeedWarningDifference:                ParamContent[uint16]{ID: 0x05b, Len: 2, Value: 80},
					T0x05CFatigueDrivingWarningInterpolation:    ParamContent[uint16]{ID: 0x05c, Len: 2, Value: 5},
					T0x05DCollisionAlarmParam:                   ParamContent[uint16]{ID: 0x05d, Len: 2, Value: 10},
					T0x05ERolloverAlarmParam:                    ParamContent[uint16]{ID: 0x05e, Len: 2, Value: 30},
					T0x064TimedPhotographyParam:                 ParamContent[uint32]{ID: 0x064, Len: 4, Value: 1},
					T0x065FixedDistanceShootingParam:            ParamContent[uint32]{ID: 0x065, Len: 4, Value: 1},
					T0x070ImageVideoQuality:                     ParamContent[uint32]{ID: 0x070, Len: 4, Value: 1},
					T0x071Brightness:                            ParamContent[uint32]{ID: 0x071, Len: 4, Value: 111},
					T0x072Contrast:                              ParamContent[uint32]{ID: 0x072, Len: 4, Value: 112},
					T0x073Saturation:                            ParamContent[uint32]{ID: 0x073, Len: 4, Value: 113},
					T0x074Chrominance:                           ParamContent[uint32]{ID: 0x074, Len: 4, Value: 114},
					T0x080VehicleOdometerReadings:               ParamContent[uint32]{ID: 0x080, Len: 4, Value: 36},
					T0x081VehicleProvinceID:                     ParamContent[uint16]{ID: 0x081, Len: 2, Value: 11},
					T0x082VehicleCityID:                         ParamContent[uint16]{ID: 0x082, Len: 2, Value: 102},
					T0x083MotorVehicleLicensePlate:              ParamContent[string]{ID: 0x083, Len: 8, Value: "京AX0001"},
					T0x084licensePlateColor:                     ParamContent[byte]{ID: 0x084, Len: 1, Value: 1},
					T0x090GNSSPositionMode:                      ParamContent[byte]{ID: 0x090, Len: 1, Value: 2},
					T0x091GNSSBaudRate:                          ParamContent[byte]{ID: 0x091, Len: 1, Value: 1},
					T0x092GNSSModePositionOutputFrequency:       ParamContent[byte]{ID: 0x092, Len: 1, Value: 1},
					T0x093GNSSModePositionAcquisitionFrequency:  ParamContent[uint32]{ID: 0x093, Len: 4, Value: 1},
					T0x094GNSSModePositionUploadMethod:          ParamContent[byte]{ID: 0x094, Len: 1, Value: 0},
					T0x095GNSSModeSetPositionUpload:             ParamContent[uint32]{ID: 0x095, Len: 4, Value: 1},
					T0x100CANCollectionTimeInterval:             ParamContent[uint32]{ID: 0x100, Len: 4, Value: 100},
					T0x101CAN1UploadTimeInterval:                ParamContent[uint16]{ID: 0x101, Len: 2, Value: 5000},
					T0x102CAN2CollectionTimeInterval:            ParamContent[uint32]{ID: 0x102, Len: 4, Value: 100},
					T0x103CAN2UploadTimeInterval:                ParamContent[uint16]{ID: 0x103, Len: 2, Value: 5000},
					T0x110CANIDSetIndividualAcquisition:         ParamContent[[8]byte]{ID: 0x110, Len: 8, Value: [8]byte{0, 0, 0, 0, 0, 0, 1, 1}},
					ParamParseBeforeFunc:                        nil,
					OtherContent: map[uint32]ParamContent[[]byte]{
						33:  {ID: 0x021, Len: 4, Value: []byte{0, 0, 0, 0}},
						117: {ID: 0x075, Len: 21, Value: []byte{0, 3, 1, 144, 50, 0, 0, 0, 40, 0, 3, 1, 144, 50, 0, 0, 0, 40, 0, 5, 1}},
						118: {ID: 0x076, Len: 19, Value: []byte{4, 0, 0, 1, 1, 0, 0, 2, 2, 0, 0, 3, 3, 0, 0, 4, 4, 0, 0}},
						119: {ID: 0x077, Len: 22, Value: []byte{1, 1, 0, 3, 1, 244, 50, 0, 0, 0, 40, 0, 3, 1, 244, 50, 0, 0, 0, 40, 0, 5}},
						121: {ID: 0x079, Len: 3, Value: []byte{40, 8, 1}},
						122: {ID: 0x07a, Len: 4, Value: []byte{0, 0, 0, 35}},
						123: {ID: 0x07b, Len: 2, Value: []byte{50, 50}},
						124: {ID: 0x07c, Len: 20, Value: []byte{5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
					},
				},
			},
		},
		{
			name: "P0x9003 平台-查询终端音视频属性",
			args: args{
				msg:      "7e9003400001000000000144199999990003147e",
				Handler:  &P0x9003{},
				bodyLens: nil,
			},
			fields: &P0x9003{},
		},
		{
			name: "T0x1003 终端-上传音视频属性",
			args: args{
				msg:      "7e1003000a12345678901200017f040200944901200808177e",
				Handler:  &T0x1003{},
				bodyLens: []int{1},
			},
			fields: &T0x1003{
				EnterAudioEncoding:       127,
				EnterAudioChannelsNumber: 4,
				EnterAudioSampleRate:     2,
				EnterAudioSampleDigits:   0,
				AudioFrameLength:         37961,
				HasSupportedAudioOutput:  1,
				VideoEncoding:            32,
				TerminalSupportedMaxNumberOfAudioPhysicalChannels: 8,
				TerminalSupportedMaxNumberOfVideoPhysicalChannels: 8,
			},
		},
		{
			name: "P0x9101 平台-实时音视频传输请求",
			args: args{
				msg:      "7e9101001712345678901200010f3132332e3132332e3132332e313233030440c60c0100a17e",
				Handler:  &P0x9101{},
				bodyLens: []int{0, 2},
			},
			fields: &P0x9101{
				ServerIPLen:  15,
				ServerIPAddr: "123.123.123.123",
				TcpPort:      772,
				UdpPort:      16582,
				ChannelNo:    12,
				DataType:     1,
				StreamType:   0,
			},
		},
		{
			name: "T0x1005 终端-上传乘客流量",
			args: args{
				msg:      "7E100500101234567890120001241001000000241002001001000200138D7E",
				Handler:  &T0x1005{},
				bodyLens: []int{15},
			},
			fields: &T0x1005{
				StartTime:    "2024-10-01 00:00:00",
				EndTime:      "2024-10-02 00:10:01",
				BoardNumber:  2,
				AlightNumber: 19,
			},
		},
		{
			name: "P0x9102 平台-音视频实时传输控制",
			args: args{
				msg:      "7e910240040112345678901234567890ffff08010203de7e",
				Handler:  &P0x9102{},
				bodyLens: []int{3},
			},
			fields: &P0x9102{
				ChannelNo:           8,
				ControlCmd:          1,
				CloseAudioVideoData: 2,
				StreamType:          3,
			},
		},
		{
			name: "P0x9201 平台-下发远程录像回放请求",
			args: args{
				msg:      "7e9201002412345678901200010d31322e31322e3132332e313233a7b93c6c320200000000200707192359200707192359617e",
				Handler:  &P0x9201{},
				bodyLens: []int{0, 3},
			},
			fields: &P0x9201{
				BaseHandle:   BaseHandle{},
				ServerIPLen:  13,
				ServerIPAddr: "12.12.123.123",
				TcpPort:      42937,
				UdpPort:      15468,
				ChannelNo:    50,
				MediaType:    2,
				StreamType:   0,
				MemoryType:   0,
				PlaybackWay:  0,
				PlaySpeed:    0,
				StartTime:    "2020-07-07 19:23:59",
				EndTime:      "2020-07-07 19:23:59",
			},
		},
		{
			name: "P0x9205 平台-查询资源列表",
			args: args{
				msg:      "7e920500181234567890120001e720070719235920070719235900000000000000009b6e00167e",
				Handler:  &P0x9205{},
				bodyLens: []int{1},
			},
			fields: &P0x9205{
				BaseHandle:  BaseHandle{},
				ChannelNo:   231,
				StartTime:   "2020-07-07 19:23:59",
				EndTime:     "2020-07-07 19:23:59",
				AlarmFlag:   0,
				MediaType:   155,
				StreamType:  110,
				StorageType: 0,
			},
		},
		{
			name: "T0x1205 终端-上传音视频资源列表",
			args: args{
				msg:      "7e1205002212345678901200000000000000010124110200000024110200010200000000000004000101010000000bb27e",
				Handler:  &T0x1205{},
				bodyLens: []int{1, 7},
			},
			fields: &T0x1205{
				SerialNumber:            0,
				AudioVideoResourceTotal: 1,
				AudioVideoResourceList: []T0x1205AudioVideoResource{
					{
						ChannelNo:              1,
						StartTime:              "2024-11-02 00:00:00",
						EndTime:                "2024-11-02 00:01:02",
						AlarmFlag:              1024,
						AudioVideoResourceType: 1,
						StreamType:             1,
						MemoryType:             1,
						FileSizeByte:           11,
					},
				},
			},
		},
		{
			name: "P0x9206 平台-文件上传指令",
			args: args{
				msg:      "7e9206004512345678901200010b3139322e3136382e312e312b2d08757365726e616d650870617373776f72640b2f616c61726d5f66696c6501200726000000200726232359000000000000000000010101227e",
				Handler:  &P0x9206{},
				bodyLens: []int{0, 14, 23, 32, 44},
			},
			fields: &P0x9206{
				FTPAddrLen:           11,
				FTPAddr:              "192.168.1.1",
				Port:                 11053,
				UsernameLen:          8,
				Username:             "username",
				PasswordLen:          8,
				Password:             "password",
				FileUploadPathLen:    11,
				FileUploadPath:       "/alarm_file",
				ChannelNo:            1,
				StartTime:            "2020-07-26 00:00:00",
				EndTime:              "2020-07-26 23:23:59",
				AlarmFlag:            0,
				MediaType:            0,
				StreamType:           1,
				MemoryPosition:       1,
				TaskExecuteCondition: 1,
			},
		},
		{
			name: "T0x1206 终端-文件上传完成通知",
			args: args{
				msg:      "7e120640030112345678901234567890ffff1b8a01c67e",
				Handler:  &T0x1206{},
				bodyLens: []int{1},
			},
			fields: &T0x1206{
				RespondSerialNumber: 7050,
				Result:              1,
			},
		},
		{
			name: "P0x9207 平台-文件上传控制",
			args: args{
				msg:      "7e92070003123456789012000169fd028b7e",
				Handler:  &P0x9207{},
				bodyLens: []int{1},
			},
			fields: &P0x9207{
				RespondSerialNumber: 27133,
				UploadControl:       2,
			},
		},
		{
			name: "P0x8003 平台-补发分包请求",
			args: args{
				msg:      "7e800300150123456789017fff1099090001000200030004000500060007000800091f7e",
				Handler:  &P0x8003{},
				bodyLens: []int{2, 4},
			},
			fields: &P0x8003{
				OriginalSerialNumber: 4249,
				AgainPackageCount:    9,
				AgainPackageList:     []uint16{1, 2, 3, 4, 5, 6, 7, 8, 9},
			},
		},
		{
			name: "P0x9105 平台-音视频实时传输状态通知",
			args: args{
				msg:      "7e91050002123456789012000102031c7e",
				Handler:  &P0x9105{},
				bodyLens: []int{1},
			},
			fields: &P0x9105{
				ChannelNo:       2,
				PackageLossRate: 3,
			},
		},
		{
			name: "P0x9202 平台-下发远程录像回放控制",
			args: args{
				msg:      "7e920200091234567890120001110103200707192359427e",
				Handler:  &P0x9202{},
				bodyLens: []int{1},
			},
			fields: &P0x9202{
				ChannelNo:   17,
				PlayControl: 1,
				PlaySpeed:   3,
				DateTime:    "2020-07-07 19:23:59",
			},
		},
		{
			name: "P0x8103 平台-设置终端参数",
			args: args{
				msg:      "7e810302101234567890120000280000000104626a6a65000000020442434b6d00000003044b456863000000040445456357000000050434516a39000000060441464d5f00000007043173666c00000010104a4468326c32394e6a75416e726c58750000001110666b756d6349376c7a4d5f76776f7a43000000121034356e3077523932445570555a7a7258000000131071444b6e4636666c6974694377554d4b0000001410476a7071376f6d55553834686e646561000000151070676c785f375251677971467648725700000016105559317a4e574b706754656a715f79300000001710725a554238704c4476516363743857680000001a10656d48775f6d317263547550374756370000001b044c5832510000001c04754d457a0000001d10464d6f524f627a30594573534147686400000020044471634c000000220443335f310000002310556f5234774d494438506669456267560000002410484954455f76684273496742376f5057000000251039364753596448434d3733676e53536800000026107759765248434f6a346135573351465a0000002704366375620000002804573854610000002904566d4e6b0000002c04324c5f540000002d0441764a560000002e04537443410000002f045958376f00000030047078566800000031025a6900000032040930213000000092010c0000011008000102030405060700000018046f78335100000019047a4a6158000000210434303749ac7e",
				Handler:  &P0x8103{},
				bodyLens: []int{0},
			},
			fields: &P0x8103{
				ParamTotal: 40,
				TerminalParamDetails: TerminalParamDetails{
					T0x001HeartbeatInterval: ParamContent[uint32]{
						ID:    0x01,
						Len:   4,
						Value: 1651141221,
					},
					T0x002TCPRespondOverTime: ParamContent[uint32]{
						ID:    0x02,
						Len:   4,
						Value: 1111706477,
					},
					T0x003TCPRetransmissionCount: ParamContent[uint32]{
						ID:    0x03,
						Len:   4,
						Value: 1262839907,
					},
					T0x004UDPRespondOverTime: ParamContent[uint32]{
						ID:    0x04,
						Len:   4,
						Value: 1162175319,
					},
					T0x005UDPRetransmissionCount: ParamContent[uint32]{
						ID:    0x05,
						Len:   4,
						Value: 877750841,
					},
					T0x006SMSRetransmissionCount: ParamContent[uint32]{
						ID:    0x06,
						Len:   4,
						Value: 1095126367,
					},
					T0x007SMSRetransmissionCount: ParamContent[uint32]{
						ID:    0x07,
						Len:   4,
						Value: 829646444,
					},
					T0x010APN: ParamContent[string]{
						ID:    0x10,
						Len:   byte(len("JDh2l29NjuAnrlXu")),
						Value: "JDh2l29NjuAnrlXu",
					},
					T0x011WIFIUsername: ParamContent[string]{
						ID:    0x11,
						Len:   byte(len("fkumcI7lzM_vwozC")),
						Value: "fkumcI7lzM_vwozC",
					},
					T0x012WIFIPassword: ParamContent[string]{
						ID:    0x12,
						Len:   byte(len("45n0wR92DUpUZzrX")),
						Value: "45n0wR92DUpUZzrX",
					},
					T0x013Address: ParamContent[string]{
						ID:    0x13,
						Len:   byte(len("qDKnF6flitiCwUMK")),
						Value: "qDKnF6flitiCwUMK",
					},
					T0x014BackupServerAPN: ParamContent[string]{
						ID:    0x14,
						Len:   byte(len("Gjpq7omUU84hndea")),
						Value: "Gjpq7omUU84hndea",
					},
					T0x015BackupServerWIFIUsername: ParamContent[string]{
						ID:    0x15,
						Len:   byte(len("pglx_7RQgyqFvHrW")),
						Value: "pglx_7RQgyqFvHrW",
					},
					T0x016BackupServerWIFIPassword: ParamContent[string]{
						ID:    0x16,
						Len:   byte(len("UY1zNWKpgTejq_y0")),
						Value: "UY1zNWKpgTejq_y0",
					},
					T0x017BackupServerAddress: ParamContent[string]{
						ID:    0x17,
						Len:   byte(len("rZUB8pLDvQcct8Wh")),
						Value: "rZUB8pLDvQcct8Wh",
					},
					T0x01AICCardAddress: ParamContent[string]{
						ID:    0x1a,
						Len:   byte(len("emHw_m1rcTuP7GV7")),
						Value: "emHw_m1rcTuP7GV7",
					},
					T0x01BICCardTCPPort: ParamContent[uint32]{
						ID:    0x1b,
						Len:   4,
						Value: 1280848465,
					},
					T0x01CICCardUDPPort: ParamContent[uint32]{
						ID:    0x1c,
						Len:   4,
						Value: 1967998330,
					},
					T0x01DICCardAddress: ParamContent[string]{
						ID:    0x1d,
						Len:   byte(len("FMoRObz0YEsSAGhd")),
						Value: "FMoRObz0YEsSAGhd",
					},
					T0x020PositionReportingStrategy: ParamContent[uint32]{
						ID:    0x20,
						Len:   4,
						Value: 1148281676,
					},
					T0x022DriverReportingInterval: ParamContent[uint32]{
						ID:    0x22,
						Len:   4,
						Value: 1127440177,
					},
					T0x023FromServerAPN: ParamContent[string]{
						ID:    0x23,
						Len:   byte(len("UoR4wMID8PfiEbgV")),
						Value: "UoR4wMID8PfiEbgV",
					},
					T0x024FromServerAPNWIFIUsername: ParamContent[string]{
						ID:    0x24,
						Len:   byte(len("HITE_vhBsIgB7oPW")),
						Value: "HITE_vhBsIgB7oPW",
					},
					T0x025FromServerAPNWIFIPassword: ParamContent[string]{
						ID:    0x25,
						Len:   byte(len("96GSYdHCM73gnSSh")),
						Value: "96GSYdHCM73gnSSh",
					},
					T0x026FromServerAPNWIFIAddress: ParamContent[string]{
						ID:    0x26,
						Len:   byte(len("wYvRHCOj4a5W3QFZ")),
						Value: "wYvRHCOj4a5W3QFZ",
					},
					T0x027ReportingTimeInterval: ParamContent[uint32]{
						ID:    0x27,
						Len:   4,
						Value: 912487778,
					},
					T0x028EmergencyReportingTimeInterval: ParamContent[uint32]{
						ID:    0x28,
						Len:   4,
						Value: 1463309409,
					},
					T0x029DefaultReportingTimeInterval: ParamContent[uint32]{
						ID:    0x29,
						Len:   4,
						Value: 1450004075,
					},
					T0x02CDefaultDistanceReportingTimeInterval: ParamContent[uint32]{
						ID:    0x2c,
						Len:   4,
						Value: 843865940,
					},
					T0x02DDrivingReportingDistanceInterval: ParamContent[uint32]{
						ID:    0x2d,
						Len:   4,
						Value: 1098271318,
					},
					T0x02ESleepReportingDistanceInterval: ParamContent[uint32]{
						ID:    0x2e,
						Len:   4,
						Value: 1400128321,
					},
					T0x02FAlarmReportingDistanceInterval: ParamContent[uint32]{
						ID:    0x2f,
						Len:   4,
						Value: 1498953583,
					},
					T0x030InflectionPointSupplementaryPassAngle: ParamContent[uint32]{
						ID:    0x30,
						Len:   4,
						Value: 1886934632,
					},
					T0x031GeofenceRadius: ParamContent[uint16]{
						ID:    0x31,
						Len:   2,
						Value: 23145,
					},
					T0x032IllegalDrivingTime: ParamContent[[4]byte]{
						ID:    0x32,
						Len:   4,
						Value: [4]byte{9, 48, 33, 48},
					},
					T0x092GNSSModePositionOutputFrequency: ParamContent[byte]{
						ID:    0x92,
						Len:   1,
						Value: 12,
					},
					T0x110CANIDSetIndividualAcquisition: ParamContent[[8]byte]{
						ID:    0x110,
						Len:   8,
						Value: [8]byte{0, 1, 2, 3, 4, 5, 6, 7},
					},
					OtherContent: map[uint32]ParamContent[[]byte]{
						0x18: {
							ID:    0x18,
							Len:   4,
							Value: []byte{111, 120, 51, 81},
						},
						0x19: {
							ID:    0x19,
							Len:   4,
							Value: []byte{122, 74, 97, 88},
						},
						0x21: {
							ID:    0x21,
							Len:   4,
							Value: []byte{52, 48, 55, 73},
						},
					},
				},
			},
		},
		{
			name: "P0x8801 平台-摄像头立即拍摄命令",
			args: args{
				msg:      "7e8801400c0100000000017299841738ffff0100020003010405ff7f7fff857e",
				Handler:  &P0x8801{},
				bodyLens: []int{1},
			},
			fields: &P0x8801{
				ChannelID:                1,
				ShootCommand:             2,
				PhotoIntervalOrVideoTime: 3,
				SaveFlag:                 1,
				Resolution:               4,
				VideoQuality:             5,
				Intensity:                255,
				Contrast:                 127,
				Saturation:               127,
				Chroma:                   255,
			},
		},
		{
			name: "T0x0805 终端-摄像头立即拍照",
			args: args{
				msg:      "7e080500290123456789017ffff4c0000009000000010000000200000003000000040000000500000006000000070000000800000009107e",
				Handler:  &T0x0805{},
				bodyLens: []int{4, 6, 10},
			},
			fields: &T0x0805{
				RespondSerialNumber: 62656,
				Result:              0,
				MultimediaIDNumber:  9,
				MultimediaIDList:    []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9},
			},
		},
		{
			name: "T0x0800 终端-多媒体事件信息上传",
			args: args{
				msg:      "7e080000080123456789017fff0000007b00000701757e",
				Handler:  &T0x0800{},
				bodyLens: []int{1},
			},
			fields: &T0x0800{
				MultimediaID:           123,
				MultimediaType:         0,
				MultimediaFormatEncode: 0,
				EventItemEncode:        7,
				ChannelID:              1,
			},
		},
		{
			name: "T0x0801 终端-多媒体数据上传",
			args: args{
				msg:      "7e080100290123456789017fff0000007b01020102000004000000080006eeb6ad02633df70138000300632007071923590d7b0d7b7b667e",
				Handler:  &T0x0801{},
				bodyLens: []int{1, 10},
			},
			fields: &T0x0801{
				MultimediaID:           123,
				MultimediaType:         1,
				MultimediaFormatEncode: 2,
				EventItemEncode:        1,
				ChannelID:              2,
				T0x0200LocationItem: T0x0200LocationItem{
					AlarmSign:         1024,
					StatusSign:        2048,
					Latitude:          116307629,
					Longitude:         40058359,
					Altitude:          312,
					Speed:             3,
					Direction:         99,
					DateTime:          "2020-07-07 19:23:59",
					AlarmSignDetails:  AlarmSignDetails{},
					StatusSignDetails: StatusSignDetails{},
				},
				MultimediaPackage: []byte{13, 123, 13, 123, 123},
			},
		},
		{
			name: "P0x8800 平台-多媒体上传应答",
			args: args{
				msg:      "7e880000170123456789017fff0000c15f09000100020003000400050006000700080009017e",
				Handler:  &P0x8800{},
				bodyLens: []int{1, 10},
			},
			fields: &P0x8800{
				MultimediaID:      49503,
				AgainPackageCount: 9,
				AgainPackageList:  []uint16{1, 2, 3, 4, 5, 6, 7, 8, 9},
			},
		},
		{
			name: "P0x9208 平台-报警附件上传指令",
			args: args{
				msg:      "7e9208005212345678901200010d34372e3130342e39372e313639200a200b37363534333231200707192359010101616437323133313537396535346265306230663733376366633732633564623800000000000000000000000000000000427e",
				Handler:  &P0x9208{},
				bodyLens: []int{1, 54},
			},
			fields: &P0x9208{
				ServerIPLen: 13,
				ServerAddr:  "47.104.97.169",
				TcpPort:     8202,
				UdpPort:     8203,
				P9208AlarmSign: P9208AlarmSign{
					TerminalID:   "7654321",
					Time:         "2020-07-07 19:23:59",
					SerialNumber: 1,
					AttachNumber: 1,
					AlarmReserve: []byte{1},
				},
				AlarmID: "ad72131579e54be0b0f737cfc72c5db8",
				Reserve: make([]byte, 16),
			},
		},
		{
			name: "T0x1210 终端-报警附件信息消息 (广东)",
			args: args{
				msg: "7e121000ae00000000100100013132333463642e00000000000000000000000000000000000000000000003132333463642e000000000000000000000000000000000000000000000024120619041201020000323032342d31312d32325f31305f30305f30305f00000000000000000000000000021c323032342d31312d32325f31305f30305f30305f646174612e747874000803e620323032342d31312d32325f31305f30305f30305f72747673393130312e706e670005633f147e",
				Handler: &T0x1210{
					P9208AlarmSign: P9208AlarmSign{
						ActiveSafetyType: consts.ActiveSafetyGD,
					},
				},
				bodyLens: []int{},
			},
			fields: &T0x1210{
				TerminalID: "1234cd.",
				P9208AlarmSign: P9208AlarmSign{
					TerminalID:       "1234cd.",
					Time:             "2024-12-06 19:04:12",
					SerialNumber:     1,
					AttachNumber:     2,
					AlarmReserve:     []byte{0, 0},
					ActiveSafetyType: consts.ActiveSafetyGD,
				},
				AlarmID:     "2024-11-22_10_00_00_",
				InfoType:    0,
				AttachCount: 2,
				T0x1210AlarmItemList: []T0x1210AlarmItem{
					{
						FileNameLen: 28,
						FileName:    "2024-11-22_10_00_00_data.txt",
						FileSize:    525286,
					},
					{
						FileNameLen: 32,
						FileName:    "2024-11-22_10_00_00_rtvs9101.png",
						FileSize:    353087,
					},
				},
			},
		},
		{
			name: "T0x1210 终端-报警附件信息消息 (黑龙江)",
			args: args{
				msg: "7e1210008e00000000100100013132333463642e00000000000000000000000000000000000000000000002412061836590102323032342d31312d32325f31305f30305f30305f00000000000000000000000000021c323032342d31312d32325f31305f30305f30305f646174612e747874000803e620323032342d31312d32325f31305f30305f30305f72747673393130312e706e670005633f617e",
				Handler: &T0x1210{
					P9208AlarmSign: P9208AlarmSign{
						ActiveSafetyType: consts.ActiveSafetyHLJ,
					},
				},
				bodyLens: []int{},
			},
			fields: &T0x1210{
				TerminalID: "",
				P9208AlarmSign: P9208AlarmSign{
					TerminalID:       "1234cd.",
					Time:             "2024-12-06 18:36:59",
					SerialNumber:     1,
					AttachNumber:     2,
					AlarmReserve:     nil,
					ActiveSafetyType: consts.ActiveSafetyHLJ,
				},
				AlarmID:     "2024-11-22_10_00_00_",
				InfoType:    0,
				AttachCount: 2,
				T0x1210AlarmItemList: []T0x1210AlarmItem{
					{
						FileNameLen: 28,
						FileName:    "2024-11-22_10_00_00_data.txt",
						FileSize:    525286,
					},
					{
						FileNameLen: 32,
						FileName:    "2024-11-22_10_00_00_rtvs9101.png",
						FileSize:    353087,
					},
				},
			},
		},
		{
			name: "T0x1210 终端-报警附件信息消息",
			args: args{
				msg:      "7e1210005800000000155800003132336364000031323363640000241111000000010201616161000000000000000000000000000000000000000000000000000000000000020b3132335f6161612e6a7067000004d20a63645f6161612e6d70340001e240457e",
				Handler:  &T0x1210{},
				bodyLens: []int{56, 58, 72},
			},
			fields: &T0x1210{
				TerminalID: "123cd",
				P9208AlarmSign: P9208AlarmSign{
					TerminalID:   "123cd",
					Time:         "2024-11-11 00:00:00",
					SerialNumber: 1,
					AttachNumber: 2,
					AlarmReserve: []byte{1},
				},
				AlarmID:     "aaa",
				InfoType:    0,
				AttachCount: 2,
				T0x1210AlarmItemList: []T0x1210AlarmItem{
					{
						FileNameLen: byte(len("123_aaa.jpg")),
						FileName:    "123_aaa.jpg",
						FileSize:    1234,
					},
					{
						FileNameLen: byte(len("cd_aaa.mp4")),
						FileName:    "cd_aaa.mp4",
						FileSize:    123456,
					},
				},
			},
		},
		{
			name: "T0x1211 终端-文件信息上传",
			args: args{
				msg:      "7e121140130112345678901234567890ffff0d7777772e6a74743830382e636e0100000400797e",
				Handler:  &T0x1211{},
				bodyLens: []int{1, 7},
			},
			fields: &T0x1211{
				FileNameLen: 13,
				FileName:    "www.jtt808.cn",
				FileType:    1,
				FileSize:    1024,
			},
		},
		{
			name: "T0x1212 终端-文件上传完成消息",
			args: args{
				msg:      "7e1212001312345678901200010d7777772e6a74743830382e636e0100000400b07e",
				Handler:  &T0x1212{},
				bodyLens: []int{1, 7},
			},
			fields: &T0x1212{
				T0x1211: T0x1211{
					FileNameLen: 13,
					FileName:    "www.jtt808.cn",
					FileType:    1,
					FileSize:    1024,
				},
			},
		},
		{
			name: "P0x9212 平台-文件上传完成消息应答",
			args: args{
				msg:      "7e921240190112345678901234567890ffff0d7777772e6a74743830382e636e0001010000000000000400f17e",
				Handler:  &P0x9212{},
				bodyLens: []int{1, 5, 20},
			},
			fields: &P0x9212{
				FileNameLen:            13,
				FileName:               "www.jtt808.cn",
				FileType:               0,
				UploadResult:           1,
				RetransmitPacketNumber: 1,
				P0x9212RetransmitPacketList: []P0x9212RetransmitPacket{
					{
						DataOffset: 0,
						DataLength: 1024,
					},
				},
			},
		},
		{
			name: "P0x8300 平台-文件上传完成消息应答",
			args: args{
				msg:      "7e830000150012562569271108ffb2e2cad431323340343536236162632bbde1caf8507e",
				Handler:  &P0x8300{},
				bodyLens: []int{1},
			},
			fields: &P0x8300{
				Flag: 0, // 使用0的话 会根据文本详情生成标志 相当于写了255
				Text: "测试123@456#abc+结束",
				P0x8300TextFlagDetails: P0x8300TextFlagDetails{
					Urgent:            true,
					Bit1Reserve:       true,
					Display:           true,
					TTS:               true,
					AdvertisingScreen: true,
					InfoCategory:      1,
					Bit6Reserve:       true,
					Bit7Reserve:       true,
				},
			},
		},
		{
			name: "P0x8300 平台-文件上传完成消息应答 过长的文本",
			args: args{
				msg:      "7e830003e900125625692700bc0061626331313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313131313132323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232323232ab7e",
				Handler:  &P0x8300{},
				bodyLens: []int{1},
			},
			fields: &P0x8300{
				Flag: 0,
				Text: "abc1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111112222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222" + "aaaaa",
			},
		},
		{
			name: "P0x8302 平台-提问下发",
			args: args{
				msg:      "7e8302000e001256256927001cff03313233010002414102000142327e",
				Handler:  &P0x8302{},
				bodyLens: []int{1, 5, 9},
			},
			fields: &P0x8302{
				Flag:               0, // 使用0的话 会根据详情生成标志 相当于写了255
				QuestionContentLen: 3,
				QuestionContent:    "123",
				AnswerList: []P0x8302Answer{
					{
						AnswerID:         1,
						AnswerContentLen: 2,
						AnswerContent:    "AA",
					},
					{
						AnswerID:         2,
						AnswerContentLen: 1,
						AnswerContent:    "B",
					},
				},
				P0x8302TextFlagDetails: P0x8302TextFlagDetails{
					Urgent:            true,
					Bit1Reserve:       true,
					Bit2Reserve:       true,
					TTS:               true,
					AdvertisingScreen: true,
					Bit5Reserve:       true,
					Bit6Reserve:       true,
					Bit7Reserve:       true,
				},
			},
		},
		{
			name: "T0x0302 平台-回复下发",
			args: args{
				msg:      "7e030200030123456789017fffef447fde7e",
				Handler:  &T0x0302{},
				bodyLens: []int{1},
			},
			fields: &T0x0302{
				SerialNumber: 61252,
				AnswerID:     127,
			},
		},
		{
			name: "P0x8201 平台-查询位置",
			args: args{
				msg:     "7e82010000001256256927000fa37e",
				Handler: &P0x8201{},
			},
			fields: &P0x8201{},
		},
		{
			name: "T0x0201 终端-查询位置",
			args: args{
				msg:      "7e0201001e0123456789017fff686200002a5a000074280000a3e50000db4fbc732711012c2005121212595b7e",
				Handler:  &T0x0201{},
				bodyLens: []int{1},
			},
			fields: &T0x0201{
				RespondSerialNumber: 26722,
				T0x0200LocationItem: T0x0200LocationItem{
					AlarmSign:         10842,
					StatusSign:        29736,
					Latitude:          41957,
					Longitude:         56143,
					Altitude:          48243,
					Speed:             10001,
					Direction:         300,
					DateTime:          "2020-05-12 12:12:59",
					AlarmSignDetails:  AlarmSignDetails{},
					StatusSignDetails: StatusSignDetails{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := hex.DecodeString(tt.args.msg)
			jtMsg := jt808.NewJTMessage()
			if err := jtMsg.Decode(data); err != nil {
				t.Errorf("Decode() error = %v", err)
				return
			}
			if err := tt.args.Parse(jtMsg); err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			//fmt.Println(tt.args.Handler.String())
			if tt.args.Handler.String() != tt.fields.String() {
				t.Errorf("Parse() got: \n%v\nwant:\n%v", tt.args, tt.fields)
				return
			}
			if gotBody := tt.args.Handler.Encode(); gotBody == nil && len(jtMsg.Body) > 0 {
				t.Log("暂未实现的Encode()")
			} else if fmt.Sprintf("%x", jtMsg.Body) != fmt.Sprintf("%x", gotBody) {
				t.Errorf("Encode() got: \n%x\nwant:\n%x", jtMsg.Body, gotBody)
				return
			}

			body := jtMsg.Body
			for _, bodyLen := range tt.args.bodyLens {
				jtMsg.Body = body[:bodyLen]
				if err := tt.args.Parse(jtMsg); err != nil {
					if !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
						t.Errorf("Parse() error = %v", err)
						return
					}
				}
			}
		})
	}
}

// 为了覆盖率100%增加的测试 ------------------------------------
func TestT0x0704Parse(t *testing.T) {
	msg := "7e070400610123456789017fff000301001c000004000000080006eeb6ad02633df7013800030063200707192359001c000004000000080006eeb6ad02633df70138000300632007071923590020000004000000080006eeb6ad02633df701380003006320070719235902010012177e"
	data, _ := hex.DecodeString(msg)
	jtMsg := jt808.NewJTMessage()
	_ = jtMsg.Decode(data)
	{
		var handler T0x0704
		if err := handler.Parse(jtMsg); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
			t.Errorf("T0x0704 Parse() err[%v]", err)
			return
		}
	}
	{
		handler := &T0x0704{}
		// 强制错误情况
		jtMsg.Body = jtMsg.Body[:63]
		jtMsg.Body[4] = 0x00
		if err := handler.Parse(jtMsg); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
			t.Errorf("T0x0704 Parse() err[%v]", err)
			return
		}
	}
}

func TestT0x0200LocationItemString(t *testing.T) {
	var t0x0200Item T0x0200LocationItem
	t0x0200Item.AlarmSignDetails.parse(math.MaxUint32)
	alarmSignData, _ := os.ReadFile("./testdata/0x0200_alarm_sign.txt")
	if string(alarmSignData) != t0x0200Item.AlarmSignDetails.String() {
		t.Errorf("want[%s] actual[%s]", string(alarmSignData), t0x0200Item.AlarmSignDetails.String())
		return
	}

	infos := map[uint32]string{
		1<<23 - 1:             "./testdata/0x0200_status_sign_03.txt",
		1<<23 - 1 - 256 - 512: "./testdata/0x0200_status_sign_00.txt",
		1<<23 - 1 - 256:       "./testdata/0x0200_status_sign_01.txt",
		1<<23 - 1 - 512:       "./testdata/0x0200_status_sign_02.txt",
	}
	for statusSign, signPath := range infos {
		var tmp T0x0200LocationItem
		tmp.StatusSignDetails.parse(statusSign)
		statusSignData, _ := os.ReadFile(signPath)
		if string(statusSignData) != tmp.StatusSignDetails.String() {
			t.Errorf("path[%s]\n%s", signPath, tmp.StatusSignDetails.String())
			return
		}
	}
}

func TestP9208AlarmSign(t *testing.T) {
	type want struct {
		terminalIDLen int
		alarmSignLen  int
	}
	tests := []struct {
		name string
		args consts.ActiveSafetyType
		want want
	}{
		{
			name: "江苏",
			args: consts.ActiveSafetyJS,
			want: want{
				terminalIDLen: 7,
				alarmSignLen:  16,
			},
		},
		{
			name: "黑龙江",
			args: consts.ActiveSafetyHLJ,
			want: want{
				terminalIDLen: 30,
				alarmSignLen:  38,
			},
		},
		{
			name: "广东",
			args: consts.ActiveSafetyGD,
			want: want{
				terminalIDLen: 30,
				alarmSignLen:  40,
			},
		},
		{
			name: "湖南",
			args: consts.ActiveSafetyHN,
			want: want{
				terminalIDLen: 7,
				alarmSignLen:  32,
			},
		},
		{
			name: "四川",
			args: consts.ActiveSafetySC,
			want: want{
				terminalIDLen: 30,
				alarmSignLen:  39,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &P9208AlarmSign{
				ActiveSafetyType: tt.args,
			}
			if got := p.getTerminalIDLen(); got != tt.want.terminalIDLen {
				t.Errorf("getTerminalIDLen got=[%d] want=[%d]", got, tt.want.terminalIDLen)
			}
			if got := p.getAlarmSignLen(); got != tt.want.alarmSignLen {
				t.Errorf("getAlarmSignLen got=[%d] want=[%d]", got, tt.want.alarmSignLen)
			}
		})
	}
}

func TestP9208AlarmSignEncode(t *testing.T) {
	p1 := P9208AlarmSign{
		TerminalID:       "1234cd.",
		Time:             "2024-12-06 19:04:12",
		SerialNumber:     1,
		AttachNumber:     2,
		AlarmReserve:     []byte{0, 0},
		ActiveSafetyType: consts.ActiveSafetyGD,
	}
	p2 := P9208AlarmSign{
		TerminalID:       "1234cd.",
		Time:             "2024-12-06 19:04:12",
		SerialNumber:     1,
		AttachNumber:     2,
		ActiveSafetyType: consts.ActiveSafetyGD,
	}
	p1.encode()
	p2.encode()
	if p1.String() != p2.String() {
		t.Errorf("want[%s] actual[%s]", p1.String(), p2.String())
		return
	}
}
