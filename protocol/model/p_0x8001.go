package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x8001 struct {
	// RespondSerialNumber 应答消息流水号 = 这个消息发送时候的流水号
	RespondSerialNumber uint16
	// RespondID 应答消息ID = 这个消息发送时候的ID
	RespondID uint16
	// Result 结果 // 0-成功 1-失败 2-消息有误 3-不支持 4-报警处理确认
	Result byte
}

func (p *P0x8001) Protocol() consts.JT808CommandType {
	return consts.P8001GeneralRespond
}

func (p *P0x8001) ReplyProtocol() consts.JT808CommandType {
	return 0
}

func (p *P0x8001) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) != 5 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.RespondSerialNumber = binary.BigEndian.Uint16(body[:2])
	p.RespondID = binary.BigEndian.Uint16(body[2:4])
	p.Result = body[4]
	return nil
}

func (p *P0x8001) Encode() []byte {
	data := make([]byte, 5)
	binary.BigEndian.PutUint16(data[0:2], p.RespondSerialNumber)
	binary.BigEndian.PutUint16(data[2:4], p.RespondID)
	data[4] = p.Result
	return data
}

func (p *P0x8001) String() string {
	str := "数据体对象:{\n"
	str += fmt.Sprintf("\t%s:[%10x]", p.Protocol(), p.Encode())
	return strings.Join([]string{
		str,
		fmt.Sprintf("\t[%04x] 应答消息流水号:[%d]", p.RespondSerialNumber, p.RespondSerialNumber),
		fmt.Sprintf("\t[%04x] 应答消息ID:[%d]", p.RespondID, p.RespondID),
		fmt.Sprintf("\t[%02x] 结果:[%d] 0-成功 1-失败 "+
			"2-消息有误 3-不支持 4-报警处理确认", p.Result, p.Result),
		"}",
	}, "\n")
}
