package service

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
	"sync"
	"time"
)

type ActiveMessage struct {
	// once 保证完成情况只能被触发一次
	once sync.Once
	// header 设备消息固体头 使用的是第一次报文的固定头
	header *jt808.Header
	// replyChan 用于获取终端应答情况
	replyChan chan *Message
	// completeChan 用于判断完成情况
	completeChan chan struct{}
	// Key 唯一标识符 默认手机号
	Key string `json:"key"`
	// Command 平台下发的指令
	Command consts.JT808CommandType `json:"command"`
	// Body 平台下发的数据
	Body []byte `json:"body"`
	// Data 平台最终下发的数据
	Data []byte `json:"data"`
	// OverTimeDuration  超时时间 默认5秒
	OverTimeDuration time.Duration `json:"overTimeDuration"`
}

func NewActiveMessage(key string, command consts.JT808CommandType, body []byte, overTimeDuration time.Duration) *ActiveMessage {
	return &ActiveMessage{Key: key, Command: command, Body: body, OverTimeDuration: overTimeDuration}
}

func (a *ActiveMessage) hasComplete() bool {
	select {
	case <-a.completeChan:
		return true
	default:
	}
	ok := true
	a.once.Do(func() {
		close(a.completeChan)
		ok = false
	})
	return ok
}

func (a *ActiveMessage) String() string {
	return strings.Join([]string{
		fmt.Sprintf("key[%s]", a.Key),
		fmt.Sprintf("指令[%x] [%s]", a.Command, a.Command),
		fmt.Sprintf("body[%x]", a.Body),
	}, "\n")
}
