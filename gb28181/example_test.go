package gb28181

import (
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"github.com/cuteLittleDevil/go-jt808/gb28181/internal/stream"
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
			info.JT1078Info.RtpTypeConvert = func(pt jt1078.PTType) (byte, bool) {
				// 默认是按照h264=98的国标 但是zlm会不失败 因此可以自行修改
				if pt == jt1078.PTH264 {
					return 96, true
				}
				// 其他情况使用默认的国标规范
				return 0, false
			}
			// 默认jt1078收流端口是 gb28181 - 100
			// 如gb28181收流端口是10100 则jt1078收流端口是10000
			info.JT1078Info.Port = info.Port - 100
			return info
		}),
		WithJT1078ToGB28181er(func() command.JT1078ToGB28181er {
			// 目前模拟jt1078包转gb28181在m7s上成功 zlm上播放失败
			// 目前这是内部包的实现 不暴露出来
			return stream.NewJT1078T0GB28181()
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
