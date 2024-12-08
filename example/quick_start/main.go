package main

import (
	"github.com/cuteLittleDevil/go-jt808/attachment"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"os"
)

var goJt808 *service.GoJT808

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(logger)

	attach := attachment.New(
		attachment.WithNetwork("tcp"),
		attachment.WithHostPorts("0.0.0.0:10001"),
		attachment.WithActiveSafetyType(consts.ActiveSafetyJS), // 默认苏标 支持黑标 广东标 湖南标 四川标
		attachment.WithFileEventerFunc(func() attachment.FileEventer {
			return &meFileEvent{} // 自定义文件处理 开始 结束 当前进度 补传 完成等事件
		}),
	)
	go attach.Run()
}

func main() {
	goJt808 = service.New(
		service.WithHostPorts("0.0.0.0:808"),
		service.WithNetwork("tcp"),
		service.WithCustomTerminalEventer(func() service.TerminalEventer {
			return &meTerminal{} // 自定义终端事件 终端进入 离开 读写报文事件
		}),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				consts.T0200LocationReport: &meLocation{}, // 自定义0x0200位置解析等
			}
		}),
	)
	goJt808.Run()
}
