package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x8003 struct {
	BaseHandle
	// OriginalSerialNumber 原始消息流水号 对应要求补传的原始消息第一包的消息流水号
	OriginalSerialNumber uint16 `json:"originalSerialNumber"`
	// AgainPackageCount 重传包总数
	AgainPackageCount byte `json:"againPackageCount"`
	// AgainPackageList 重传包ID列表 BYTE[2*n] 重传包序号顺序排列，如“包 ID1 包 ID2......包 IDn
	AgainPackageList []uint16 `json:"againPackageList"`
}

func (p *P0x8003) Protocol() consts.JT808CommandType {
	return consts.P8003ReissueSubcontractingRequest
}

func (p *P0x8003) ReplyProtocol() consts.JT808CommandType {
	return 0
}

func (p *P0x8003) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 3 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.OriginalSerialNumber = binary.BigEndian.Uint16(body[:2])
	p.AgainPackageCount = body[2]
	if len(body) != 3+2*int(p.AgainPackageCount) {
		return protocol.ErrBodyLengthInconsistency
	}
	for i := 0; i < int(p.AgainPackageCount); i++ {
		id := binary.BigEndian.Uint16(body[3+(2*i) : 3+(2*i)+2])
		p.AgainPackageList = append(p.AgainPackageList, id)
	}
	return nil
}

func (p *P0x8003) Encode() []byte {
	data := make([]byte, 3)
	binary.BigEndian.PutUint16(data[:2], p.OriginalSerialNumber)
	data[2] = p.AgainPackageCount
	for _, v := range p.AgainPackageList {
		data = binary.BigEndian.AppendUint16(data, v)
	}
	return data
}

func (p *P0x8003) HasReply() bool {
	return false
}

func (p *P0x8003) String() string {
	str := "\t重传包ID列表:"
	for _, v := range p.AgainPackageList {
		str += fmt.Sprintf("\n\t[%04x] 重传包ID:[%v]", v, v)
	}
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%04x] 原始消息流水号:[%d]", p.OriginalSerialNumber, p.OriginalSerialNumber),
		fmt.Sprintf("\t[%02x] 重传包总数:[%d]", p.AgainPackageCount, p.AgainPackageCount),
		str,
		"}",
	}, "\n")
}
