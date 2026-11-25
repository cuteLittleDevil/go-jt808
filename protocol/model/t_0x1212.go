package model

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type T0x1212 struct {
	T0x1211
	P0x9212RetransmitPacketList []P0x9212RetransmitPacket
}

func (t *T0x1212) Protocol() consts.JT808CommandType {
	return consts.T1212FileUploadComplete
}

func (t *T0x1212) ReplyProtocol() consts.JT808CommandType {
	return consts.P9212FileUploadCompleteRespond
}

func (t *T0x1212) ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error) {
	var f T0x1211
	if err := f.Parse(jtMsg); err != nil {
		return nil, err
	}
	p9212 := P0x9212{
		FileNameLen: f.FileNameLen,
		FileName:    f.FileName,
		FileType:    f.FileType,
	}
	if rLen := len(t.P0x9212RetransmitPacketList); rLen > 0 {
		p9212.UploadResult = 1
		p9212.RetransmitPacketNumber = byte(rLen)
		p9212.P0x9212RetransmitPacketList = t.P0x9212RetransmitPacketList
	}
	return p9212.Encode(), nil
}
