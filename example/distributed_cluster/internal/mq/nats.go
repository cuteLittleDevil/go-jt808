package mq

import (
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"log/slog"
	"time"
)

var _natsManage *Manage

type Manage struct {
	conn *nats.Conn
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

func (m *Manage) Sub(subject string, duration time.Duration) (data []byte, err error) {
	ch := make(chan []byte)
	defer func() {
		close(ch)
	}()
	sub, err := m.conn.Subscribe(subject, func(msg *nats.Msg) {
		ch <- msg.Data
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = sub.Unsubscribe()
	}()
	select {
	case tmp := <-ch:
		data = tmp
	case <-time.After(duration):
		err = errors.New(fmt.Sprintf("timeout sub %s", subject))
	}
	return
}
