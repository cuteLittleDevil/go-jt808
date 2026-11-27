package adapter

import (
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type Adapter struct {
	opts *Options
}

func New(opts ...Option) *Adapter {
	return &Adapter{
		opts: newOptions(opts),
	}
}

// Run 启动适配器并持续接收真实终端连接。
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
	a.serve(in)
}

func (a *Adapter) serve(in net.Listener) {
	for {
		c, err := in.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			slog.Warn("accept fail",
				slog.Any("err", err))
			time.Sleep(100 * time.Millisecond) // 如文件描述符耗尽 稍等再重试 避免空转
			continue
		}

		s := newSession(c, a.opts, a.createTerminals())
		go s.run()
	}
}

func (a *Adapter) createTerminals() []Terminal {
	ts := make([]Terminal, 0, len(a.opts.Terminals))
	seen := make(map[string]int, len(a.opts.Terminals))
	for _, t := range a.opts.Terminals {
		if t.TargetAddr == "" {
			continue
		}
		cmds := make([]consts.JT808CommandType, 0, len(t.AllowCommands)+len(a.opts.AllowCommand))
		cmds = append(cmds, t.AllowCommands...)
		cmds = append(cmds, a.opts.AllowCommand...)
		t.AllowCommands = cmds
		if idx, ok := seen[t.TargetAddr]; ok {
			ts[idx] = t
			continue
		}
		seen[t.TargetAddr] = len(ts)
		ts = append(ts, t)
	}
	return ts
}
