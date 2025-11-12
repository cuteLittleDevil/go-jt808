package service

import (
	"bytes"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"log/slog"
	"time"
)

type (
	packageParse struct {
		historyData          []byte
		subcontractingRecord map[uint16][][]byte
		timeoutRecord        map[uint16]*packageComplete
	}

	packageComplete struct {
		createTime time.Time
		updateTime time.Time
		initHeader *jt808.Header
	}
)

func newPackageParse() *packageParse {
	return &packageParse{
		subcontractingRecord: make(map[uint16][][]byte),
		timeoutRecord:        make(map[uint16]*packageComplete),
		historyData:          make([]byte, 0),
	}
}

func (p *packageParse) clear() {
	clear(p.historyData)
	for id, datas := range p.subcontractingRecord {
		slog.Warn("package no complete",
			slog.Any("id", id),
			slog.Int("data sum", len(datas)))
	}
	clear(p.subcontractingRecord)
	clear(p.timeoutRecord)
}

// parse 返回一个或者多个完成的包.
func (p *packageParse) parse(data []byte) ([]*Message, error) {
	msgs, err := p.unpack(data)
	for _, msg := range msgs {
		if completeMsg, ok := p.completePack(msg); ok {
			msgs = append(msgs, completeMsg)
		}
	}
	if len(p.timeoutRecord) > 0 {
		p.deleteTimeoutPackage()
		if v, ok := p.supplementarySubPackage(); ok {
			msgs = append(msgs, v...)
		}
	}
	return msgs, err
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
			msg := newTerminalMessage(jtMsg, data)
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
		msg := newTerminalMessage(jtMsg, originalData)
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

func (p *packageParse) completePack(msg *Message) (*Message, bool) {
	header := msg.JTMessage.Header
	if sum := int(header.SubPackageSum); sum > 0 {
		id := header.ID
		seq := int(header.SubPackageNo)
		if seq == 1 {
			// 第一个包 如果有老id未完成的 就覆盖了
			if _, ok := p.subcontractingRecord[id]; ok {
				slog.Warn("not complete package",
					slog.String("phone", header.TerminalPhoneNo),
					slog.Int("seq", seq),
					slog.Any("id", id))
			}
			p.add(id, header)
		}

		if seq > len(p.subcontractingRecord[id]) || seq <= 0 {
			slog.Warn("completePack",
				slog.Int("seq", seq),
				slog.Int("record sum", len(p.subcontractingRecord[id])),
				slog.String("phone", header.TerminalPhoneNo),
				slog.Any("id", id))
			return nil, false
		}

		// 分包的情况 每一次都确保是新的
		p.subcontractingRecord[id][seq-1] = make([]byte, len(msg.JTMessage.Body))
		copy(p.subcontractingRecord[id][seq-1], msg.JTMessage.Body)

		p.timeoutRecord[id].updateTime = time.Now()
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
			p.remove(id)
			completeMsg := newTerminalMessage(msg.JTMessage, data)
			completeMsg.Body = data
			completeMsg.ExtensionFields.SubcontractComplete = true
			return completeMsg, true
		}
	}
	return nil, false
}

func (p *packageParse) add(id uint16, header *jt808.Header) {
	p.subcontractingRecord[id] = make([][]byte, header.SubPackageSum)
	now := time.Now()
	p.timeoutRecord[id] = &packageComplete{
		createTime: now,
		updateTime: now,
		initHeader: header,
	}
}

func (p *packageParse) remove(id uint16) {
	delete(p.subcontractingRecord, id)
	delete(p.timeoutRecord, id)
}

func (p *packageParse) deleteTimeoutPackage() {
	now := time.Now().Add(-60 * time.Second)
	for k, v := range p.timeoutRecord {
		if now.After(v.createTime) { // x秒内还没有完成的 就删除了
			p.remove(k)
			slog.Warn("timeout",
				slog.Any("id", k),
				slog.String("remove", v.initHeader.String()))
		}
	}
}

func (p *packageParse) supplementarySubPackage() ([]*Message, bool) {
	msgs := make([]*Message, 0)
	now := time.Now().Add(-5 * time.Second)
	for id, v := range p.timeoutRecord {
		if now.After(v.updateTime) {
			seqs := make([]uint16, 0, v.initHeader.SubPackageSum)
			for k, record := range p.subcontractingRecord[id] {
				if len(record) == 0 {
					seqs = append(seqs, uint16(k+1))
				}
			}
			p0x8003 := model.P0x8003{
				BaseHandle:           model.BaseHandle{},
				OriginalSerialNumber: v.initHeader.SerialNumber,
				AgainPackageCount:    byte(len(seqs)),
				AgainPackageList:     seqs,
			}
			v.initHeader.ReplyID = uint16(p0x8003.Protocol())
			v.initHeader.Property.PacketFragmented = 0
			data := v.initHeader.Encode(p0x8003.Encode())
			jtMsg := jt808.NewJTMessage()
			_ = jtMsg.Decode(data)
			subMsg := newTerminalMessage(jtMsg, data)
			msgs = append(msgs, subMsg)
			v.updateTime = time.Now()
		}
	}
	return msgs, len(msgs) != 0
}
