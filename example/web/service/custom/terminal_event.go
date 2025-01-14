package custom

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"time"
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
	openNats   bool
}

func NewTerminalEvent() service.TerminalEventer {
	return &terminalEvent{
		id:         conf.GetData().JTConfig.ID,
		httpPrefix: conf.GetData().JTConfig.HTTPPrefix,
		attachIP:   conf.GetData().FileConfig.AttachConfig.IP,
		attachPort: conf.GetData().FileConfig.AttachConfig.Port,
		key:        "",
		openNats:   conf.GetData().NatsConfig.Open,
	}
}

func (t *terminalEvent) OnJoinEvent(msg *service.Message, key string, err error) {
	if err == nil {
		record.Join(*msg)
		fmt.Println("加入", time.Now().Format(time.DateTime), msg.Command.String(), key)
		t.key = key
		data := shared.NewEventData(shared.OnInit, key,
			shared.WithIDAndHTTPPrefix(t.id, t.httpPrefix),
			shared.WithMessage(*msg))
		t.pub(data)
	}
}

func (t *terminalEvent) OnLeaveEvent(key string) {
	fmt.Println("离开", time.Now().Format(time.DateTime), key)
	jtMsg := jt808.NewJTMessage()
	jtMsg.Header.TerminalPhoneNo = key
	data := shared.NewEventData(shared.OnLeave, key,
		shared.WithIDAndHTTPPrefix(t.id, t.httpPrefix),
		shared.WithMessage(service.Message{
			JTMessage: jtMsg,
		}))
	t.pub(data)
	record.Leave(key)
}

func (t *terminalEvent) OnNotSupportedEvent(msg *service.Message) {
	data := shared.NewEventData(shared.OnNotSupported, t.key,
		shared.WithIDAndHTTPPrefix(t.id, t.httpPrefix),
		shared.WithMessage(*msg))
	t.pub(data)
}

func (t *terminalEvent) OnReadExecutionEvent(msg *service.Message) {
	if t.hasHandle(msg.Command) {
		return
	}
	data := shared.NewEventData(shared.OnRead, t.key,
		shared.WithIDAndHTTPPrefix(t.id, t.httpPrefix),
		shared.WithAttachIPAndPort(t.attachIP, t.attachPort),
		shared.WithMessage(*msg))
	t.pub(data)
}

func (t *terminalEvent) OnWriteExecutionEvent(msg service.Message) {
	if t.hasHandle(msg.Command) {
		return
	}
	go record.AddMessage(msg)
	data := shared.NewEventData(shared.OnWrite, t.key,
		shared.WithIDAndHTTPPrefix(t.id, t.httpPrefix),
		shared.WithAttachIPAndPort(t.attachIP, t.attachPort),
		shared.WithMessage(msg))
	t.pub(data)
	if msg.ExtensionFields.ActiveSend && !conf.GetData().NatsConfig.Open {
		fmt.Println(fmt.Sprintf("主动发送的数据: %x", msg.ExtensionFields.PlatformData))
		fmt.Println(fmt.Sprintf("设备回复的数据: %x", msg.ExtensionFields.TerminalData))
	}
}

func (t *terminalEvent) pub(data *shared.EventData) {
	sub := data.Subject
	if t.openNats {
		if err := mq.Default().Pub(sub, data.ToBytes()); err != nil {
			slog.Error("pub fail",
				slog.String("sub", sub),
				slog.String("err", err.Error()))
		}
	} else {
		switch data.Type {
		case shared.OnWrite:
			slog.Debug("pub",
				slog.String("sub", sub),
				slog.String("read", fmt.Sprintf("%x", data.ExtensionFields.TerminalData)),
				slog.String("write", fmt.Sprintf("%x", data.ExtensionFields.PlatformData)),
				slog.String("describe", consts.JT808CommandType(data.JTMessage.Header.ID).String()))
		default:
		}
	}
}

func (t *terminalEvent) hasHandle(command consts.JT808CommandType) bool {
	return command == consts.T0801MultimediaDataUpload
}
