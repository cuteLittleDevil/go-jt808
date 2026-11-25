// Package model 定义 JT/T 808、JT/T 1078 及主动安全扩展的消息体模型。
//
// 设计约定（显式长度/数量字段严格按输入编码）：
// 结构体中凡显式声明了协议长度或数量字段（如 P0x9101.ServerIPLen、P0x8103.ParamTotal、
// ParamContent.Len、P0x8302.QuestionContentLen、T0x0704.Num 等），Encode() 均严格按调用方传入的字段值原样编码，
// 不做隐式重算或覆盖；若传入值与实际内容长度不一致，按输入原样组包，由调用方保证字段一致性（便于协议联调与异常报文模拟，
// 并确保 Encode() 与 String() 输出完全一致）。
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

func (b *BaseHandle) ReplyProtocol() consts.JT808CommandType {
	return consts.P8001GeneralRespond
}
