package gb28181

import (
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
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
		OnInviteEventFunc func(*command.InviteInfo) *command.InviteInfo
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

// WithInviteEventFunc 收到invite事件时触发.
func WithInviteEventFunc(f func(*command.InviteInfo) *command.InviteInfo) Option {
	return Option{F: func(o *Options) {
		o.OnInviteEventFunc = f
	}}
}
