package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type CustomPlatform struct {
	model.BaseHandle
	currentSeq uint16
}

func (cp *CustomPlatform) Protocol() consts.JT808CommandType {
	return CustomPlatformCommand
}

func (cp *CustomPlatform) OnReadExecutionEvent(msg *service.Message) {
	fmt.Println("------- platform read event")
	cp.currentSeq = msg.JTMessage.Header.SerialNumber
}

func (cp *CustomPlatform) OnWriteExecutionEvent(msg service.Message) {
	fmt.Println("------- platform write event")
	if msg.JTMessage.Header.ID == CustomTerminalCommand { // 收到回复了 看看是不是匹配的
		var tmp CustomTerminalReply
		_ = tmp.Parse(msg.JTMessage)
		fmt.Println("data", tmp.Data)
		if tmp.Data == cp.currentSeq {
			fmt.Println("匹配的 本次0x6666成功了")
		}
	}
}

func (cp *CustomPlatform) Encode() []byte {
	return nil
}
