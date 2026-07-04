package adapter

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
)

type packageParse struct {
	frameReader *jt808.FrameReader
}

func newPackageParse() *packageParse {
	return &packageParse{
		frameReader: jt808.NewFrameReader(),
	}
}

func (p *packageParse) clear() {
	p.frameReader.Clear()
}

func (p *packageParse) unpack(data []byte) (msgs []*message, err error) {
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

func (p *packageParse) decodeFrame(originalData []byte, msgs []*message) ([]*message, error) {
	jtMsg := jt808.NewJTMessage()
	if err := jtMsg.Decode(originalData); err != nil {
		return msgs, fmt.Errorf("%w [%x]", err, originalData)
	}
	return append(msgs, &message{
		JTMessage:    jtMsg,
		originalData: originalData,
	}), nil
}
