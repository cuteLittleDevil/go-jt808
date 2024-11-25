package attachment

type Option struct {
	F func(o *Options)
}

const (
	defaultAddr    = "0.0.0.0:10808" // 服务默认地址
	defaultNetwork = "tcp"           // 服务默认网络协议
)

type Options struct {
	Addr                 string
	Network              string
	FileEventerFunc      func() FileEventer
	StreamDataHandleFunc func() StreamDataHandler
	JT808DataHandleFunc  func() JT808DataHandler
}

func newOptions(opts []Option) *Options {
	options := &Options{
		Addr:    defaultAddr,
		Network: defaultNetwork,
		FileEventerFunc: func() FileEventer {
			return nil
		},
		StreamDataHandleFunc: func() StreamDataHandler {
			return newSuBiaoStreamDataHandle()
		},
		JT808DataHandleFunc: func() JT808DataHandler {
			return nil
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

// WithNetwork 修改启动协议 默认TCP
func WithNetwork(network string) Option {
	return Option{F: func(o *Options) {
		o.Network = network
	}}
}

// WithFileEventerFunc 自定义文件事件变化
func WithFileEventerFunc(handleFunc func() FileEventer) Option {
	return Option{F: func(o *Options) {
		o.FileEventerFunc = handleFunc
	}}
}

// WithStreamDataHandlerFunc 自定义主动安全的文件流报文处理
func WithStreamDataHandlerFunc(handleFunc func() StreamDataHandler) Option {
	return Option{F: func(o *Options) {
		o.StreamDataHandleFunc = handleFunc
	}}
}

// WithJT808DataHandleFunc 自定义主动安全报文处理
func WithJT808DataHandleFunc(handleFunc func() JT808DataHandler) Option {
	return Option{F: func(o *Options) {
		o.JT808DataHandleFunc = handleFunc
	}}
}
