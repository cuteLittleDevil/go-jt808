package attachment

import (
	"log/slog"
	"net"
)

type GoJT808 struct {
	opts *Options
}

func New(opts ...Option) *GoJT808 {
	options := newOptions(opts)
	g := &GoJT808{
		opts: options,
	}
	return g
}

func (g *GoJT808) Run() {
	in, err := net.Listen(g.opts.Network, g.opts.Addr)
	if err != nil {
		slog.Error("listen fail",
			slog.Any("addr", g.opts.Addr),
			slog.Any("err", err))
		return
	}

	for {
		c, err := in.Accept()
		if err != nil {
			slog.Warn("accept fail",
				slog.Any("err", err))
			continue
		}
		conn := newConnection(c, g.opts.StreamDataHandleFunc,
			g.opts.JT808DataHandleFunc(), g.opts.FileEventerFunc())
		go conn.run()
	}
}
