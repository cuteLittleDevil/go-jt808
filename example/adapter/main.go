package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/adapter"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"os"
	"time"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(logger)

	{
		goJt808 := service.New( // 模拟已经在用的其他808服务
			service.WithHostPorts("0.0.0.0:18081"),
			service.WithNetwork("tcp"),
		)
		go goJt808.Run()
	}
	{
		ch := make(chan string, 10)
		goJt808 := service.New( // 使用go-jt808
			service.WithHostPorts("0.0.0.0:18082"),
			service.WithNetwork("tcp"),
			service.WithCustomTerminalEventer(func() service.TerminalEventer {
				return &meTerminal{ch: ch}
			}),
		)
		go func() {
			for key := range ch {
				time.Sleep(5 * time.Second) // 确保发送92025的时候 序号是3 这样就可以有正确的回复
				p9205 := model.P0x9205{
					BaseHandle:  model.BaseHandle{},
					ChannelNo:   1,
					StartTime:   time.Now().Add(-24 * time.Hour).Format(time.DateTime),
					EndTime:     time.Now().Format(time.DateTime),
					AlarmFlag:   0,
					MediaType:   0,
					StreamType:  0,
					StorageType: 0,
				}
				replyMsg := goJt808.SendActiveMessage(&service.ActiveMessage{
					Key:              key,
					Command:          p9205.Protocol(),
					Body:             p9205.Encode(),
					OverTimeDuration: 5 * time.Second,
				})
				if replyMsg.ExtensionFields.Err == nil {
					fmt.Println("9205 获取到回复")
				} else {
					fmt.Println("9205 获取回复失败", replyMsg.ExtensionFields.Err)
				}
			}
		}()
		go goJt808.Run()
	}
}

func main() {
	address := "0.0.0.0:18080"
	adapterGroup := adapter.New(
		adapter.WithHostPorts(address),
		//adapter.WithAllowCommand( // 全局都允许的向设备写的命令
		//	consts.P9101RealTimeAudioVideoRequest,
		//),
		adapter.WithTimeoutRetry(10*time.Second), // 模拟连接断开后 多久重试一次
		adapter.WithTerminals(
			adapter.Terminal{
				Mode:       adapter.Leader,    // 服务和设备之间读写全部正常
				TargetAddr: "127.0.0.1:18081", // 其他项目的jt808服务
			},
			adapter.Terminal{
				Mode:       adapter.Follower,  // 服务读正常 写默认拒绝（只下发指定命令）
				TargetAddr: "127.0.0.1:18082", // go-jt808项目的jt808服务
				AllowCommands: []consts.JT808CommandType{
					consts.P9205QueryResourceList, // 允许向设备发送的命令
				},
			},
		),
	)
	//go func() {
	//	time.Sleep(time.Second)
	//	client("1001", address)
	//}()
	adapterGroup.Run()
}
