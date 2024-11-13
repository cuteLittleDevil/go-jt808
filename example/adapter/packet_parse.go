package main

import (
	"bytes"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
)

type packageParse struct {
	historyData []byte
}

func newPackageParse() *packageParse {
	return &packageParse{
		historyData: make([]byte, 0),
	}
}

func (p *packageParse) unpack(data []byte) (msgs []*Message, err error) {
	const sign = 0x7e
	if len(p.historyData) == 0 && len(data) > 2 && data[len(data)-1] == sign {
		// 快速路径 直接完成 从头到尾只有两个7e的情况
		count := 0
		index := bytes.IndexFunc(data, func(r rune) bool {
			if r == sign {
				count++
			}
			return count == 2
		})
		if index == len(data)-1 {
			jtMsg := jt808.NewJTMessage()
			if err := jtMsg.Decode(data); err != nil {
				return nil, fmt.Errorf("%w [%x]", err, data)
			}
			msg := &Message{
				JTMessage:    jtMsg,
				OriginalData: data,
			}
			return []*Message{msg}, nil
		}
	}
	p.historyData = append(p.historyData, data...)
	for {
		end := -1
		if len(p.historyData) > 2 && p.historyData[0] == sign {
			for i := 1; i < len(p.historyData); i++ {
				if p.historyData[i] == sign {
					end = i + 1
					break
				}
			}
		}
		if end == -1 {
			break
		}
		originalData := p.historyData[:end]
		jtMsg := jt808.NewJTMessage()
		if err := jtMsg.Decode(originalData); err != nil {
			p.historyData = p.historyData[end:]
			return msgs, fmt.Errorf("%w [%x]", err, originalData)
		}
		msg := &Message{
			JTMessage:    jtMsg,
			OriginalData: data,
		}
		msgs = append(msgs, msg)
		if end == len(p.historyData) {
			// 没有遗留的数据
			p.historyData = p.historyData[0:0]
			return msgs, nil
		}
		p.historyData = p.historyData[end:]
	}
	return msgs, nil
}
