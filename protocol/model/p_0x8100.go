package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x8100 struct {
	// RespondSerialNumber 应答消息流水号 = 这个消息发送时候的流水号
	RespondSerialNumber uint16
	// Result 结果 0-成功 1-车辆已被注册 2-数据库中无该车辆 3-终端已被注册 4-数据库中无该终端
	Result byte
	// AuthCode 鉴权码
	AuthCode string
}

func (p *P0x8100) Protocol() consts.JT808CommandType {
	return consts.P8100RegisterRespond
}

func (p *P0x8100) ReplyProtocol() consts.JT808CommandType {
	return 0
}

func (p *P0x8100) Encode() []byte {
	code := utils.String2FillingBytes(p.AuthCode, len(p.AuthCode))
	tmp := []byte{
		byte(p.RespondSerialNumber >> 8),
		byte(p.RespondSerialNumber & 0xFF),
		p.Result,
	}
	tmp = append(tmp, code...)
	return tmp
}

func (p *P0x8100) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 3 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.RespondSerialNumber = binary.BigEndian.Uint16(body[:2])
	p.Result = body[2]
	p.AuthCode = string(body[3:])
	return nil
}

func (p *P0x8100) String() string {
	str := "数据体对象:{\n"
	body := p.Encode()
	str += fmt.Sprintf("\t注册消息应答:[%x]", body)
	return strings.Join([]string{
		str,
		fmt.Sprintf("\t[%04x] 应答流水号:[%d]", p.RespondSerialNumber, p.RespondSerialNumber),
		fmt.Sprintf("\t[%02x] 结果:[%d] 0-成功 1-失败 "+
			"2-消息有误 3-不支持 4-报警处理确认", p.Result, p.Result),
		fmt.Sprintf("\t[%x] 鉴权码:[%s]", body[3:], p.AuthCode),
		"}",
	}, "\n")
}
