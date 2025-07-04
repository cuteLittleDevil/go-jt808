package stream

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
)

type packageParse struct {
	historyData []byte
	record      map[jt1078.DataType][]byte
}

func newPackageParse() *packageParse {
	return &packageParse{
		historyData: make([]byte, 0, 1023),
		record:      make(map[jt1078.DataType][]byte),
	}
}

func (p *packageParse) clear() {
	clear(p.historyData)
	clear(p.record)
}

func (p *packageParse) parse(data []byte) func(func(packet *jt1078.Packet, err error) bool) {
	p.historyData = append(p.historyData, data...)
	return func(yield func(packet *jt1078.Packet, err error) bool) {
		for len(p.historyData) > 0 {
			packet := jt1078.NewPacket()
			if remainData, err := packet.Decode(p.historyData); err == nil {
				p.historyData = remainData
				complete := false
				switch packet.SubcontractType {
				case jt1078.SubcontractTypeAtomic:
					p.record[packet.DataType] = packet.Body
					complete = true
				case jt1078.SubcontractTypeFirst:
					p.record[packet.DataType] = nil
					p.record[packet.DataType] = packet.Body
				case jt1078.SubcontractTypeLast:
					p.record[packet.DataType] = append(p.record[packet.DataType], packet.Body...)
					complete = true
				case jt1078.SubcontractTypeMiddle:
					p.record[packet.DataType] = append(p.record[packet.DataType], packet.Body...)
				default:
					yield(nil, fmt.Errorf("unknown SubcontractType %s", packet.SubcontractType))
					return
				}
				if complete {
					packet.Body = p.record[packet.DataType]
					yield(packet, nil)
				}
			} else {
				yield(nil, err)
				return
			}
		}
	}
}
