package main

import (
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/nats-io/nats.go"
	"log/slog"
	"os"
	"web/alarm/command"
	"web/alarm/conf"
	"web/alarm/file"
	"web/internal/mq"
	"web/internal/shared"
)

func init() {
	if err := conf.InitConfig("./config.yaml"); err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(logger)
	hlog.SetLevel(3)
}

func main() {
	natsAddr := conf.GetData().NatsConfig.Address
	if err := mq.Init(natsAddr); err != nil {
		slog.Error("nats init fail",
			slog.String("nats", natsAddr),
			slog.String("err", err.Error()))
		return
	}

	mq.Default().Run(map[string]nats.MsgHandler{
		shared.ReadSubjectPrefix + ".*.*.1796": func(msg *nats.Msg) {
			var data shared.EventData
			if err := data.Parse(msg.Data); err == nil {
				var t0x0704 command.BatchLocation
				if err := t0x0704.Parse(data.JTMessage); err == nil {
					// 收到位置信息 保存经纬度
					//for _, item := range t0x0704.Items {
					//	fmt.Println(fmt.Sprintf("0x0704保存经纬度 id[%s] sim[%s] %s",
					//		data.ID, data.Key, item.T0x0200LocationItem.String()))
					//}
					for _, location := range t0x0704.AlarmLocations {
						// 附件处理
						go file.OnAlarmEvent(data, location)
					}
				}
			}
		},
		// 事件.服务ID.手机号.报文类型 512 = 0x0200
		shared.ReadSubjectPrefix + ".*.*.512": func(msg *nats.Msg) {
			var data shared.EventData
			if err := data.Parse(msg.Data); err == nil {
				// 收到位置信息 保存经纬度
				var t0x0200 command.Location
				if err := t0x0200.Parse(data.JTMessage); err == nil {
					//fmt.Println(fmt.Sprintf("保存经纬度 id[%s] sim[%s] %s",
					//	data.ID, data.Key, t0x0200.String()))
					// 附件处理
					go file.OnAlarmEvent(data, t0x0200)
				}
			}
		},
	})
	select {}
}
