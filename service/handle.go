package service

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type (
	Handler interface {
		JT808Handler
		Eventer
	}

	JT808Handler interface {
		Parse(jtMsg *jt808.JTMessage) error               // 解析终端上传的body数据
		HasReply() bool                                   // 回复的消息 是否需要回复
		Protocol() consts.JT808CommandType                // 协议类型
		ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error) // 平台回复给终端的body部分数据
		ReplyProtocol() consts.JT808CommandType           // 回复的协议类型
	}

	Eventer interface {
		OnReadExecutionEvent(message *Message) // 读到完整的jt808数据时
		OnWriteExecutionEvent(message Message) // 写入数据给终端后
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
