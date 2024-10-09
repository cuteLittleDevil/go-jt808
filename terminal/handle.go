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

func defaultProtocolHandles() map[consts.JT808CommandType]Handler {
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
		},
		&model.T0x0102{
			AuthCodeLen:     uint8(len("987654321")),
			AuthCode:        "987654321",
			TerminalIMEI:    "123456789012345",
			SoftwareVersion: "3.7.15",
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
	}
	protocolHandles := make(map[consts.JT808CommandType]Handler, len(handles))
	for _, v := range handles {
		protocolHandles[v.Protocol()] = v
	}
	return protocolHandles
}
