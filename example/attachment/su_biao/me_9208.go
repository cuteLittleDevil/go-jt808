package main

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
)

type me0x9208 struct {
	model.P0x9208
}

func (m *me0x9208) OnReadExecutionEvent(_ *service.Message) {}

func (m *me0x9208) OnWriteExecutionEvent(_ service.Message) {}
