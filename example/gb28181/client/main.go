package main

import (
	"client/conf"
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/gb28181"
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"
)

func init() {
	if err := conf.InitConfig("./config.yaml"); err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "source" {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					// 只保留文件名部分
					a.Value = slog.AnyValue(filepath.Base(source.File))
				}
			}
			return a
		},
	}))
	slog.SetDefault(logger)
}

func main() {
	var (
		platform = conf.GetData().GB28181.Platform
		device   = conf.GetData().GB28181.Device
	)
	client := gb28181.New("1001", gb28181.WithPlatformInfo(gb28181.PlatformInfo{
		Domain:   platform.Domain,
		ID:       platform.ID,
		Password: platform.Password,
		IP:       platform.IP,
		Port:     platform.Port,
	}), gb28181.WithDeviceInfo(gb28181.DeviceInfo{
		ID: device.ID,
		// 实际不会用到设备的IP和端口 只是sip传输过去
		IP:   device.IP,
		Port: device.Port,
	}),
		gb28181.WithTransport(conf.GetData().GB28181.Transport), // 信令默认使用UDP 也可以TCP
		gb28181.WithKeepAliveSecond(30),                         // 心跳保活周期30秒
		gb28181.WithInviteEventFunc(func(info *command.InviteInfo) *command.InviteInfo {
			// 默认jt1078收流端口是 gb28181 - 100
			// 如gb28181收流端口是10100 则jt1078收流端口是10000
			// 完成9101请求 让设备发送jt1078流
			// 流媒体默认选择的是音视频流 视频h264 音频g711a 测试文件是h264的jt1078
			//info.JT1078Info.StreamTypes = []jt1078.PTType{jt1078.PTH264}
			info.JT1078Info.Port = info.Port - 100
			info.JT1078Info.RtpTypeConvert = func(pt jt1078.PTType) (byte, bool) {
				// 默认是按照h264=98的国标 但是zlm会失败 因此可以自行修改
				if pt == jt1078.PTH264 {
					return 96, true
				}
				// 其他情况使用默认的国标规范
				return 0, false
			}
			// 只支持TCP被动模式 即设备TCP链接到gb28181平台
			// 模拟9101下发到设备 设备上传jt1078流到info.JT1078Info.Port
			go sendJT1078Packet(info.JT1078Info.Port)
			return info
		}),
	)
	if err := client.Init(); err != nil {
		panic(err)
	}
	go client.Run()
	time.Sleep(3000 * time.Second)
	client.Stop()
}

func sendJT1078Packet(port int) {
	time.Sleep(time.Second) // 模拟下发指令给设备 设备推流
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", port))
	if err == nil {
		content, err := os.ReadFile(conf.GetData().JT1078.File)
		if err != nil {
			panic(err)
		}
		data, _ := hex.DecodeString(string(content))
		const groupSum = 1023
		for {
			start := 0
			end := 0
			for i := 0; i < len(data)/groupSum; i++ {
				start = i * groupSum
				end = start + groupSum
				_, _ = conn.Write(data[start:end])
				time.Sleep(20 * time.Millisecond)
			}
			_, _ = conn.Write(data[end:])
		}
	}
}
