package adapter

import (
	"log/slog"
	"net"
)

type Adapter struct {
	opts *Options
}

func New(opts ...Option) *Adapter {
	options := newOptions(opts)
	g := &Adapter{
		opts: options,
	}
	return g
}

func (a *Adapter) Run() {
	addr, err := net.ResolveTCPAddr("tcp", a.opts.Addr)
	if err != nil {
		slog.Error("resolve tcp addr error",
			slog.String("addr", a.opts.Addr),
			slog.Any("err", err))
	}

	in, err := net.ListenTCP("tcp", addr)
	if err != nil {
		slog.Error("tcp listen fail",
			slog.Any("addr", addr),
			slog.Any("err", err))
		return
	}

	for {
		c, err := in.AcceptTCP()
		if err != nil {
			slog.Warn("accept fail",
				slog.Any("err", err))
			continue
		}

		g := newGroup(c, a.opts.TimeoutRetry, a.createTerminals())
		go g.run()
	}
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
