package custom

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"web/internal/mq"
	"web/internal/shared"
	"web/service/conf"
	"web/service/record"
)

type terminalEvent struct {
	id  string
	key string
}

func NewTerminalEvent(id string) service.TerminalEventer {
	return &terminalEvent{id: id}
}

func (t *terminalEvent) OnJoinEvent(msg *service.Message, key string, err error) {
	if err == nil {
		record.Join(*msg)
		fmt.Println("加入", msg.Command.String(), key)
		t.key = key
		data := shared.NewEventData(t.id, shared.OnInit, key,
			shared.WithMessage(*msg))
		t.pub(data)
	}
}

func (t *terminalEvent) OnLeaveEvent(key string) {
	fmt.Println("离开", key)
	jtMsg := jt808.NewJTMessage()
	jtMsg.Header.TerminalPhoneNo = key
	data := shared.NewEventData(t.id, shared.OnLeave, key,
		shared.WithMessage(service.Message{
			JTMessage: jtMsg,
		}))
	t.pub(data)
	record.Leave(key)
}

func (t *terminalEvent) OnNotSupportedEvent(msg *service.Message) {
	data := shared.NewEventData(t.id, shared.OnNotSupported, t.key,
		shared.WithMessage(*msg))
	t.pub(data)
}

func (t *terminalEvent) OnReadExecutionEvent(msg *service.Message) {
	go record.AddMessage(*msg)
	if msg.Command == consts.T0801MultimediaDataUpload {
		// 直接保存在本地处理了 不需要传其他地方去
		return
	}
	data := shared.NewEventData(t.id, shared.OnRead, t.key,
		shared.WithMessage(*msg))
	fmt.Println(fmt.Sprintf("---- %x", msg.ExtensionFields.TerminalData))
	t.pub(data)
}

func (t *terminalEvent) OnWriteExecutionEvent(msg service.Message) {
	go record.AddMessage(msg)
	data := shared.NewEventData(t.id, shared.OnWrite, t.key,
		shared.WithMessage(msg))
	t.pub(data)
	if msg.ExtensionFields.ActiveSend {
		fmt.Println(fmt.Sprintf("主动发送的数据: %x", msg.ExtensionFields.PlatformData))
		fmt.Println(fmt.Sprintf("设备回复的数据: %x", msg.ExtensionFields.TerminalData))
	}
}

func (t *terminalEvent) pub(data *shared.EventData) {
	sub := data.Subject
	if conf.GetData().NatsConfig.Open {
		if err := mq.Default().Pub(sub, data.ToBytes()); err != nil {
			slog.Error("pub fail",
				slog.String("sub", sub),
				slog.String("err", err.Error()))
		}
	} else {
		switch data.Type {
		case shared.OnRead, shared.OnWrite:
			//fmt.Println(fmt.Sprintf("主题[%s] 数据[%s]\n", sub, data.Message.JTMessage.Header.String()))
		default:
		}
	}
}
