package main

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"simulator/internal/mq"
	"simulator/internal/shared"
)

type T0x0704 struct {
	model.T0x0704
}

func (t *T0x0704) OnReadExecutionEvent(message *service.Message) {
	var t0x0704 model.T0x0704
	if err := t0x0704.Parse(message.JTMessage); err != nil {
		return
	}

	locations := make([]*shared.Location, 0, len(t0x0704.Items))
	for _, item := range t0x0704.Items {
		location := shared.NewLocation(message.Header.TerminalPhoneNo, item.Latitude, item.Longitude)
		locations = append(locations, location)
	}
	batch := shared.NewLocationBatch(locations...)
	if err := mq.Default().Pub(shared.SubLocationBatch, batch.Encode()); err != nil {
		return
	}
}

func (t *T0x0704) OnWriteExecutionEvent(_ service.Message) {}
