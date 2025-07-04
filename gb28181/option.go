package gb28181

import (
	"gb28181/command"
	"time"
)

type Option struct {
	F func(o *Options)
}

type (
	Options struct {
		Sim string
		DeviceInfo
		PlatformInfo
		Transport         string
		UserAgent         string
		KeepAlive         time.Duration
		MappingRuleFunc   func(gb28181Port int) (jt1078Port int)
		OnInviteEventFunc func(command.InviteInfo)
	}

	PlatformInfo struct {
		Domain string
		// ID 服务器ID
		ID       string
		Password string
		IP       string
		Port     int
	}

	// DeviceInfo 实际不会用到设备的IP和端口 只是sip传输过去
	// 通道默认是ID最后一位换1-4.
	DeviceInfo struct {
		ID   string
		IP   string
		Port int
	}
)

func WithPlatformInfo(p PlatformInfo) Option {
	return Option{F: func(o *Options) {
		o.PlatformInfo = p
	}}
}

func WithDeviceInfo(d DeviceInfo) Option {
	return Option{F: func(o *Options) {
		o.DeviceInfo = d
	}}
}

func WithTransport(transport string) Option {
	return Option{F: func(o *Options) {
		o.Transport = transport
	}}
}

func WithKeepAliveSecond(second int) Option {
	return Option{F: func(o *Options) {
		o.KeepAlive = time.Duration(second) * time.Second
	}}
}

// WithMappingRuleFunc 设置映射规则 输入gb28181的收流端口 输出jt1078的收流端口.
func WithMappingRuleFunc(f func(gb28181Port int) int) Option {
	return Option{F: func(o *Options) {
		o.MappingRuleFunc = f
	}}
}

// WithInviteEventFunc 收到invite事件时触发.
func WithInviteEventFunc(f func(command.InviteInfo)) Option {
	return Option{F: func(o *Options) {
		o.OnInviteEventFunc = f
	}}
}
