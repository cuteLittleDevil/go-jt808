package service

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"log/slog"
	"net"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

// GoJT808 是 JT808 服务的核心入口对象。
type GoJT808 struct {
	opts *Options
	*sessionManager
}

// New 创建 JT808 服务实例并初始化会话管理器.
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

// Run 启动 TCP 服务并持续接收终端连接.
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
		conn, err := in.AcceptTCP()
		if err != nil {
			slog.Warn("accept fail",
				slog.Any("err", err))
			continue
		}
		client := newConnection(connectionParams{
			conn:                   conn,
			handles:                g.createCommandHandle(),
			activeRespondHandles:   g.createActiveRespondHandle(),
			terminalEvent:          g.opts.CustomTerminalEventerFunc(),
			filter:                 g.opts.FilterSubcontract,
			onTerminalTimeoutEvent: g.opts.OnTerminalTimeoutEvent,
			timeout: TerminalTimeout{
				ConnectionStartTime: time.Now(),
				Address:             conn.RemoteAddr().String(),
				IdleTimeout:         g.opts.IdleTimeout,
			},
			onJoinEvent:  g.sessionManager.join,
			onLeaveEvent: g.sessionManager.leave,
		})
		go client.run()
	}
}

// SendActiveMessage 将平台主动消息（下行指令）路由到对应的终端会话，并等待终端应答结果。
//
// 路由策略与流程（按顺序执行）：
//  1. 根据 activeMsg.Key 在会话管理器中查找对应终端的在线会话
//  2. 若找到在线会话：
//     - 注入会话 Header（包含手机号、终端ID等信息）
//     - 设置 replyChan 用于接收应答结果
//     - 将主动消息转发到该终端的 activeMsgChan 下行通道
//  3. 若未找到对应会话：立即返回 ErrNotExistKey 错误
//
// 返回值 *Message：
//   - 成功时：ExtensionFields.Err == nil，且包含终端的应答数据（例如 0x0001 通用应答）
//   - 失败时：ExtensionFields.Err 携带具体失败原因，例如：
//   - ErrNotExistKey：终端不在线或 Key 不存在
//   - ErrWriteDataOverTime：下发超时
//   - 其他网络错误
func (g *GoJT808) SendActiveMessage(activeMsg *ActiveMessage) *Message {
	return g.sessionManager.write(activeMsg)
}

func (g *GoJT808) createCommandHandle() map[consts.JT808CommandType]Handler {
	handles := g.createDefaultHandle()
	customHandles := g.opts.CustomHandleFunc()
	for k, v := range customHandles {
		handles[k] = v
	}
	return handles
}

func (g *GoJT808) createActiveRespondHandle() map[consts.JT808CommandType]func(
	platformMsg *ActiveMessage, terminalMsg *Message) bool {
	handles := g.createDefaultRespondHandle()
	customHandles := g.opts.CustomActiveRespondHandlerFunc()
	for k, v := range customHandles {
		handles[k] = v
	}
	return handles
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
		consts.P9105AudioVideoControlStatusNotice:     newDefaultHandle(&model.P0x9105{}),
		consts.P9201SendVideoRecordRequest:            newDefaultHandle(&model.P0x9201{}),
		consts.P9202SendVideoRecordControl:            newDefaultHandle(&model.P0x9202{}),
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
		consts.T1211FileInfoUpload:            newDefaultHandle(&model.T0x1211{}),
		consts.T1212FileUploadComplete:        newDefaultHandle(&model.T0x1212{}),
		consts.P9212FileUploadCompleteRespond: newDefaultHandle(&model.P0x9212{}),
	}
}

// HasMatch 判断当前应答是否匹配某条平台下发的主动指令.
func (g *GoJT808) createDefaultRespondHandle() map[consts.JT808CommandType]func(
	platformMsg *ActiveMessage, terminalMsg *Message) bool {
	return map[consts.JT808CommandType]func(*ActiveMessage, *Message) bool{
		// 平台-查询终端音视频属性 (只需要匹配指令)
		consts.T1003UploadAudioVideoAttr: func(activeMsg *ActiveMessage, _ *Message) bool {
			return activeMsg.Command == consts.P9003QueryTerminalAudioVideoProperties
		},
		// 通用应答，有部分指令先忽略，如8801 -> 0001(跳过) -> 8805
		consts.T0001GeneralRespond: g.makeSerialMatchHandler(func(jtMsg *jt808.JTMessage) (uint16, error) {
			var tmp model.T0x0001
			if err := tmp.Parse(jtMsg); err != nil {
				return 0, err
			}
			return tmp.SerialNumber, nil
		}),
		// 以下情况匹配下发指令的序列号
		consts.T0104QueryParameter: g.makeSerialMatchHandler(func(jtMsg *jt808.JTMessage) (uint16, error) {
			var tmp model.T0x0104
			if err := tmp.Parse(jtMsg); err != nil {
				return 0, err
			}
			return tmp.RespondSerialNumber, nil
		}),
		consts.T0201QueryLocation: g.makeSerialMatchHandler(func(jtMsg *jt808.JTMessage) (uint16, error) {
			var tmp model.T0x0201
			if err := tmp.Parse(jtMsg); err != nil {
				return 0, err
			}
			return tmp.RespondSerialNumber, nil
		}),
		consts.T0302QuestionAnswer: g.makeSerialMatchHandler(func(jtMsg *jt808.JTMessage) (uint16, error) {
			var tmp model.T0x0302
			if err := tmp.Parse(jtMsg); err != nil {
				return 0, err
			}
			return tmp.SerialNumber, nil
		}),
		consts.T1205UploadAudioVideoResourceList: g.makeSerialMatchHandler(func(jtMsg *jt808.JTMessage) (uint16, error) {
			var tmp model.T0x1205
			if err := tmp.Parse(jtMsg); err != nil {
				return 0, err
			}
			return tmp.SerialNumber, nil
		}),
		consts.T1206FileUploadCompleteNotice: g.makeSerialMatchHandler(func(jtMsg *jt808.JTMessage) (uint16, error) {
			var tmp model.T0x1206
			if err := tmp.Parse(jtMsg); err != nil {
				return 0, err
			}
			return tmp.RespondSerialNumber, nil
		}),
		consts.T0805CameraShootImmediately: g.makeSerialMatchHandler(func(jtMsg *jt808.JTMessage) (uint16, error) {
			var tmp model.T0x0805
			if err := tmp.Parse(jtMsg); err != nil {
				return 0, err
			}
			return tmp.RespondSerialNumber, nil
		}),
	}
}

// makeSerialMatchHandler 匹配流水号.
func (g *GoJT808) makeSerialMatchHandler(parseSerial func(*jt808.JTMessage) (uint16, error)) func(*ActiveMessage, *Message) bool {
	return func(platformMsg *ActiveMessage, terminalMsg *Message) bool {
		if terminalMsg.Command == consts.T0001GeneralRespond && isWaitingCommand(platformMsg.Command) {
			return false
		}

		serial, err := parseSerial(terminalMsg.JTMessage)
		if err != nil {
			return false
		}
		return platformMsg.ExtensionFields.PlatformSeq == serial
	}
}

func isWaitingCommand(cmd consts.JT808CommandType) bool {
	// 如果是这些命令的话 等待后续应答 如 8801 -> 8805
	switch cmd {
	case consts.P8801CameraShootImmediateCommand,
		consts.P9003QueryTerminalAudioVideoProperties,
		consts.P9205QueryResourceList,
		consts.P9206FileUploadInstructions:
		return true
	}
	return false
}
