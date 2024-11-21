package main

import (
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"os"
)

var (
	address string
	phone   string
	goJt808 *service.GoJT808
)

func init() {
	flag.StringVar(&address, "address", "0.0.0.0:808", "监听的地址")
	flag.StringVar(&phone, "phone", "1001", "测试用的手机号")
	flag.Parse()
	fmt.Println("监听的地址:", address, "测试用的手机号:", phone)
	goJt808 = service.New(
		service.WithHostPorts(address),
		service.WithNetwork("tcp"),
		service.WithCustomTerminalEventer(func() service.TerminalEventer {
			return &meTerminal{} // 自定义终端 设备加入后 开始录像和拍照
		}),
		service.WithHasSubcontract(false), // 不过滤分包 则每一个分包都会触发回复 需要自己去控制
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				consts.T0801MultimediaDataUpload: &meT0801{phone: phone}, // 自定义0801多媒体的处理
			}
		}),
	)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(logger)
}

func main() {
	goJt808.Run()
}
