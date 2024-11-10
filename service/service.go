package service

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"net"
)

type GoJT808 struct {
	opts *Options
	*sessionManager
}

func New(opts ...Option) *GoJT808 {
	options := NewOptions(opts)
	g := &GoJT808{
		opts: options,
	}
	keyFunc := g.opts.KeyFunc
	g.sessionManager = newSessionManager(keyFunc)
	go g.sessionManager.run()
	return g
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
		customHandles := g.opts.CustomHandleFunc()
		for k, v := range customHandles {
			handles[k] = v
		}
		terminalEvent := g.opts.CustomTerminalEventerFunc()
		conn := newConnection(c, handles, terminalEvent, g.opts.FilterSubPack,
			g.sessionManager.join, g.sessionManager.leave)
		go conn.Start()
	}
}

func (g *GoJT808) SendActiveMessage(activeMsg *ActiveMessage) *Message {
	return g.sessionManager.write(activeMsg)
}

func (g *GoJT808) createDefaultHandle() map[consts.JT808CommandType]Handler {
	return map[consts.JT808CommandType]Handler{
		consts.T0100Register:                          newDefaultHandle(&model.T0x0100{}),
		consts.T0102RegisterAuth:                      newDefaultHandle(&model.T0x0102{}),
		consts.T0002HeartBeat:                         newDefaultHandle(&model.T0x0002{}),
		consts.T0200LocationReport:                    newDefaultHandle(&model.T0x0200{}),
		consts.T0704LocationBatchUpload:               newDefaultHandle(&model.T0x0704{}),
		consts.T0104QueryParameter:                    newDefaultHandle(&model.T0x0104{}),
		consts.P8104QueryTerminalParams:               newDefaultHandle(&model.P0x8104{}),
		consts.P9003QueryTerminalAudioVideoProperties: newDefaultHandle(&model.P0x9003{}),
		consts.T1003UploadAudioVideoAttr:              newDefaultHandle(&model.T0x1003{}),
		consts.T1005UploadPassengerFlow:               newDefaultHandle(&model.T0x1005{}),
		consts.P9101RealTimeAudioVideoRequest:         newDefaultHandle(&model.P0x9101{}),
		consts.P9102AudioVideoControl:                 newDefaultHandle(&model.P0x9102{}),
		consts.P9105AudioVideoControlStatusNotice:     newDefaultHandle(&model.P0x9205{}),
		consts.T1205UploadAudioVideoResourceList:      newDefaultHandle(&model.T0x1205{}),
		consts.P9206FileUploadInstructions:            newDefaultHandle(&model.P0x9206{}),
		consts.T1206FileUploadCompleteNotice:          newDefaultHandle(&model.T0x1206{}),
		consts.P9207FileUploadControl:                 newDefaultHandle(&model.P0x9207{}),
		consts.T0001GeneralRespond:                    newDefaultHandle(&model.T0x0001{}),
		consts.P8003ReissueSubcontractingRequest:      newDefaultHandle(&model.P0x8003{}),
		consts.P8103SetTerminalParams:                 newDefaultHandle(&model.P0x8103{}),
	}
}
