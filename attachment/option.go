package attachment

import "github.com/cuteLittleDevil/go-jt808/shared/consts"

type Option struct {
	F func(o *Options)
}

const (
	defaultAddr             = "0.0.0.0:10808" // 服务默认地址
	defaultNetwork          = "tcp"           // 服务默认网络协议
	defaultActiveSafetyType = consts.ActiveSafetyJS
)

type Options struct {
	Addr             string
	Network          string
	FileEventerFunc  func() FileEventer
	ActiveSafetyType consts.ActiveSafetyType
	DataHandleFunc   func() DataHandler
}

func newOptions(opts []Option) *Options {
	options := &Options{
		Addr:             defaultAddr,
		Network:          defaultNetwork,
		ActiveSafetyType: defaultActiveSafetyType,
		FileEventerFunc: func() FileEventer {
			return newFileEvent()
		},
	}
	for _, op := range opts {
		op.F(options)
	}
	return options
}

// WithHostPorts 修改运行地址 默认0.0.0.0:808.
func WithHostPorts(address string) Option {
	return Option{F: func(o *Options) {
		o.Addr = address
	}}
}

// WithNetwork 修改启动协议 默认TCP.
func WithNetwork(network string) Option {
	return Option{F: func(o *Options) {
		o.Network = network
	}}
}

// WithFileEventerFunc 自定义文件事件变化.
func WithFileEventerFunc(handleFunc func() FileEventer) Option {
	return Option{F: func(o *Options) {
		o.FileEventerFunc = handleFunc
	}}
}

// WithActiveSafetyType 使用什么标准的主动安全报文.
func WithActiveSafetyType(activeSafetyType consts.ActiveSafetyType) Option {
	return Option{F: func(o *Options) {
		o.ActiveSafetyType = activeSafetyType
	}}
}

// WithDataHandleFunc 自定义数据处理.
func WithDataHandleFunc(handleFunc func() DataHandler) Option {
	return Option{F: func(o *Options) {
		o.DataHandleFunc = handleFunc
	}}
}
