package service

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
)

type Message struct {
	*jt808.JTMessage
	Handler
	OriginalData []byte `json:"-"`
	ReplyData    []byte `json:"-"`
	WriteErr     error  `json:"-"`
}

func NewMessage(originalData []byte) *Message {
	return &Message{
		JTMessage:    jt808.NewJTMessage(),
		OriginalData: originalData,
	}
}

func newErrMessage(err error) *Message {
	return &Message{WriteErr: err}
}
