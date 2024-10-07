package service

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"net"
)

type GoJT808 struct {
	opts *Options
}

func New(opts ...Option) *GoJT808 {
	options := NewOptions(opts)
	return &GoJT808{
		opts: options,
	}
}

func (g *GoJT808) Run() {
	addr, err := net.ResolveTCPAddr(g.opts.Network, g.opts.Addr)
	if err != nil {
		slog.Error("resolve tcp addr error",
			slog.String("addr", g.opts.Addr),
			slog.String("network", g.opts.Network),
			slog.Any("err", err))
	}

	in, err := net.ListenTCP(g.opts.Network, addr)
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
		handles := g.createDefaultHandle()
		if g.opts.CustomHandleFunc != nil {
			customHandles := g.opts.CustomHandleFunc()
			for k, v := range customHandles {
				handles[k] = v
			}
		}
		conn := newConnection(c, handles)
		go conn.Start()
	}
}

func (g *GoJT808) createDefaultHandle() map[consts.JT808CommandType]Handler {
	return map[consts.JT808CommandType]Handler{
		consts.T0100Register:            newDefaultHandle(&model.T0x0100{}),
		consts.T0102RegisterAuth:        newDefaultHandle(&model.T0x0102{}),
		consts.T0002HeartBeat:           newDefaultHandle(&model.T0x0002{}),
		consts.T0200LocationReport:      newDefaultHandle(&model.T0x0200{}),
		consts.T0704LocationBatchUpload: newDefaultHandle(&model.T0x0704{}),
	}
}
