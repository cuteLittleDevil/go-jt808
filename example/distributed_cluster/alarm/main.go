package main

import (
	"distributed_cluster/internal/mq"
	"distributed_cluster/internal/shared"
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/nats-io/nats.go"
	"log/slog"
)

func main() {
	var (
		natsAddr string
		id       string
	)
	flag.StringVar(&natsAddr, "nats", "192.168.1.66:4222", "nats address")
	flag.StringVar(&id, "id", "1", "service id 用于区分服务来源")
	flag.Parse()

	if err := mq.Init(natsAddr); err != nil {
		slog.Error("nats init fail",
			slog.String("nats", natsAddr),
			slog.String("err", err.Error()))
		return
	}

	mq.Default().Run(map[string]nats.MsgHandler{
		// shared.ReadSubjectPrefix + ".*.*"
		// shared.ReadSubjectPrefix + ".>"
		// shared.ReadSubjectPrefix + ".*.512" 512是0x0200
		//shared.ReadSubjectPrefix + ".*.*": func(msg *nats.Msg) {
		//	var data shared.Data
		//	if err := data.Parse(msg.Data); err == nil {
		//		switch data.Message.Command {
		//		case consts.T0200LocationReport:
		//			// 收到位置信息 保存经纬度
		//			var t0x0200 model.T0x0200
		//			if err := t0x0200.Parse(data.Message.JTMessage); err == nil {
		//				fmt.Println(fmt.Sprintf("保存经纬度 id[%s] sim[%s] %s",
		//					data.ID, data.Key, t0x0200.String()))
		//			}
		//		}
		//	}
		//},
		shared.ReadSubjectPrefix + ".*.512": func(msg *nats.Msg) {
			var data shared.Data
			if err := data.Parse(msg.Data); err == nil {
				// 收到位置信息 保存经纬度
				var t0x0200 model.T0x0200
				if err := t0x0200.Parse(data.Message.JTMessage); err == nil {
					fmt.Println(fmt.Sprintf("保存经纬度 id[%s] sim[%s] %s",
						data.ID, data.Key, t0x0200.String()))
				} else {
					panic(err)
				}
			}
		},
	})
	select {}
}
