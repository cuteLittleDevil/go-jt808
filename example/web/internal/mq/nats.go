package mq

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"log/slog"
	"web/internal/shared"
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

func (m *Manage) Run(handlers map[string]func(data shared.EventData)) {
	for sub, handle := range handlers {
		if _, err := m.conn.Subscribe(sub, func(msg *nats.Msg) {
			var data shared.EventData
			if err := data.Parse(msg.Data); err == nil {
				handle(data)
			} else {
				slog.Error("parse fail",
					slog.String("sub", sub),
					slog.String("err", err.Error()))
			}
		}); err != nil {
			slog.Error("nats sub fail",
				slog.String("sub", sub),
				slog.String("err", err.Error()))
		}
	}
}

func (m *Manage) SubNotice(ctx context.Context, subject string) (chan []byte, error) {
	ch := make(chan []byte, 10)
	sub, err := m.conn.Subscribe(subject, func(msg *nats.Msg) {
		select {
		case <-ctx.Done():
			return
		default:
			ch <- msg.Data
		}
	})
	go func() {
		select {
		case <-ctx.Done():
			_ = sub.Unsubscribe()
			close(ch)
		}
	}()
	return ch, err
}
