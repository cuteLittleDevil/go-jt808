package gb28181

import (
	"gb28181/command"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"time"
)

func Example() {
	client := New("1001", WithPlatformInfo(PlatformInfo{
		Domain:   "34020000002",
		ID:       "34020000002000000001",
		Password: "123456",
		IP:       "127.0.0.1",
		Port:     15060,
	}), WithDeviceInfo(DeviceInfo{
		ID: "34020000001320000330",
		// 实际不会用到设备的IP和端口 只是sip传输过去
		IP:   "127.0.0.1",
		Port: 5060,
	}),
		WithTransport("UDP"),
		WithKeepAliveSecond(10),
		WithInviteEventFunc(func(info *command.InviteInfo) *command.InviteInfo {
			// 完成9101请求 让设备发送jt1078流
			// 流媒体默认选择的是音视频流 视频h264 音频g711a
			info.JT1078Info.StreamTypes = []jt1078.PTType{jt1078.PTH264, jt1078.PTG711A}
			// 默认jt1078收流端口是 gb28181 - 100
			// 如gb28181收流端口是10100 则jt1078收流端口是10000
			info.JT1078Info.Port = info.Port - 100
			return info
		}),
	)
	if err := client.Init(); err != nil {
		panic(err)
	}
	go client.Run()
	time.Sleep(30 * time.Second)
	client.Stop()

	// Output:
}
