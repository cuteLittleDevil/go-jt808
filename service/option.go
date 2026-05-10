package service

import (
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"time"
)

const (
	defaultAddr              = "0.0.0.0:808" // 服务默认地址
	defaultNetwork           = "tcp"         // 服务默认网络协议
	defaultFilterSubcontract = true          // 是否过滤分包的情况
	defaultTimeout           = 0             // 默认不开启超时检测
)

type (
	Option struct {
		F func(o *Options)
	}

	Options struct {
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
		// CustomActiveRespondHandlerFunc 自定义消息回复处理
		CustomActiveRespondHandlerFunc func() map[consts.JT808CommandType]func(
			platformMsg *ActiveMessage, terminalMsg *Message) bool
		// IdleTimeout 空闲超时,默认0，不设置
		IdleTimeout time.Duration
		// OnTerminalTimeoutEvent 空闲超时事件，设置IdleTimeout时触发
		OnTerminalTimeoutEvent func(TerminalTimeout)
	}

	TerminalTimeout struct {
		// Key 连接终端唯一标识，通常是sim卡号.
		Key string
		// ConnectionStartTime 连接建立开始时间.
		ConnectionStartTime time.Time
		// Address 客户端的地址. 格式如127.0.0.1:58994.
		Address string
		// FirstPacketTime 收到第一个报文的时间.
		FirstPacketTime time.Time
		// LastPacketTime 收到最后一个报文的时间.
		LastPacketTime time.Time
		// IdleTimeout 设置的空闲超时间隔.
		IdleTimeout time.Duration
	}
)

func newOptions(opts []Option) *Options {
	options := &Options{
		Addr:              defaultAddr,
		Network:           defaultNetwork,
		FilterSubcontract: defaultFilterSubcontract,
		IdleTimeout:       defaultTimeout,
		KeyFunc: func(message *Message) (string, bool) {
			return message.JTMessage.Header.TerminalPhoneNo, true
		},
		CustomTerminalEventerFunc: func() TerminalEventer {
			return &defaultTerminalEvent{}
		},
		CustomHandleFunc: func() map[consts.JT808CommandType]Handler {
			return map[consts.JT808CommandType]Handler{}
		},
		CustomActiveRespondHandlerFunc: func() map[consts.JT808CommandType]func(
			activeMsg *ActiveMessage, terminalMsg *Message) bool {
			return map[consts.JT808CommandType]func(activeMsg *ActiveMessage, terminalMsg *Message) bool{}
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

// WithNetwork 修改启动协议,默认TCP.(暂不支持udp).
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
// 用于覆盖或扩展默认的 JT808 指令处理逻辑,每一个连接都是独立的.
func WithCustomHandleFunc(customFunc func() map[consts.JT808CommandType]Handler) Option {
	return Option{F: func(o *Options) {
		o.CustomHandleFunc = customFunc
	}}
}

// WithCustomActiveRespondHandlerFunc 自定义主动消息回复报文处理.
func WithCustomActiveRespondHandlerFunc(customFunc func() map[consts.JT808CommandType]func(
	activeMsg *ActiveMessage, terminalMsg *Message) bool) Option {
	return Option{F: func(o *Options) {
		o.CustomActiveRespondHandlerFunc = customFunc
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

// WithTerminalTimeout 设置终端空闲超时时间和超时事件处理函数
//
// 参数说明：
//
//		idleTimeout    - 空闲超时时间（单位：秒建议使用 N * time.Second）
//		                 设置为 <= 0 表示不启用超时检测（默认行为）
//		onTimeoutEvent - 超时事件回调函数，为 nil 时使用默认处理
//	                  firstPacketTime为首次报文时间，lastPacketTime为最后一次报文时间
//
// 使用示例：
//
//	service.WithTerminalTimeout(30*time.Second, func(t TerminalTimeout) {
//		fmt.Println(fmt.Sprintf("key=[%s] addr[%s] 首次报文时间[%s] 最后一次报文时间[%s] 运行时间[%v]",
//			t.Key, t.Address, t.FirstPacketTime.Format(time.DateTime),
//			t.LastPacketTime.Format(time.DateTime), time.Since(t.ConnectionStartTime)))
//	}),
func WithTerminalTimeout(idleTimeout time.Duration, onTimeoutEvent func(TerminalTimeout)) Option {
	return Option{F: func(o *Options) {
		if idleTimeout > 0 {
			o.IdleTimeout = idleTimeout
		}
		if onTimeoutEvent != nil {
			o.OnTerminalTimeoutEvent = onTimeoutEvent
		}
	}}
}
