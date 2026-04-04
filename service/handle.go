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
		// OnReadExecutionEvent 读事件触发时机：
		//   1. 终端主动上传报文并完成Parse后，如0x0200位置信息
		//   2. 平台主动下发指令给终端后.如0x9101实时视频
		//
		// 注意：
		//   msg.JTMessage.Body 大部分情况下是复用的（出于性能考虑）
		//   如果需要长期保存 Body 数据，必须自行进行深拷贝.
		//   仅分包组装完成后的总包，Body 才是独立的新切片.
		OnReadExecutionEvent(msg *Message)

		// OnWriteExecutionEvent 写事件触发时机：
		//   1. 平台回复终端主动上传的报文（如 0x0200 → 0x8001）
		//   2. 终端回复平台主动下发的指令（如 0x9001 → 0x0001）
		//   3. 平台主动下发指令后发生超时 默认3秒.
		OnWriteExecutionEvent(msg Message)
	}
)

// defaultHandle 默认的 JT808 处理器包装
// 用于将单纯的 JT808Handler 转换为同时满足 Handler 接口的类型.
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

// OnJoinEvent 默认实现：终端加入成功时记录时间并输出调试日志.
func (d *defaultTerminalEvent) OnJoinEvent(msg *Message, key string, err error) {
	if err == nil {
		d.createTime = time.Now()
		slog.Debug("join",
			slog.String("key", key),
			slog.String("create time", d.createTime.Format(time.DateTime)),
			slog.String("data", fmt.Sprintf("%x", msg.ExtensionFields.TerminalData)))
	}
}

// OnLeaveEvent 默认实现：终端离开时输出在线时长.
func (d *defaultTerminalEvent) OnLeaveEvent(key string) {
	if !d.createTime.IsZero() {
		slog.Debug("leave",
			slog.String("key", key),
			slog.String("create time", d.createTime.Format(time.DateTime)),
			slog.Float64("online time second", time.Since(d.createTime).Seconds()))
	}
}

// OnNotSupportedEvent 默认实现：以警告级别记录未支持的指令.
func (d *defaultTerminalEvent) OnNotSupportedEvent(msg *Message) {
	slog.Warn("not supported",
		slog.Any("seq", msg.ExtensionFields.TerminalSeq),
		slog.Any("id", msg.JTMessage.Header.ID),
		slog.String("data", fmt.Sprintf("%x", msg.ExtensionFields.TerminalData)),
		slog.String("remark", msg.Command.String()))
}

func (d *defaultTerminalEvent) OnReadExecutionEvent(_ *Message) {}

func (d *defaultTerminalEvent) OnWriteExecutionEvent(_ Message) {}
