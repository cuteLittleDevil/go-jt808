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
	_ = t.T0x1211.Parse(jtMsg)
	p9202 := P0x9212{
		FileNameLen:                 t.FileNameLen,
		FileName:                    t.FileName,
		FileType:                    t.FileType,
		UploadResult:                0,
		RetransmitPacketNumber:      0,
		P0x9212RetransmitPacketList: nil,
	}
	if rLen := len(t.P0x9212RetransmitPacketList); rLen > 0 {
		p9202.UploadResult = 1
		p9202.RetransmitPacketNumber = byte(rLen)
		p9202.P0x9212RetransmitPacketList = t.P0x9212RetransmitPacketList
	}
	return p9202.Encode(), nil
}
