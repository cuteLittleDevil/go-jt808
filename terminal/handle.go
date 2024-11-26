package terminal

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type Handler interface {
	String() string
	Protocol() consts.JT808CommandType
	Encode() []byte
	Parse(jtMsg *jt808.JTMessage) error
	ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error)
	ReplyProtocol() consts.JT808CommandType
}

type meHandle interface {
	String() string
	Protocol() consts.JT808CommandType
	Encode() []byte
	Parse(jtMsg *jt808.JTMessage) error
	ReplyProtocol() consts.JT808CommandType
}

func defaultProtocolHandles(protocolVersion consts.ProtocolVersionType) map[consts.JT808CommandType]Handler {
	item := model.T0x0200LocationItem{
		AlarmSign:  1024,
		StatusSign: 2048,
		Latitude:   116307629,
		Longitude:  40058359,
		Altitude:   312,
		Speed:      3,
		Direction:  99,
		DateTime:   "2024-10-01 23:59:59",
	}
	return map[consts.JT808CommandType]Handler{
		consts.T0001GeneralRespond:                    newT0x0001(),
		consts.T0002HeartBeat:                         &model.T0x0002{},
		consts.T0100Register:                          newT0x0100(protocolVersion),
		consts.T0102RegisterAuth:                      newT0x0102(protocolVersion),
		consts.T0200LocationReport:                    newT0x0200(item),
		consts.T0704LocationBatchUpload:               newT0x0704(item),
		consts.T1003UploadAudioVideoAttr:              newT0x1003(),
		consts.T1205UploadAudioVideoResourceList:      newT0x1205(),
		consts.T1206FileUploadCompleteNotice:          newT0x1206(),
		consts.P8001GeneralRespond:                    newDefaultHandle(consts.P8001GeneralRespond),
		consts.P8100RegisterRespond:                   newDefaultHandle(consts.P8100RegisterRespond),
		consts.P8104QueryTerminalParams:               &model.P0x8104{},
		consts.P9003QueryTerminalAudioVideoProperties: &model.P0x9003{},
		consts.P9101RealTimeAudioVideoRequest:         newP0x9101(),
		consts.P9102AudioVideoControl:                 newP0x9102(),
		consts.P9201SendVideoRecordRequest:            newP0x9201(),
		consts.P9205QueryResourceList:                 newP0x9205(),
		consts.P9206FileUploadInstructions:            newP0x9206(),
		consts.P9207FileUploadControl:                 newP0x9207(),
	}
}

type defaultHandle struct {
	meHandle
}

func newDefaultHandle(command consts.JT808CommandType) *defaultHandle {
	var tmp meHandle
	switch command {
	case consts.P8001GeneralRespond:
		tmp = &model.P0x8001{
			RespondSerialNumber: 1,
			RespondID:           0x0200,
			Result:              0,
		}
	case consts.P8100RegisterRespond:
		tmp = &model.P0x8100{
			RespondSerialNumber: 1,
			Result:              0,
			AuthCode:            "1234567890abcdefghijk",
		}
	}
	return &defaultHandle{meHandle: tmp}
}

func (d defaultHandle) ReplyBody(_ *jt808.JTMessage) ([]byte, error) {
	return nil, nil
}

func newT0x0001() *model.T0x0001 {
	return &model.T0x0001{
		SerialNumber: 0,
		ID:           1,
		Result:       0,
	}
}

func newT0x0100(protocolVersion consts.ProtocolVersionType) *model.T0x0100 {
	return &model.T0x0100{
		ProvinceID:         31,
		CityID:             110,
		ManufacturerID:     "cd123456789",
		TerminalModel:      "www.808.com",
		TerminalID:         "7654321",
		PlateColor:         1,
		LicensePlateNumber: "测A12345678",
		Version:            protocolVersion,
	}
}

func newT0x0102(protocolVersion consts.ProtocolVersionType) *model.T0x0102 {
	return &model.T0x0102{
		AuthCodeLen:     uint8(len("987654321")),
		AuthCode:        "987654321",
		TerminalIMEI:    "123456789012345",
		SoftwareVersion: "3.7.15",
		Version:         protocolVersion,
	}
}

func newT0x0200(item model.T0x0200LocationItem) *model.T0x0200 {
	return &model.T0x0200{
		T0x0200LocationItem: item,
	}
}

func newT0x0704(item model.T0x0200LocationItem) *model.T0x0704 {
	return &model.T0x0704{
		Num:          2,
		LocationType: 0,
		Items: []model.T0x0704LocationItem{
			{
				Len:                 28,
				T0x0200LocationItem: item,
			},
			{
				Len:                 28,
				T0x0200LocationItem: item,
			},
		},
	}
}

func newT0x1003() *model.T0x1003 {
	return &model.T0x1003{
		EnterAudioEncoding:       1,
		EnterAudioChannelsNumber: 1,
		EnterAudioSampleRate:     2,
		EnterAudioSampleDigits:   2,
		AudioFrameLength:         3,
		HasSupportedAudioOutput:  2,
		VideoEncoding:            1,
		TerminalSupportedMaxNumberOfAudioPhysicalChannels: 1,
		TerminalSupportedMaxNumberOfVideoPhysicalChannels: 2,
	}
}

func newP0x9101() *model.P0x9101 {
	return &model.P0x9101{
		ServerIPLen:  12,
		ServerIPAddr: "49.234.235.7",
		TcpPort:      1078,
		UdpPort:      0,
		ChannelNo:    1,
		DataType:     1,
		StreamType:   1,
	}
}

func newP0x9102() *model.P0x9102 {
	return &model.P0x9102{
		ChannelNo:           1,
		ControlCmd:          1,
		CloseAudioVideoData: 2,
		StreamType:          1,
	}
}

func newP0x9201() *model.P0x9201 {
	return &model.P0x9201{
		ServerIPLen:  12,
		ServerIPAddr: "49.234.235.7",
		TcpPort:      1078,
		UdpPort:      0,
		ChannelNo:    1,
		MediaType:    1,
		StreamType:   0,
		MemoryType:   0,
		PlaybackWay:  0,
		PlaySpeed:    0,
		StartTime:    "2024-10-07 19:23:59",
		EndTime:      "2024-10-07 20:23:59",
	}
}

func newP0x9205() *model.P0x9205 {
	return &model.P0x9205{
		ChannelNo:   1,
		StartTime:   "2024-10-07 19:23:59",
		EndTime:     "2024-10-07 20:23:59",
		AlarmFlag:   0,
		MediaType:   1,
		StreamType:  1,
		StorageType: 1,
	}
}

func newT0x1205() *model.T0x1205 {
	return &model.T0x1205{
		SerialNumber:            0,
		AudioVideoResourceTotal: 1,
		AudioVideoResourceList: []model.T0x1205AudioVideoResource{
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
	}
}

func newP0x9206() *model.P0x9206 {
	return &model.P0x9206{
		FTPAddrLen:           9,
		FTPAddr:              "127.0.0.1",
		Port:                 10001,
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
	}
}

func newT0x1206() *model.T0x1206 {
	return &model.T0x1206{
		RespondSerialNumber: 0,
		Result:              0,
	}
}

func newP0x9207() *model.P0x9207 {
	return &model.P0x9207{
		RespondSerialNumber: 0,
		UploadControl:       2,
	}
}
