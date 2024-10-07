package service

import (
	"bytes"
	"fmt"
	"log/slog"
)

type packageParse struct {
	historyData []byte
	// 1. 暂不支持补传分包
	// 2. 必须全部完成才行 有没有收到的部分话 下一次这个id的包发送时丢弃
	subcontractingRecord map[uint16][][]byte
}

func newPackageParse() *packageParse {
	return &packageParse{
		subcontractingRecord: make(map[uint16][][]byte),
		historyData:          make([]byte, 0),
	}
}

func (p *packageParse) clear() {
	p.historyData = nil
	for id, datas := range p.subcontractingRecord {
		slog.Warn("package no complete",
			slog.Any("id", id),
			slog.Int("data sum", len(datas)))
	}
	p.subcontractingRecord = nil
}

// parse 返回一个或者多个完成的包
func (p *packageParse) parse(data []byte) ([]*Message, error) {
	msgs, err := p.unpack(data)
	if len(msgs) > 0 {
		return p.completePack(msgs), err
	}
	return nil, err
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
			msg := NewMessage(data)
			if err := msg.JTMessage.Decode(data); err != nil {
				return nil, fmt.Errorf("%w [%x]", err, data)
			}
			return []*Message{msg}, nil
		}
	}
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
		msg := NewMessage(originalData)
		if err := msg.JTMessage.Decode(originalData); err != nil {
			return nil, fmt.Errorf("%w [%x]", err, originalData)
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

func (p *packageParse) completePack(msgs []*Message) []*Message {
	completeMsgs := make([]*Message, 0, len(msgs))
	for _, msg := range msgs {
		header := msg.JTMessage.Header
		if sum := int(header.SubPackageSum); sum > 0 {
			id := header.ID
			seq := int(header.SubPackageNo)
			if seq == 1 {
				// 第一个包 如果有老id未完成的 就覆盖了
				if _, ok := p.subcontractingRecord[id]; ok {
					slog.Warn("not complete package",
						slog.String("phone", header.TerminalPhoneNo),
						slog.Any("id", id))
				}
				p.subcontractingRecord[id] = make([][]byte, header.SubPackageSum)
			}

			if seq > len(p.subcontractingRecord[id]) {
				slog.Warn("abnormal packet length",
					slog.Int("seq", seq),
					slog.Int("record sum", len(p.subcontractingRecord[id])),
					slog.String("phone", header.TerminalPhoneNo),
					slog.Any("id", id))
				continue
			}

			p.subcontractingRecord[id][seq-1] = msg.JTMessage.Body
			receivedSum := 0
			for _, data := range p.subcontractingRecord[id] {
				if len(data) != 0 {
					receivedSum++
				}
			}
			// 接收的和记录的一样 说明完成了
			if receivedSum == sum {
				data := make([]byte, 0, sum*1023)
				for i := 0; i < sum; i++ {
					data = append(data, p.subcontractingRecord[id][i]...)
				}
				msg.Body = data
				completeMsgs = append(completeMsgs, msg)
				delete(p.subcontractingRecord, id)
			}
			continue
		}
		completeMsgs = append(completeMsgs, msg)
	}
	return completeMsgs
}
