// Package model 定义 JT/T 808、JT/T 1078 及主动安全扩展的消息体模型。
//
// 核心设计与运行时约定：
//  1. 单连接独立实例：每个连接协程持有独立的协议处理 map 与模型实例，不存在跨协程并发共享同一模型实例的场景
//     （如 P0x8300 / P0x8302 在 Encode 时按需回填 Flag 属于单协程安全操作）。
//  2. Header 与卡号唯一确定：JTMessage.Header、Property 及手机号/SIM 卡 BCD 在首包解码（或构造函数）时唯一确定，
//     运行期保证非 nil 且不发生中途改号。
//  3. 显式长度/数量字段严格按输入编码：凡结构体声明了协议长度或数量字段（如 P0x9101.ServerIPLen、P0x8103.ParamTotal、
//     ParamContent.Len、P0x8302.QuestionContentLen、T0x0704.Num 等），Encode() 严格按输入值原样组包，不做隐式重算覆盖。
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
