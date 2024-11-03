package service

import (
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type Option struct {
	F func(o *Options)
}

const (
	defaultAddr             = ":8888"
	defaultNetwork          = "tcp"
	defaultHasFilterSubPack = true // 读写事件是否过滤分包的情况
)

type Options struct {
	CustomHandleFunc func() map[consts.JT808CommandType]Handler
	Addr             string
	Network          string
	HasFilterSubPack bool
	KeyFunc          func(message *Message) string
}

func (o *Options) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}

func NewOptions(opts []Option) *Options {
	options := &Options{
		Addr:             defaultAddr,
		Network:          defaultNetwork,
		HasFilterSubPack: defaultHasFilterSubPack,
	}
	options.Apply(opts)
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
		o.HasFilterSubPack = hasFilter
	}}
}

func WithCustomHandleFunc(customHandleFunc func() map[consts.JT808CommandType]Handler) Option {
	return Option{F: func(o *Options) {
		o.CustomHandleFunc = customHandleFunc
	}}
}

func WithKeyFunc(keyFunc func(message *Message) string) Option {
	return Option{F: func(o *Options) {
		o.KeyFunc = keyFunc
	}}
}
