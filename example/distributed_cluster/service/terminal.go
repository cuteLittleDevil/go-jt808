package main

import (
	"distributed_cluster/internal/mq"
	"distributed_cluster/internal/shared"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/service"
	"log/slog"
)

type meTerminal struct {
	id string
}

func (m *meTerminal) OnJoinEvent(msg *service.Message, key string, err error) {
	if err == nil {
		fmt.Println("加入", key)
		data := shared.NewData(m.id, shared.OnInit, *msg)
		m.pub(data)
	}
}

func (m *meTerminal) OnLeaveEvent(key string) {
	fmt.Println("离开", key)
	jtMsg := jt808.NewJTMessage()
	jtMsg.Header.TerminalPhoneNo = key
	data := shared.NewData(m.id, shared.OnLeave, service.Message{
		JTMessage: jtMsg,
	})
	m.pub(data)
}

func (m *meTerminal) OnNotSupportedEvent(msg *service.Message) {
	data := shared.NewData(m.id, shared.OnNotSupported, *msg)
	m.pub(data)
}

func (m *meTerminal) OnReadExecutionEvent(msg *service.Message) {
	data := shared.NewData(m.id, shared.OnRead, *msg)
	m.pub(data)
}

func (m *meTerminal) OnWriteExecutionEvent(msg service.Message) {
	data := shared.NewData(m.id, shared.OnWrite, msg)
	m.pub(data)
	if msg.ExtensionFields.ActiveSend {
		fmt.Println(fmt.Sprintf("主动发送的数据: %x", msg.ExtensionFields.PlatformData))
		fmt.Println(fmt.Sprintf("设备回复的数据: %x", msg.ExtensionFields.TerminalData))
	}
}

func (m *meTerminal) pub(data *shared.Data) {
	sub := data.Subject
	if err := mq.Default().Pub(sub, data.ToBytes()); err != nil {
		slog.Error("pub fail",
			slog.String("sub", sub),
			slog.String("err", err.Error()))
	}
}
