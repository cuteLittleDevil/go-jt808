package terminal

import (
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type Option struct {
	F func(o *Options)
}

type Options struct {
	Header                   *jt808.Header
	CustomProtocolHandleFunc func() map[consts.JT808CommandType]Handler
}

func (o *Options) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}

func NewOptions(opts []Option) *Options {
	options := &Options{}
	options.Apply(opts)
	return options
}

func WithCustomHeader(header *jt808.Header) Option {
	return Option{F: func(o *Options) {
		o.Header = header
	}}
}

func WithHeader(protocolVersion consts.ProtocolVersionType, phone string) Option {
	return Option{F: func(o *Options) {
		msg := "7e0002000001234567890100008a7e"
		phone = fmt.Sprintf("%012s", phone)
		if protocolVersion == consts.JT808Protocol2019 {
			msg = "7e0002400001000000000172998417380000027e"
			phone = fmt.Sprintf("%020s", phone)
		}
		var jtMsg *jt808.JTMessage
		jtMsg = jt808.NewJTMessage()
		data, _ := hex.DecodeString(msg)
		_ = jtMsg.Decode(data)
		jtMsg.Header.TerminalPhoneNo = phone // 终端手机号
		jtMsg.Header.SerialNumber = 0        // 流水号
		jtMsg.Header.ProtocolVersion = protocolVersion
		o.Header = jtMsg.Header
	}}
}

func WithCustomProtocolHandleFunc(customFunc func() map[consts.JT808CommandType]Handler) Option {
	return Option{F: func(o *Options) {
		o.CustomProtocolHandleFunc = customFunc
	}}
}
