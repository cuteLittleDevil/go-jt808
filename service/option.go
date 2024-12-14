package service

import (
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type Option struct {
	F func(o *Options)
}

const (
	defaultAddr              = "0.0.0.0:808" // 服务默认地址
	defaultNetwork           = "tcp"         // 服务默认网络协议
	defaultFilterSubcontract = true          // 是否过滤分包的情况
)

type Options struct {
	// Addr 服务地址 默认0.0.0.0:808.
	Addr string
	// Network 服务协议 默认tcp.
	Network string
	// FilterSubcontract 是否过滤分包的情况 默认true.
	FilterSubcontract bool
	// KeyFunc 用于获取终端唯一标识 默认手机号.
	KeyFunc func(message *Message) (string, bool)
	// CustomTerminalEventerFunc 自定义终端事件.
	CustomTerminalEventerFunc func() TerminalEventer
	// CustomHandleFunc 自定义消息处理.
	CustomHandleFunc func() map[consts.JT808CommandType]Handler
}

func newOptions(opts []Option) *Options {
	options := &Options{
		Addr:              defaultAddr,
		Network:           defaultNetwork,
		FilterSubcontract: defaultFilterSubcontract,
		KeyFunc: func(message *Message) (string, bool) {
			return message.JTMessage.Header.TerminalPhoneNo, true
		},
		CustomTerminalEventerFunc: func() TerminalEventer {
			return &defaultTerminalEvent{}
		},
		CustomHandleFunc: func() map[consts.JT808CommandType]Handler {
			return map[consts.JT808CommandType]Handler{}
		},
	}
	for _, op := range opts {
		op.F(options)
	}
	return options
}

// WithHostPorts 修改运行地址,默认0.0.0.0:808.
func WithHostPorts(address string) Option {
	return Option{F: func(o *Options) {
		o.Addr = address
	}}
}

// WithNetwork 修改启动协议,默认TCP.
func WithNetwork(network string) Option {
	return Option{F: func(o *Options) {
		o.Network = network
	}}
}

// WithHasSubcontract 是否过滤分包的报文,默认过滤.
func WithHasSubcontract(filter bool) Option {
	return Option{F: func(o *Options) {
		o.FilterSubcontract = filter
	}}
}

// WithCustomHandleFunc 自定义报文处理方式.
func WithCustomHandleFunc(customHandleFunc func() map[consts.JT808CommandType]Handler) Option {
	return Option{F: func(o *Options) {
		o.CustomHandleFunc = customHandleFunc
	}}
}

// WithKeyFunc 自定义Key方式,key必须唯一,默认是手机号.
func WithKeyFunc(keyFunc func(message *Message) (string, bool)) Option {
	return Option{F: func(o *Options) {
		o.KeyFunc = keyFunc
	}}
}

// WithCustomTerminalEventer 自定义终端事件,包括加入,退出,读取数据等事件.
func WithCustomTerminalEventer(customTerminalEventerFunc func() TerminalEventer) Option {
	return Option{F: func(o *Options) {
		o.CustomTerminalEventerFunc = customTerminalEventerFunc
	}}
}
