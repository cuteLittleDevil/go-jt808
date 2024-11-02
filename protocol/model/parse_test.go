package model

import (
	"encoding/hex"
	"errors"
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
	}
	type args struct {
		msg string
		Handler
		bodyLens []int // 用于覆盖率100测试 强制替换了解析正确的body
	}
	tests := []struct {
		name   string
		fields Handler
		args   args
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
				t.Errorf("Parse() want: \n%v\nactual:\n%v", tt.args, tt.fields)
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
