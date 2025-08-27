package service

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
	"time"
)

type ActiveMessage struct {
	// header 设备消息固体头 使用的是第一次报文的固定头
	header *jt808.Header
	// replyChan 用于获取终端应答情况
	replyChan chan *Message
	// convertMessage 平台最终转换的Message
	convertMessage *Message
	// Key 唯一标识符 默认手机号
	Key string `json:"key"`
	// Command 平台下发的指令
	Command consts.JT808CommandType `json:"command"`
	// Body 平台下发的数据
	Body []byte `json:"body"`
	// OverTimeDuration  超时时间 默认3秒
	OverTimeDuration time.Duration `json:"overTimeDuration"`
	ExtensionFields  struct {
		// PlatformSeq 平台下发的流水号
		PlatformSeq uint16 `json:"platformSeq,omitempty"`
		// Data 平台最终下发的数据
		Data []byte `json:"data,omitempty"`
	}
}

func NewActiveMessage(key string, command consts.JT808CommandType, body []byte, overTimeDuration time.Duration) *ActiveMessage {
	return &ActiveMessage{Key: key, Command: command, Body: body, OverTimeDuration: overTimeDuration}
}

func (a *ActiveMessage) String() string {
	return strings.Join([]string{
		fmt.Sprintf("key[%s]", a.Key),
		fmt.Sprintf("指令[%s]", a.Command),
		fmt.Sprintf("body[%x]", a.Body),
	}, "\n")
}
