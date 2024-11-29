package main

import (
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/attachment"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"os"
)

var (
	jt808Addr  string
	attachIP   string
	attachPort int
	goJt808    *service.GoJT808
)

func init() {
	flag.StringVar(&jt808Addr, "jt808Addr", "0.0.0.0:808", "jt808服务地址")
	flag.StringVar(&attachIP, "attachIP", "127.0.0.1", "主动安全服务IP")
	flag.IntVar(&attachPort, "attachPort", 17017, "主动安全服务端口")
	flag.Parse()

	fmt.Println("808地址", jt808Addr, "文件服务地址", fmt.Sprintf("%s:%d", attachIP, attachPort))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(logger)

	goJt808 = service.New(
		service.WithHostPorts(jt808Addr),
		service.WithNetwork("tcp"),
		service.WithCustomTerminalEventer(func() service.TerminalEventer {
			f, _ := os.OpenFile("jt808指令.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
			return newMeTerminal(f)
		}),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				consts.P9208AlarmAttachUpload: &me0x9208{},
			}
		}),
	)
	go goJt808.Run()
}

func main() {
	attach := attachment.New(
		attachment.WithNetwork("tcp"),
		attachment.WithHostPorts(fmt.Sprintf("%s:%d", attachIP, attachPort)),
	)
	attach.Run()
}
