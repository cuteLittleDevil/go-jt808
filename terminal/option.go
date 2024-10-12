package terminal

import (
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
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
		body := "000200000123456789010000"
		phone = fmt.Sprintf("%012s", phone)
		body = strings.Replace(body, "012345678901", phone, 1)
		if protocolVersion == consts.JT808Protocol2019 {
			body = "000240000112345678901234567890000002"
			phone = fmt.Sprintf("%020s", phone)
			body = strings.Replace(body, "12345678901234567890", phone, 1)
		}
		bodyData, _ := hex.DecodeString(body)
		code := utils.CreateVerifyCode(bodyData)
		data := []byte{0x7e}
		data = append(data, bodyData...)
		data = append(data, code)
		data = append(data, 0x7e)
		var jtMsg *jt808.JTMessage
		jtMsg = jt808.NewJTMessage()
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