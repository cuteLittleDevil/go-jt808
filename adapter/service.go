package adapter

import (
	"errors"
	"log/slog"
	"net"
	"sync"
)

type Adapter struct {
	opts     *Options
	ln       net.Listener
	stopCh   chan struct{}
	stopOnce sync.Once
	lnOnce   sync.Once
}

func New(opts ...Option) *Adapter {
	options := newOptions(opts)
	return &Adapter{
		opts:   options,
		stopCh: make(chan struct{}),
	}
}

// Run 启动适配器并阻塞接受真实终端连接，直到 Stop 或监听失败。
func (a *Adapter) Run() {
	addr, err := net.ResolveTCPAddr("tcp", a.opts.Addr)
	if err != nil {
		slog.Error("resolve tcp addr error",
			slog.String("addr", a.opts.Addr),
			slog.Any("err", err))
		return
	}

	in, err := net.ListenTCP("tcp", addr)
	if err != nil {
		slog.Error("tcp listen fail",
			slog.Any("addr", addr),
			slog.Any("err", err))
		return
	}
	a.lnOnce.Do(func() { a.ln = in })

	go func() {
		<-a.stopCh
		_ = in.Close()
	}()

	for {
		c, err := in.AcceptTCP()
		if err != nil {
			select {
			case <-a.stopCh:
				return
			default:
			}
			// 监听被关闭时 Accept 也会失败
			if errors.Is(err, net.ErrClosed) {
				return
			}
			slog.Warn("accept fail",
				slog.Any("err", err))
			continue
		}

		s := newSession(c, a.opts.TimeoutRetry, a.createTerminals())
		go s.run()
	}
}

// Stop 关闭监听，使 Run 返回；已建立的 session 随真实连接断开或各自 stop 结束。
func (a *Adapter) Stop() {
	a.stopOnce.Do(func() {
		close(a.stopCh)
		if a.ln != nil {
			_ = a.ln.Close()
		}
	})
}

func (a *Adapter) createTerminals() []Terminal {
	ts := make([]Terminal, 0, len(a.opts.Terminals))
	for _, t := range a.opts.Terminals {
		for _, command := range a.opts.AllowCommand {
			t.AllowCommands = append(t.AllowCommands, command)
		}
		ts = append(ts, t)
	}
	return ts
}
