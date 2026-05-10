package main

import (
	"encoding/binary"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type CustomTerminalReply struct {
	model.BaseHandle
	RespondSerialNumber uint16
}

func (m *CustomTerminalReply) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 2 {
		return protocol.ErrBodyLengthInconsistency
	}
	m.RespondSerialNumber = binary.BigEndian.Uint16(body[0:2])
	return nil
}

func (m *CustomTerminalReply) Encode() []byte {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data[0:2], m.RespondSerialNumber)
	return data
}

func (m *CustomTerminalReply) Protocol() consts.JT808CommandType {
	return consts.JT808CommandType(0x6661)
}

func (m *CustomTerminalReply) OnReadExecutionEvent(msg *service.Message) {}

func (m *CustomTerminalReply) OnWriteExecutionEvent(msg service.Message) {}

type CustomTerminalRequest struct {
	model.T0x0001
}

func (c *CustomTerminalRequest) OnReadExecutionEvent(msg *service.Message) {
}

func (c *CustomTerminalRequest) OnWriteExecutionEvent(msg service.Message) {
}
