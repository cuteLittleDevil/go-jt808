package main

import (
	"distributed_cluster/internal/mq"
	"distributed_cluster/internal/shared"
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

var (
	natsAddr string
	records  sync.Map
)

func main() {
	flag.StringVar(&natsAddr, "nats", "192.168.1.66:4222", "nats address")
	flag.Parse()

	if err := mq.Init(natsAddr); err != nil {
		slog.Error("nats init fail",
			slog.String("nats", natsAddr),
			slog.String("err", err.Error()))
		return
	}

	mq.Default().Run(map[string]nats.MsgHandler{
		shared.InitSubjectPrefix + ".*.*": func(msg *nats.Msg) {
			var data shared.Data
			if err := data.Parse(msg.Data); err == nil {
				fmt.Println("加入", data.Key)
				records.Store(data.Key, struct{}{})
			}
		},
		shared.LeaveSubjectPrefix + ".*.*": func(msg *nats.Msg) {
			var data shared.Data
			if err := data.Parse(msg.Data); err == nil {
				fmt.Println("离开", data.Key)
				records.Delete(data.Key)
			}
		},
		shared.ReadSubjectPrefix + ".*.*": func(msg *nats.Msg) {
			var data shared.Data
			if err := data.Parse(msg.Data); err == nil {
				if _, ok := records.Load(data.Key); !ok {
					fmt.Println("收到读数据 加入", data.Key)
					records.Store(data.Key, struct{}{})
				}
			}
		},
	})
	http.HandleFunc("/8103", handleRequest)
	_ = http.ListenAndServe(":12310", nil)
}

func handleRequest(writer http.ResponseWriter, _ *http.Request) {
	var (
		wg           sync.WaitGroup
		replyRecords sync.Map
	)
	str := ""
	records.Range(func(key, _ any) bool {
		sim := key.(string)
		p8103 := model.P0x8103{
			ParamTotal: 2,
			TerminalParamDetails: model.TerminalParamDetails{
				T0x001HeartbeatInterval: model.ParamContent[uint32]{
					ID:    0x01,
					Len:   4,
					Value: 10, // 心跳10秒
				},
				T0x029DefaultReportingTimeInterval: model.ParamContent[uint32]{
					ID:    0x29,
					Len:   4,
					Value: 5, // 默认上报间隔5秒
				},
			},
		}
		notice := shared.Notice{
			Key:              sim,
			Command:          p8103.Protocol(),
			Body:             p8103.Encode(),
			OverTimeDuration: 3 * time.Second,
			UUID:             uuid.New().String(),
		}
		_ = mq.Default().Pub(notice.Subject(), notice.ToBytes())
		str += fmt.Sprintf("发送通知sim [%s]\n", sim)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if data, err := mq.Default().Sub(notice.ReplySubject(), 3*time.Second); err == nil {
				replyRecords.Store(sim, data)
			}
		}()
		return true
	})
	wg.Wait()

	replyRecords.Range(func(key, value any) bool {
		sim := key.(string)
		v := value.([]byte)
		var data shared.Data
		_ = data.Parse(v)
		var t0x0001 model.T0x0001
		if err := t0x0001.Parse(data.Message.JTMessage); err == nil {
			str += fmt.Sprintf("收到回复 sim[%s] %s\n", sim, t0x0001.String())
		}
		return true
	})

	_, _ = writer.Write([]byte(str))
}
