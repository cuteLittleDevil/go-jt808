package adapter

import (
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"time"
)

type Option struct {
	F func(o *Options)
}

const (
	defaultAddr         = "0.0.0.0:18080" // 服务默认地址
	defaultTimeoutRetry = 10 * time.Second
)

type Options struct {
	Addr         string
	TimeoutRetry time.Duration
	Terminals    []Terminal
	AllowCommand []consts.JT808CommandType
}

func newOptions(opts []Option) *Options {
	options := &Options{
		Addr:         defaultAddr,
		TimeoutRetry: defaultTimeoutRetry,
		AllowCommand: []consts.JT808CommandType{
			// 平台下发的
			consts.P8003ReissueSubcontractingRequest,
			consts.P8103SetTerminalParams,
			consts.P8104QueryTerminalParams,
			consts.P8801CameraShootImmediateCommand,
			// JT1078相关的
			consts.P9003QueryTerminalAudioVideoProperties,
			consts.P9101RealTimeAudioVideoRequest,
			consts.P9102AudioVideoControl,
			consts.P9205QueryResourceList,
			consts.P9207FileUploadControl,
			// 主动安全扩展的
			consts.P9208AlarmAttachUpload,
		},
	}
	for _, op := range opts {
		op.F(options)
	}
	return options
}

// WithHostPorts 修改运行地址 默认0.0.0.0:808
func WithHostPorts(address string) Option {
	return Option{F: func(o *Options) {
		o.Addr = address
	}}
}

// WithTerminals 自定义转发的客户端情况
func WithTerminals(terminals ...Terminal) Option {
	return Option{F: func(o *Options) {
		o.Terminals = terminals
	}}
}

// WithAllowCommand 自定义Follower模式允许的回复命令
func WithAllowCommand(commands ...consts.JT808CommandType) Option {
	return Option{F: func(o *Options) {
		o.AllowCommand = commands
	}}
}

// WithTimeoutRetry 自定义超时重试时间 和自定义各808服务异常断开后 多久尝试重新连接
func WithTimeoutRetry(timeout time.Duration) Option {
	return Option{F: func(o *Options) {
		o.TimeoutRetry = timeout
	}}
}
