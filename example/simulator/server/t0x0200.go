package main

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"simulator/internal/mq"
	"simulator/internal/shared"
)

type T0x0200 struct {
	model.T0x0200
}

func (t *T0x0200) OnReadExecutionEvent(message *service.Message) {
	var t0x0200 model.T0x0200
	if err := t0x0200.Parse(message.JTMessage); err != nil {
		return
	}
	location := shared.NewLocation(message.Header.TerminalPhoneNo, t0x0200.Latitude, t0x0200.Longitude)
	if err := mq.Default().Pub(shared.SubLocation, location.Encode()); err != nil {
		return
	}
}

func (t *T0x0200) OnWriteExecutionEvent(_ service.Message) {}
