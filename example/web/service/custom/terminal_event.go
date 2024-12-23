package custom

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"strings"
	"web/internal/mq"
	"web/internal/shared"
	"web/service/conf"
	"web/service/record"
)

type terminalEvent struct {
	id         string
	httpPrefix string
	attachIP   string
	attachPort int
	key        string
}

func NewTerminalEvent() service.TerminalEventer {
	return &terminalEvent{
		id:         conf.GetData().JTConfig.ID,
		httpPrefix: conf.GetData().JTConfig.HttpPrefix,
		attachIP:   conf.GetData().FileConfig.AttachIP,
		attachPort: conf.GetData().FileConfig.AttachPort,
		key:        "",
	}
}

func (t *terminalEvent) OnJoinEvent(msg *service.Message, key string, err error) {
	if err == nil {
		record.Join(*msg)
		fmt.Println("加入", msg.Command.String(), key)
		t.key = key
		data := shared.NewEventData(shared.OnInit, key,
			shared.WithIDAndAddress(t.id, t.httpPrefix),
			shared.WithMessage(*msg))
		t.pub(data)
	}
}

func (t *terminalEvent) OnLeaveEvent(key string) {
	fmt.Println("离开", key)
	jtMsg := jt808.NewJTMessage()
	jtMsg.Header.TerminalPhoneNo = key
	data := shared.NewEventData(shared.OnLeave, key,
		shared.WithIDAndAddress(t.id, t.httpPrefix),
		shared.WithMessage(service.Message{
			JTMessage: jtMsg,
		}))
	t.pub(data)
	record.Leave(key)
}

func (t *terminalEvent) OnNotSupportedEvent(msg *service.Message) {
	data := shared.NewEventData(shared.OnNotSupported, t.key,
		shared.WithIDAndAddress(t.id, t.httpPrefix),
		shared.WithMessage(*msg))
	t.pub(data)
}

func (t *terminalEvent) OnReadExecutionEvent(msg *service.Message) {
	if msg.Command == consts.T0801MultimediaDataUpload {
		// 直接保存在本地处理了 不需要传其他地方去
		return
	}
	data := shared.NewEventData(shared.OnRead, t.key,
		shared.WithIDAndAddress(t.id, t.httpPrefix),
		shared.WithAttachIPAndPort(t.attachIP, t.attachPort),
		shared.WithMessage(*msg))
	t.pub(data)
}

func (t *terminalEvent) OnWriteExecutionEvent(msg service.Message) {
	go record.AddMessage(msg)
	data := shared.NewEventData(shared.OnWrite, t.key,
		shared.WithIDAndAddress(t.id, t.httpPrefix),
		shared.WithAttachIPAndPort(t.attachIP, t.attachPort),
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
			str := data.JTMessage.Header.String()
			str = strings.ReplaceAll(str, "\t", "")
			str = strings.ReplaceAll(str, "\n", "")
			slog.Debug("pub",
				slog.String("sub", sub),
				slog.Any("data", str))
		default:
		}
	}
}
