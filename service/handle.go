package service

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"time"
)

type (
	Handler interface {
		JT808Handler
		Eventer
	}

	JT808Handler interface {
		Parse(jtMsg *jt808.JTMessage) error               // 解析终端上传的body数据
		Protocol() consts.JT808CommandType                // 协议类型
		HasReply() bool                                   // 终端主动上传的情况 自动回复的消息 是否需要回复
		ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error) // 终端主动上传的情况 平台回复给终端的body部分数据
		ReplyProtocol() consts.JT808CommandType           // 终端主动上传的情况 回复的协议类型
	}

	TerminalEventer interface {
		OnJoinEvent(msg *Message, key string, err error) // 终端加入事件
		OnLeaveEvent(key string)                         // 终端离开事件
		OnNotSupportedEvent(msg *Message)                // 读取终端报文 发现未实现的报文处理
		Eventer
	}

	Eventer interface {
		// OnReadExecutionEvent 读事件 以下两种情况触发
		// 1. 终端上传的报文到平台完成时 如0x0200位置信息
		// 2. 平台主动下发报文到终端后 如0x9101实时视频
		// 注意这里面的msg.JTMessage.Body在指令的情况[]byte是复用的（因此需要使用请自己深拷贝)
		// 仅在分包完成的总包时是新[]byte
		OnReadExecutionEvent(msg *Message)
		// OnWriteExecutionEvent 写事件 以下情况触发
		// 1. 回复终端主动上传的报文 如0x0200 -> 0x8001 收到设备位置信息 平台回复通用应答
		// 2. 收到终端应答 如0x9001 -> 0x0001 平台主动发送实时视频请求 终端回复通用应答
		// 3. 超时（平台主动下发指令的情况)
		OnWriteExecutionEvent(msg Message)
	}
)

type defaultHandle struct {
	JT808Handler
}

func newDefaultHandle(JT808Handler JT808Handler) *defaultHandle {
	return &defaultHandle{JT808Handler: JT808Handler}
}

func (d *defaultHandle) OnReadExecutionEvent(_ *Message) {}
func (d *defaultHandle) OnWriteExecutionEvent(_ Message) {
	//fmt.Println(fmt.Sprintf("read %x", message.OriginalData))
	//fmt.Println(fmt.Sprintf("write %x", message.ReplyData))
}

type defaultTerminalEvent struct {
	createTime time.Time
}

func (d *defaultTerminalEvent) OnJoinEvent(msg *Message, key string, err error) {
	if err == nil {
		d.createTime = time.Now()
		slog.Debug("join",
			slog.String("key", key),
			slog.String("create time", d.createTime.Format(time.DateTime)),
			slog.String("data", fmt.Sprintf("%x", msg.ExtensionFields.TerminalData)))
	}
}

func (d *defaultTerminalEvent) OnLeaveEvent(key string) {
	if !d.createTime.IsZero() {
		slog.Debug("leave",
			slog.String("key", key),
			slog.String("create time", d.createTime.Format(time.DateTime)),
			slog.Float64("online time second", time.Since(d.createTime).Seconds()))
	}
}

func (d *defaultTerminalEvent) OnNotSupportedEvent(msg *Message) {
	slog.Warn("not supported",
		slog.Any("seq", msg.ExtensionFields.TerminalSeq),
		slog.Any("id", msg.JTMessage.Header.ID),
		slog.String("data", fmt.Sprintf("%x", msg.ExtensionFields.TerminalData)),
		slog.String("remark", msg.Command.String()))
}

func (d *defaultTerminalEvent) OnReadExecutionEvent(_ *Message) {}

func (d *defaultTerminalEvent) OnWriteExecutionEvent(_ Message) {}
