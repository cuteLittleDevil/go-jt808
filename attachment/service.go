package attachment

import (
	"errors"
	"log/slog"
	"net"
)

var netListen = net.Listen

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
	in, err := netListen(g.opts.Network, g.opts.Addr)
	if err != nil {
		slog.Error("listen fail",
			slog.Any("addr", g.opts.Addr),
			slog.Any("err", err))
		return
	}

	for {
		c, err := in.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			slog.Warn("accept fail",
				slog.Any("err", err))
			continue
		}
		conn := newConnection(c, g.opts.ActiveSafetyType, g.opts.DataHandleFunc, g.opts.FileEventerFunc(), g.opts.ReadTimeout)
		go conn.run()
	}
}
