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
	options := newOptions(opts)
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
		conn := newConnection(c, handles, terminalEvent, g.opts.FilterSubcontract,
			g.sessionManager.join, g.sessionManager.leave)
		go conn.Start()
	}
}

func (g *GoJT808) SendActiveMessage(activeMsg *ActiveMessage) *Message {
	return g.sessionManager.write(activeMsg)
}

func (g *GoJT808) createDefaultHandle() map[consts.JT808CommandType]Handler {
	return map[consts.JT808CommandType]Handler{
		// 终端上传的
		consts.T0001GeneralRespond:            newDefaultHandle(&model.T0x0001{}),
		consts.T0100Register:                  newDefaultHandle(&model.T0x0100{}),
		consts.T0102RegisterAuth:              newDefaultHandle(&model.T0x0102{}),
		consts.T0002HeartBeat:                 newDefaultHandle(&model.T0x0002{}),
		consts.T0200LocationReport:            newDefaultHandle(&model.T0x0200{}),
		consts.T0201QueryLocation:             newDefaultHandle(&model.T0x0201{}),
		consts.T0302QuestionAnswer:            newDefaultHandle(&model.T0x0302{}),
		consts.T0704LocationBatchUpload:       newDefaultHandle(&model.T0x0704{}),
		consts.T0104QueryParameter:            newDefaultHandle(&model.T0x0104{}),
		consts.T0805CameraShootImmediately:    newDefaultHandle(&model.T0x0805{}),
		consts.T0800MultimediaEventInfoUpload: newDefaultHandle(&model.T0x0800{}),
		consts.T0801MultimediaDataUpload:      newDefaultHandle(&model.T0x0801{}),

		// 平台下发的
		consts.P8003ReissueSubcontractingRequest: newDefaultHandle(&model.P0x8003{}),
		consts.P8103SetTerminalParams:            newDefaultHandle(&model.P0x8103{}),
		consts.P8104QueryTerminalParams:          newDefaultHandle(&model.P0x8104{}),
		consts.P8201QueryLocation:                newDefaultHandle(&model.P0x8201{}),
		consts.P8202TmpLocationTrack:             newDefaultHandle(&model.P0x8202{}),
		consts.P8300TextInfoDistribution:         newDefaultHandle(&model.P0x8300{}),
		consts.P8302QuestionDistribution:         newDefaultHandle(&model.P0x8302{}),
		consts.P8801CameraShootImmediateCommand:  newDefaultHandle(&model.P0x8801{}),

		// JT1078相关的
		consts.P9003QueryTerminalAudioVideoProperties: newDefaultHandle(&model.P0x9003{}),
		consts.T1003UploadAudioVideoAttr:              newDefaultHandle(&model.T0x1003{}),
		consts.T1005UploadPassengerFlow:               newDefaultHandle(&model.T0x1005{}),
		consts.P9101RealTimeAudioVideoRequest:         newDefaultHandle(&model.P0x9101{}),
		consts.P9102AudioVideoControl:                 newDefaultHandle(&model.P0x9102{}),
		consts.P9205QueryResourceList:                 newDefaultHandle(&model.P0x9205{}),
		consts.T1205UploadAudioVideoResourceList:      newDefaultHandle(&model.T0x1205{}),
		consts.P9206FileUploadInstructions:            newDefaultHandle(&model.P0x9206{}),
		consts.T1206FileUploadCompleteNotice:          newDefaultHandle(&model.T0x1206{}),
		consts.P9207FileUploadControl:                 newDefaultHandle(&model.P0x9207{}),

		// 主动安全的 默认苏标
		consts.P9208AlarmAttachUpload: newDefaultHandle(&model.P0x9208{
			P9208AlarmSign: model.P9208AlarmSign{
				ActiveSafetyType: consts.ActiveSafetyJS,
			},
		}),
		consts.T1210AlarmAttachInfoMessage: newDefaultHandle(&model.T0x1210{
			P9208AlarmSign: model.P9208AlarmSign{
				ActiveSafetyType: consts.ActiveSafetyJS,
			},
		}),
		consts.T1211FileInfoUpload:     newDefaultHandle(&model.T0x1211{}),
		consts.T1212FileUploadComplete: newDefaultHandle(&model.T0x1212{}),
	}
}
