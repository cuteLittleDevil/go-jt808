package help

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/service"
)

type LogTerminal struct {
}

func (l *LogTerminal) OnJoinEvent(msg *service.Message, key string, err error) {
	if err != nil {
		fmt.Println("终端加入", key, fmt.Sprintf("%x", msg.ExtensionFields.TerminalData))
	}
}

func (l *LogTerminal) OnLeaveEvent(key string) {
	fmt.Println("终端退出", key)
}

func (l *LogTerminal) OnNotSupportedEvent(msg *service.Message) {
	fmt.Println("未实现的报文", fmt.Sprintf("%x", msg.ExtensionFields.TerminalData))
}

func (l *LogTerminal) OnReadExecutionEvent(msg *service.Message) {
	extension := msg.ExtensionFields
	fmt.Println("读取到终端的数据", fmt.Sprintf("%d %x", extension.TerminalSeq, extension.TerminalData))
}

func (l *LogTerminal) OnWriteExecutionEvent(msg service.Message) {
	extension := msg.ExtensionFields
	if extension.ActiveSend {
		fmt.Println("平台主动下发的", fmt.Sprintf("%d %x", extension.PlatformSeq, extension.PlatformData))
		return
	}
	fmt.Println("平台被动回复的", fmt.Sprintf("%d %x", extension.PlatformSeq, extension.PlatformData))
}
