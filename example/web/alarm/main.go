package main

import (
	"encoding/json"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"web/alarm/command"
	"web/alarm/conf"
	"web/alarm/file"
	"web/alarm/pool"
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

	isAlarm := conf.GetData().AlarmConfig.Enable
	batchPool := pool.NewHandle[command.BatchLocation]()
	go batchPool.Run()
	locationPool := pool.NewHandle[command.Location]()
	go locationPool.Run()
	filePool := pool.NewHandle[shared.T0x0801File]()
	go filePool.Run()

	mq.Default().Run(map[string]func(data shared.EventData){
		shared.ReadSubjectPrefix + ".*.*.1796": func(data shared.EventData) {
			var t0x0704 command.BatchLocation
			if err := t0x0704.Parse(data.JTMessage); err == nil {
				batchPool.Do(t0x0704)
				if isAlarm {
					for _, location := range t0x0704.AlarmLocations {
						// 附件处理
						go file.OnAlarmEvent(data, location)
					}
				}
			}
		},
		// 事件.服务ID.手机号.报文类型 512 = 0x0200
		shared.ReadSubjectPrefix + ".*.*.512": func(data shared.EventData) {
			var t0x0200 command.Location
			if err := t0x0200.Parse(data.JTMessage); err == nil {
				locationPool.Do(t0x0200)
				if isAlarm {
					go file.OnAlarmEvent(data, t0x0200)
				}
			}
		},
		// 事件.服务ID.手机号.报文类型 2049 = 0x0801
		shared.CustomSubjectPrefix + ".*.*.2049": func(data shared.EventData) {
			if v, ok := data.CustomData.(map[string]any); ok {
				b, _ := json.Marshal(v)
				var t0801 shared.T0x0801File
				if err := json.Unmarshal(b, &t0801); err == nil {
					filePool.Do(t0801)
				}
			}
		},
	})

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT) // kill -2
	<-quit
}
