package mq

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log/slog"
	"time"
)

var _natsManage *Manage

type Manage struct {
	conn *nats.Conn
	sum  int
}

func Init(address string) error {
	c, err := nats.Connect(fmt.Sprintf("nats://%s", address))
	if err != nil {
		return err
	}
	_natsManage = &Manage{conn: c}
	return nil
}

func Default() *Manage {
	return _natsManage
}

func (m *Manage) Pub(subject string, data []byte) error {
	m.sum++
	if m.sum%1e5 == 0 {
		fmt.Println("发送的消息条数", m.sum, time.Now().Format(time.RFC3339))
	}
	return m.conn.Publish(subject, data)
}

func (m *Manage) Run(handlers map[string]nats.MsgHandler) {
	for sub, msgHandler := range handlers {
		if _, err := m.conn.Subscribe(sub, msgHandler); err != nil {
			slog.Error("nats sub fail",
				slog.String("sub", sub),
				slog.String("err", err.Error()))
		}
	}
}
