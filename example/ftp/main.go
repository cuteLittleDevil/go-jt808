package main

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cuteLittleDevil/go-jt808/service"
	"log/slog"
	"os"
	"time"
)

const (
	jt808ID    = "jt808"
	testPhone  = "1001"
	baseFTPDir = "/tmp/ftp"
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
	h := server.New(
		server.WithALPN(true),
		server.WithHostPorts("0.0.0.0:8080"),
		server.WithHandleMethodNotAllowed(true),
	)
	goJt808 := service.New(
		service.WithHostPorts("0.0.0.0:808"),
		service.WithNetwork("tcp"),
		service.WithCustomTerminalEventer(func() service.TerminalEventer {
			return &meTerminal{} // 自定义终端事件 终端进入 离开 读写报文事件
		}),
	)
	go goJt808.Run()
	h.Use(func(c context.Context, ctx *app.RequestContext) {
		ctx.Set(jt808ID, goJt808)
	})
	group := h.Group("/api/v1/ftp/")
	{
		group.POST("/9205", P9205)
		group.POST("/9206", P9206)
	}
	go func() {
		time.Sleep(time.Second) // 等待服务启动完成
		client(testPhone)
	}()
	h.Spin()
}
