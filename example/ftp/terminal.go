package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/service"
	"os"
	"sync"
)

type meTerminal struct {
	file *os.File
	once sync.Once
}

func (m *meTerminal) OnJoinEvent(_ *service.Message, key string, err error) {
	str := fmt.Sprintf("终端加入: 手机号[%s] err[%v]", key, err)
	m.println(str)
	m.once.Do(func() {
		m.file, err = os.OpenFile("terminal.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
	})
}

func (m *meTerminal) OnLeaveEvent(key string) {
	str := fmt.Sprintf("终端离开: %s", key)
	m.println(str)
}

func (m *meTerminal) OnNotSupportedEvent(msg *service.Message) {
	str := fmt.Sprintf("暂未实现的报文: %s: [%x]", msg.Command, msg.ExtensionFields.TerminalData)
	m.println(str)
}

func (m *meTerminal) OnReadExecutionEvent(_ *service.Message) {}

func (m *meTerminal) OnWriteExecutionEvent(msg service.Message) {
	ex := msg.ExtensionFields
	if ex.ActiveSend {
		m.println(fmt.Sprintf("主动发送: [%s] [%x]", ex.PlatformCommand, ex.PlatformData))
		m.println(fmt.Sprintf("主动回复: [%s] [%x]", ex.TerminalCommand, ex.TerminalData))
	} else {
		m.println(fmt.Sprintf("->: [%s] [%x]", ex.TerminalCommand, ex.TerminalData))
		m.println(fmt.Sprintf("<-: [%s] [%x]", ex.PlatformCommand, ex.PlatformData))
	}
}

func (m *meTerminal) println(str string) {
	str += "\n"
	fmt.Println(str)
	if m.file != nil {
		_, _ = m.file.WriteString(str)
		_ = m.file.Sync()
	}
}
