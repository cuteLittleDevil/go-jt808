package main

import (
	"github.com/cuteLittleDevil/go-jt808/gb28181"
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"log/slog"
	"os"
	"path/filepath"
	"rtsp/conf"
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
		gb28181.WithToGBType(command.CustomPS),
		gb28181.WithToGB28181er(func() command.ToGB28181er {
			// rtsp流转ps流推送 流程成功了 但是没有实际的rtsp流 未测试
			return &zlmHandle{}
		}),
	)
	if err := client.Init(); err != nil {
		panic(err)
	}
	go client.Run()
	time.Sleep(3000 * time.Second)
	client.Stop()
}
