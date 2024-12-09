package main

import (
	"distributed_cluster/internal/mq"
	"distributed_cluster/internal/shared"
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/nats-io/nats.go"
	"log/slog"
	"os"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(logger)
}

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

	goJt808 := service.New(
		service.WithHostPorts("0.0.0.0:808"),
		service.WithNetwork("tcp"),
		service.WithCustomTerminalEventer(func() service.TerminalEventer {
			return &meTerminal{id: id} // 自定义终端事件 终端进入 离开 读写报文事件
		}),
	)

	mq.Default().Run(map[string]nats.MsgHandler{
		shared.NoticeSubjectPrefix + ".*": func(msg *nats.Msg) {
			var notice shared.Notice
			if err := notice.Parse(msg.Data); err == nil {
				fmt.Println("收到主动下发请求", notice.Key)
				go func() {
					replyMsg := goJt808.SendActiveMessage(&service.ActiveMessage{
						Key:              notice.Key,
						Command:          notice.Command,
						Body:             notice.Body,
						OverTimeDuration: notice.OverTimeDuration,
					})
					if replyMsg.ExtensionFields.Err == nil {
						data := shared.NewData(id, shared.OnNoticeComplete, *replyMsg)
						_ = mq.Default().Pub(notice.ReplySubject(), data.ToBytes())
					} else {
						fmt.Println("回复的异常情况", replyMsg.ExtensionFields.Err)
					}
				}()
			}
		},
	})

	goJt808.Run()
}
