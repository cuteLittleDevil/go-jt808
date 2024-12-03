package pkg

import (
	"context"
	"github.com/cuteLittleDevil/go-jt808/m7s-jt1078/pkg/jt1078"
	"github.com/go-resty/resty/v2"
	"log/slog"
	"m7s.live/v5"
	"net"
	"time"
)

func NewService(addr string, log *slog.Logger, opts ...Option) *Service {
	options := &Options{
		pubFunc: func(ctx context.Context, pack *jt1078.Packet) (publisher *m7s.Publisher, err error) {
			return nil, nil
		},
	}
	for _, op := range opts {
		op.F(options)
	}
	s := &Service{
		Logger: log,
		addr:   addr,
		opts:   options,
	}
	return s
}

type Service struct {
	*slog.Logger
	addr string
	opts *Options
}

func (s *Service) Run() {
	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.Error("listen error",
			slog.String("addr", s.addr),
			slog.String("err", err.Error()))
		return
	}
	s.Info("listen tcp",
		slog.String("addr", s.addr))
	for {
		conn, err := listen.Accept()
		if err != nil {
			s.Warn("accept error",
				slog.String("err", err.Error()))
			return
		}
		client := newConnection(conn, s.Logger, s.opts.ptsFunc)
		var (
			audioChan <-chan []byte
			httpBody  = map[string]any{}
			audioPort int
		)
		if s.opts.audio {
			audioChan, audioPort, err = s.opts.sessions.allocate()
			if err != nil {
				s.Warn("allocate error",
					slog.String("err", err.Error()))
				audioPort = -1
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		client.onJoinEvent = func(c *connection, pack *jt1078.Packet) error {
			publisher, err := s.opts.pubFunc(ctx, pack)
			if err != nil {
				return err
			}
			c.publisher = publisher
			httpBody = map[string]any{
				"streamPath": c.publisher.StreamPath,
				"sim":        pack.Sim,
				"channel":    pack.LogicChannel,
				"audioPort":  audioPort,
			}
			s.onEvent(s.opts.onJoinURL, httpBody)
			return nil
		}
		client.onLeaveEvent = func() {
			if s.opts.audio {
				s.opts.sessions.recycle(audioPort)
			}
			if len(httpBody) > 0 {
				s.onEvent(s.opts.onLeaveURL, httpBody)
			}
			cancel()
		}
		go func() {
			if err := client.run(audioChan); err != nil {
				s.Warn("run error",
					slog.Any("http body", httpBody),
					slog.String("err", err.Error()))
			}
		}()
	}
}

func (s *Service) onEvent(url string, httpBody map[string]any) {
	client := resty.New()
	client.SetTimeout(5 * time.Second)
	_, _ = client.R().
		SetBody(httpBody).
		ForceContentType("application/json; charset=utf-8").
		Post(url)
}
