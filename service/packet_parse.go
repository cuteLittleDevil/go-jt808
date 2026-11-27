package service

import (
	"bytes"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
)

type (
	packageParse struct {
		frameReader *jt808.FrameReader
		records     map[uint16]*packageComplete
	}

	packageComplete struct {
		slots      [][]byte
		createTime time.Time
		updateTime time.Time
		initHeader *jt808.Header
	}
)

func newPackageParse() *packageParse {
	return &packageParse{
		frameReader: jt808.NewFrameReader(),
		records:     make(map[uint16]*packageComplete),
	}
}

func (p *packageParse) clear() {
	p.frameReader.Clear()
	for id, rec := range p.records {
		slog.Warn("package no complete",
			slog.Any("id", id),
			slog.Int("data sum", len(rec.slots)))
	}
	clear(p.records)
}

// parse 返回一个或者多个完成的包.
func (p *packageParse) parse(data []byte) ([]*Message, error) {
	msgs, err := p.unpack(data)
	for _, msg := range msgs {
		if completeMsg, ok := p.completePack(msg); ok {
			msgs = append(msgs, completeMsg)
		}
	}
	if len(p.records) > 0 {
		p.deleteTimeoutPackage()
		if v, ok := p.supplementarySubPackage(); ok {
			msgs = append(msgs, v...)
		}
	}
	return msgs, err
}

func (p *packageParse) unpack(data []byte) (msgs []*Message, err error) {
	if frame, ok := p.frameReader.FeedSingleComplete(data); ok {
		return p.decodeFrame(frame, msgs)
	}
	p.frameReader.Append(data)
	for {
		frame, ok := p.frameReader.PopFrame()
		if !ok {
			break
		}
		msgs, err = p.decodeFrame(frame, msgs)
		if err != nil {
			return msgs, err
		}
	}
	return msgs, nil
}

func (p *packageParse) decodeFrame(originalData []byte, msgs []*Message) ([]*Message, error) {
	jtMsg := jt808.NewJTMessage()
	if err := jtMsg.Decode(originalData); err != nil {
		return msgs, fmt.Errorf("%w [%x]", err, originalData)
	}
	msg := newTerminalMessage(jtMsg, originalData)
	return append(msgs, msg), nil
}

func (p *packageParse) completePack(msg *Message) (*Message, bool) {
	header := msg.JTMessage.Header
	sum := int(header.SubPackageSum)
	if sum <= 0 {
		return nil, false
	}
	id := header.ID
	seq := int(header.SubPackageNo)
	if seq == 1 {
		// 第一个包 如果有老id未完成的 就覆盖了
		if _, ok := p.records[id]; ok {
			slog.Warn("not complete package",
				slog.String("phone", header.TerminalPhoneNo),
				slog.Int("seq", seq),
				slog.Any("id", id))
		}
		p.add(id, header)
	}

	rec := p.records[id]
	var recordSum int
	if rec != nil {
		recordSum = len(rec.slots)
	}
	if rec == nil || seq <= 0 || seq > recordSum {
		slog.Warn("completePack",
			slog.Int("seq", seq),
			slog.Int("record sum", recordSum),
			slog.String("phone", header.TerminalPhoneNo),
			slog.Any("id", id))
		return nil, false
	}

	rec.slots[seq-1] = bytes.Clone(msg.JTMessage.Body)
	rec.updateTime = time.Now()
	for _, slot := range rec.slots {
		if len(slot) == 0 {
			return nil, false
		}
	}

	data := make([]byte, 0, sum*1023)
	for _, slot := range rec.slots {
		data = append(data, slot...)
	}
	delete(p.records, id)
	completeMsg := newTerminalMessage(msg.JTMessage, data)
	completeMsg.Body = data
	completeMsg.ExtensionFields.SubcontractComplete = true
	return completeMsg, true
}

func (p *packageParse) add(id uint16, header *jt808.Header) {
	now := time.Now()
	// 深拷贝一份 避免外部修改已分发消息的Header影响补传记录
	initHeader := *header
	property := *header.Property
	initHeader.Property = &property
	p.records[id] = &packageComplete{
		slots:      make([][]byte, header.SubPackageSum),
		createTime: now,
		updateTime: now,
		initHeader: &initHeader,
	}
}

func (p *packageParse) deleteTimeoutPackage() {
	now := time.Now().Add(-60 * time.Second)
	for k, v := range p.records {
		if now.After(v.createTime) { // 超时仍未完成的删除记录
			delete(p.records, k)
			slog.Warn("timeout",
				slog.Any("id", k),
				slog.String("remove", v.initHeader.String()))
		}
	}
}

func (p *packageParse) supplementarySubPackage() ([]*Message, bool) {
	msgs := make([]*Message, 0, len(p.records))
	now := time.Now().Add(-5 * time.Second)
	for _, v := range p.records {
		if !now.After(v.updateTime) {
			continue
		}
		seqs := make([]uint16, 0, v.initHeader.SubPackageSum)
		for k, slot := range v.slots {
			if len(slot) == 0 {
				seqs = append(seqs, uint16(k+1))
			}
		}
		// 0x8003 重传包总数为BYTE 单次最多请求255个 剩余的下一轮再请求
		if len(seqs) > math.MaxUint8 {
			seqs = seqs[:math.MaxUint8]
		}
		p0x8003 := model.P0x8003{
			OriginalSerialNumber: v.initHeader.SerialNumber,
			AgainPackageCount:    byte(len(seqs)),
			AgainPackageList:     seqs,
		}
		v.initHeader.ReplyID = uint16(p0x8003.Protocol())
		v.initHeader.Property.PacketFragmented = 0
		data := v.initHeader.Encode(p0x8003.Encode())
		jtMsg := jt808.NewJTMessage()
		_ = jtMsg.Decode(data)
		msgs = append(msgs, newTerminalMessage(jtMsg, data))
		v.updateTime = time.Now()
	}
	return msgs, len(msgs) != 0
}
