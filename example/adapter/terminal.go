package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/service"
)

type meTerminal struct {
	ch chan<- string
}

func (m *meTerminal) OnJoinEvent(_ *service.Message, key string, err error) {
	fmt.Println(fmt.Sprintf("终端加入: %s err[%v]", key, err))
	go func() {
		m.ch <- key
	}()
}

func (m *meTerminal) OnLeaveEvent(key string) {
	fmt.Println(fmt.Sprintf("终端离开: %s", key))
}

func (m *meTerminal) OnNotSupportedEvent(msg *service.Message) {
	fmt.Println(fmt.Sprintf("暂未实现的报文: %s: [%x]", msg.Command, msg.ExtensionFields.TerminalData))
}

func (m *meTerminal) OnReadExecutionEvent(msg *service.Message) {
	fmt.Println(fmt.Sprintf("读到的报文 %s: [%x]", msg.Command, msg.ExtensionFields.TerminalData))
}

func (m *meTerminal) OnWriteExecutionEvent(msg service.Message) {
	str := "回复的报文"
	if msg.ExtensionFields.ActiveSend {
		str = "主动发送的报文"
	}
	str += fmt.Sprintf(" %s: [%x]", msg.ExtensionFields.PlatformCommand, msg.ExtensionFields.PlatformData)
	fmt.Println(str)
}
