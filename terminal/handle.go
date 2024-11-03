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
	handles := []Handler{
		&model.T0x0100{
			ProvinceID:         31,
			CityID:             110,
			ManufacturerID:     "cd123456789",
			TerminalModel:      "www.808.com",
			TerminalID:         "7654321",
			PlateColor:         1,
			LicensePlateNumber: "测A12345678",
			Version:            protocolVersion,
		},
		&model.T0x0102{
			AuthCodeLen:     uint8(len("987654321")),
			AuthCode:        "987654321",
			TerminalIMEI:    "123456789012345",
			SoftwareVersion: "3.7.15",
			Version:         protocolVersion,
		},
		&model.T0x0002{},
		&model.T0x0200{
			T0x0200LocationItem: item,
		},
		&model.T0x0704{
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
		},
		&model.T0x0001{
			SerialNumber: 0,
			ID:           1,
			Result:       0,
		},
		&p0x8001{
			P0x8001: model.P0x8001{
				RespondSerialNumber: 1,
				RespondID:           0x0200,
				Result:              0,
			},
		},
		&p0x8100{
			P0x8100: model.P0x8100{
				RespondSerialNumber: 1,
				Result:              0,
				AuthCode:            "1234567890abcdefghijk",
			},
		},
		&model.P0x8104{},
		&model.P0x9003{},
		&model.P0x9101{
			ServerIPLen:  12,
			ServerIPAddr: "49.234.235.7",
			TcpPort:      1078,
			UdpPort:      0,
			ChannelNo:    1,
			DataType:     1,
			StreamType:   1,
		},
		&model.P0x9102{
			ChannelNo:           1,
			ControlCmd:          1,
			CloseAudioVideoData: 2,
			StreamType:          1,
		},
		&model.P0x9201{
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
		},
		&model.P0x9205{
			ChannelNo:   1,
			StartTime:   "2024-10-07 19:23:59",
			EndTime:     "2024-10-07 20:23:59",
			AlarmFlag:   0,
			MediaType:   1,
			StreamType:  1,
			StorageType: 1,
		},
		&model.T0x1205{
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
		},
		&model.P0x9206{
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
		},
		&model.T0x1206{
			RespondSerialNumber: 0,
			Result:              0,
		},
		&model.P0x9207{
			RespondSerialNumber: 0,
			UploadControl:       2,
		},
	}
	protocolHandles := make(map[consts.JT808CommandType]Handler, len(handles))
	for _, v := range handles {
		protocolHandles[v.Protocol()] = v
	}
	return protocolHandles
}

type p0x8001 struct {
	model.P0x8001
}

func (p *p0x8001) ReplyBody(_ *jt808.JTMessage) ([]byte, error) {
	return nil, nil
}

type p0x8100 struct {
	model.P0x8100
}

func (p p0x8100) ReplyBody(_ *jt808.JTMessage) ([]byte, error) {
	return nil, nil
}
