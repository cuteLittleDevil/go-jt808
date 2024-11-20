package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x8800 struct {
	BaseHandle
	// MultimediaID  多媒体数据ID 值大于0
	MultimediaID uint32 `json:"multimediaIDNumber"`
	// AgainPackageCount 重传包总数
	AgainPackageCount byte `json:"againPackageCount"`
	// AgainPackageList 重传包ID列表 BYTE[2*n] 重传包序号顺序排列，如“包 ID1 包 ID2......包 IDn
	AgainPackageList []uint16 `json:"againPackageList"`
}

func (p *P0x8800) Protocol() consts.JT808CommandType {
	return consts.P8800MultimediaUploadRespond
}

func (p *P0x8800) ReplyProtocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (p *P0x8800) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 5 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.MultimediaID = binary.BigEndian.Uint32(body[0:4])
	p.AgainPackageCount = body[4]
	if len(body) != 5+2*int(p.AgainPackageCount) {
		return protocol.ErrBodyLengthInconsistency
	}
	for i := 0; i < int(p.AgainPackageCount); i++ {
		id := binary.BigEndian.Uint16(body[5+(2*i) : 5+(2*i)+2])
		p.AgainPackageList = append(p.AgainPackageList, id)
	}
	return nil
}

func (p *P0x8800) Encode() []byte {
	data := make([]byte, 5)
	binary.BigEndian.PutUint32(data[0:4], p.MultimediaID)
	data[4] = p.AgainPackageCount
	for _, v := range p.AgainPackageList {
		data = binary.BigEndian.AppendUint16(data, v)
	}
	return data
}

func (p *P0x8800) HasReply() bool {
	return false
}

func (p *P0x8800) String() string {
	str := "\t重传包ID列表:"
	for _, v := range p.AgainPackageList {
		str += fmt.Sprintf("\n\t[%04x] 重传包ID:[%v]", v, v)
	}
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%04x] 多媒体数据ID:[%d]", p.MultimediaID, p.MultimediaID),
		fmt.Sprintf("\t[%02x] 重传包总数:[%d]", p.AgainPackageCount, p.AgainPackageCount),
		str,
		"}",
	}, "\n")
}
