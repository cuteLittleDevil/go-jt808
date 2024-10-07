package model

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type BaseHandle struct{}

func (b *BaseHandle) Parse(_ *jt808.JTMessage) error {
	return nil
}

func (b *BaseHandle) HasReply() bool {
	return true
}

func (b *BaseHandle) ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error) {
	head := jtMsg.Header
	// 通用应答
	p8001 := &P0x8001{
		RespondSerialNumber: head.SerialNumber,
		RespondID:           head.ID,
		Result:              0x00, // 0-成功 1-失败 2-消息有误 3-不支持 4-报警处理确认 默认成功
	}
	return p8001.Encode(), nil
}

func (b *BaseHandle) ReplyProtocol() uint16 {
	return uint16(consts.P8001GeneralRespond)
}
