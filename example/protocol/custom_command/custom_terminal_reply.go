package main

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type CustomTerminalReply struct {
	model.BaseHandle
	Handle *CustomPlatform
	Data   uint16
}

func (m *CustomTerminalReply) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	m.Data = binary.BigEndian.Uint16(body)
	return nil
}

func (m *CustomTerminalReply) Encode() []byte {
	data := make([]byte, 0, 1)
	data = binary.BigEndian.AppendUint16(data, m.Data)
	return data
}

func (m *CustomTerminalReply) Protocol() consts.JT808CommandType {
	return CustomTerminalCommand
}

func (m *CustomTerminalReply) OnReadExecutionEvent(msg *service.Message) {
	fmt.Println("------- terminal read event")
	//msg.Handler = m.Handle
	m.Handle.OnWriteExecutionEvent(*msg)
}

func (m *CustomTerminalReply) OnWriteExecutionEvent(msg service.Message) {
	fmt.Println("------- terminal write event")

}
