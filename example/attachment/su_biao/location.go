package main

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
)

type meLocation struct {
	model.T0x0200
	model.T0x0200AdditionExtension0x64
	model.T0x0200AdditionExtension0x65
	model.T0x0200AdditionExtension0x66
	model.T0x0200AdditionExtension0x67
	model.T0x0200AdditionExtension0x70
}

func (l *meLocation) Parse(jtMsg *jt808.JTMessage) error {
	l.T0x0200.CustomAdditionContentFunc = func(id uint8, content []byte) (model.AdditionContent, bool) {
		switch id {
		case 0x64:
			return l.T0x0200AdditionExtension0x64.Parse(id, content)
		case 0x65:
			return l.T0x0200AdditionExtension0x65.Parse(id, content)
		case 0x66:
			return l.T0x0200AdditionExtension0x66.Parse(id, content)
		case 0x67:
			return l.T0x0200AdditionExtension0x67.Parse(id, content)
		case 0x70:
			return l.T0x0200AdditionExtension0x70.Parse(id, content)
		}
		return model.AdditionContent{}, false
	}
	return l.T0x0200.Parse(jtMsg)
}
