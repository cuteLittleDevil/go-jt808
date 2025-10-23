package gb28181

import (
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"github.com/cuteLittleDevil/go-jt808/gb28181/internal/stream"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"time"
)

func Example() {
	client := New("1001", WithPlatformInfo(PlatformInfo{
		Domain:   "34020000002",          // 平台域
		ID:       "34020000002000000001", // 平台id
		Password: "123456",               // 平台密码
		IP:       "127.0.0.1",            // 平台ip
		Port:     15060,                  // 平台端口
	}), WithDeviceInfo(DeviceInfo{
		ID: "34020000001320000330", // 设备ID
		// 实际不会用到设备的IP和端口 只是sip传输过去
		IP:   "127.0.0.1",
		Port: 5060,
	}),
		WithTransport("UDP"),    // 默认使用UDP 也可以用TCP
		WithKeepAliveSecond(10), // 心跳保活周期10秒
		WithInviteEventFunc(func(info *command.InviteInfo) *command.InviteInfo {
			// 也可以选择收到invite 直接把ps流传输到info.IP info.Port上

			// 流媒体默认选择的是音视频流 视频h264 音频g711a
			info.JT1078Info.StreamTypes = []jt1078.PTType{jt1078.PTH264, jt1078.PTG711A}
			info.JT1078Info.RtpTypeConvert = func(pt jt1078.PTType) (byte, bool) {
				// 默认是按照h264=98的国标 但是zlm会失败 因此可以自行修改
				if pt == jt1078.PTH264 {
					return 96, true
				}
				// 其他情况使用默认的国标规范
				return 0, false
			}
			// 默认jt1078收流端口是 gb28181 - 100
			// 如gb28181收流端口是10100 则jt1078收流端口是10000
			info.JT1078Info.Port = info.Port - 100

			// 完成9101请求 让设备发送jt1078流
			// 或者参考 https://github.com/cuteLittleDevil/go-jt808/blob/main/example/gb28181/main.go
			// 直接发送jt1078流测试
			return info
		}),
		// command.JT1078ToPSFilterPacket 可选择有报文错误就过滤的
		WithToGBType(command.JT1078ToPS), // 默认报文有错误就退出
		WithToGB28181er(func() command.ToGB28181er {
			// 默认流处理 jt1078转ps流
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
