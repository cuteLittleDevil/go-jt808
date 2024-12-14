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
		OnReadExecutionEvent(msg *Message) // 读到jt808数据时
		OnWriteExecutionEvent(msg Message) // 写入数据给终端后
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
	if err != nil {
		d.createTime = time.Now()
		slog.Debug("join",
			slog.String("key", key),
			slog.String("create time", d.createTime.Format(time.DateTime)),
			slog.String("data", fmt.Sprintf("%x", msg.ExtensionFields.TerminalData)))
	}
}

func (d *defaultTerminalEvent) OnLeaveEvent(key string) {
	slog.Debug("leave",
		slog.String("key", key),
		slog.String("create time", d.createTime.Format(time.DateTime)),
		slog.Float64("online time second", time.Since(d.createTime).Seconds()))
}

func (d *defaultTerminalEvent) OnNotSupportedEvent(msg *Message) {
	slog.Warn("key not found",
		slog.Any("seq", msg.ExtensionFields.TerminalSeq),
		slog.Any("id", msg.JTMessage.Header.ID),
		slog.String("remark", msg.Command.String()))
}

func (d *defaultTerminalEvent) OnReadExecutionEvent(_ *Message) {}

func (d *defaultTerminalEvent) OnWriteExecutionEvent(_ Message) {}
