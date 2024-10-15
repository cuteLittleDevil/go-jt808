package main

import (
	"flag"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"os"
	"simulator/internal/mq"
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
	var natsAddr string
	flag.StringVar(&natsAddr, "nats", "127.0.0.1:4222", "nats address")
	flag.Parse()

	if err := mq.Init(natsAddr); err != nil {
		slog.Error("nats init fail",
			slog.String("nats", natsAddr),
			slog.String("err", err.Error()))
		return
	}

	goJt808 := service.New(
		service.WithHostPorts("0.0.0.0:8080"),
		service.WithNetwork("tcp"),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				consts.T0200LocationReport:      &T0x0200{},
				consts.T0704LocationBatchUpload: &T0x0704{},
			}
		}),
	)
	goJt808.Run()
}
