package adapter

import (
	"time"

	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type Option struct {
	F func(o *Options)
}

const (
	defaultAddr         = "0.0.0.0:18080" // 服务默认地址
	defaultTimeoutRetry = 10 * time.Second
	defaultDialTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 0
	// defaultMaxFrameBufferSize 下行粘包缓冲区默认上限 单包读取缓冲2048的2倍仍未组成完整帧视为异常.
	defaultMaxFrameBufferSize = readBufSize * 2
)

type Options struct {
	Addr               string
	TimeoutRetry       time.Duration
	DialTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	MaxFrameBufferSize int
	Terminals          []Terminal
	AllowCommand       []consts.JT808CommandType
}

func newOptions(opts []Option) *Options {
	options := &Options{
		Addr:               defaultAddr,
		TimeoutRetry:       defaultTimeoutRetry,
		DialTimeout:        defaultDialTimeout,
		WriteTimeout:       defaultWriteTimeout,
		IdleTimeout:        defaultIdleTimeout,
		MaxFrameBufferSize: defaultMaxFrameBufferSize,
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

// WithHostPorts 修改运行地址 默认0.0.0.0:18080.
func WithHostPorts(address string) Option {
	return Option{F: func(o *Options) {
		o.Addr = address
	}}
}

// WithTerminals 自定义转发的客户端情况.
func WithTerminals(terminals ...Terminal) Option {
	return Option{F: func(o *Options) {
		o.Terminals = terminals
	}}
}

// WithAllowCommand 覆盖全局Follower模式默认允许下发的命令列表(传空则清空全局默认列表,仅使用各Terminal自身的AllowCommands).
func WithAllowCommand(commands ...consts.JT808CommandType) Option {
	return Option{F: func(o *Options) {
		o.AllowCommand = commands
	}}
}

// WithTimeoutRetry 自定义各808服务异常断开后多久尝试重新连接 默认10秒(<=0保持默认).
func WithTimeoutRetry(timeout time.Duration) Option {
	return Option{F: func(o *Options) {
		if timeout > 0 {
			o.TimeoutRetry = timeout
		}
	}}
}

// WithDialTimeout 自定义连接后端808服务的拨号超时时间 默认10秒(<=0保持默认).
func WithDialTimeout(timeout time.Duration) Option {
	return Option{F: func(o *Options) {
		if timeout > 0 {
			o.DialTimeout = timeout
		}
	}}
}

// WithWriteTimeout 自定义单次TCP写超时时间 默认10秒(<=0保持默认).
func WithWriteTimeout(timeout time.Duration) Option {
	return Option{F: func(o *Options) {
		if timeout > 0 {
			o.WriteTimeout = timeout
		}
	}}
}

// WithIdleTimeout 设置真实终端连接的读空闲超时时间 默认0表示不启用.
func WithIdleTimeout(idleTimeout time.Duration) Option {
	return Option{F: func(o *Options) {
		if idleTimeout > 0 {
			o.IdleTimeout = idleTimeout
		}
	}}
}

// WithMaxFrameBufferSize 设置下行粘包缓冲区最大字节数,默认2048*2.
// 收到的数据(如7e开头但迟迟没有结束的7e,或非0x7e开头的异常数据)无法组成完整帧时会暂存在缓冲区,
// 暂存数据达到该值时断开该后端TCP连接并重连,防止异常数据导致内存无限增长. size<=0表示不限制.
func WithMaxFrameBufferSize(size int) Option {
	return Option{F: func(o *Options) {
		o.MaxFrameBufferSize = size
	}}
}
