package service

import (
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type Option struct {
	F func(o *Options)
}

const (
	defaultAddr          = "0.0.0.0:808" // 服务默认地址
	defaultNetwork       = "tcp"         // 服务默认网络协议
	defaultFilterSubPack = true          // 读写事件是否过滤分包的情况
)

type Options struct {
	Addr                      string
	Network                   string
	FilterSubPack             bool
	KeyFunc                   func(message *Message) (string, bool)
	CustomTerminalEventerFunc func() TerminalEventer
	CustomHandleFunc          func() map[consts.JT808CommandType]Handler
}

func NewOptions(opts []Option) *Options {
	options := &Options{
		Addr:          defaultAddr,
		Network:       defaultNetwork,
		FilterSubPack: defaultFilterSubPack,
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

func WithHostPorts(address string) Option {
	return Option{F: func(o *Options) {
		o.Addr = address
	}}
}

func WithNetwork(network string) Option {
	return Option{F: func(o *Options) {
		o.Network = network
	}}
}

func WithHasFilterSubPack(hasFilter bool) Option {
	return Option{F: func(o *Options) {
		o.FilterSubPack = hasFilter
	}}
}

func WithCustomHandleFunc(customHandleFunc func() map[consts.JT808CommandType]Handler) Option {
	return Option{F: func(o *Options) {
		o.CustomHandleFunc = customHandleFunc
	}}
}

func WithKeyFunc(keyFunc func(message *Message) (string, bool)) Option {
	return Option{F: func(o *Options) {
		o.KeyFunc = keyFunc
	}}
}

func WithCustomTerminalEventer(customTerminalEventerFunc func() TerminalEventer) Option {
	return Option{F: func(o *Options) {
		o.CustomTerminalEventerFunc = customTerminalEventerFunc
	}}
}
