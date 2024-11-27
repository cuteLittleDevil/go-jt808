package main

import (
	"encoding/hex"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"net"
	"os"
	"time"
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
	goJt808 := service.New(
		service.WithHostPorts("0.0.0.0:8080"),
		service.WithNetwork("tcp"),
		service.WithCustomTerminalEventer(func() service.TerminalEventer {
			return &meTerminal{} // 自定义开始 结束 报文处理等事件
		}),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				//consts.T0200LocationReport: &Location{},
			}
		}),
	)
	go goJt808.Run()

	time.Sleep(time.Second)
	conn, _ := net.Dial("tcp", "127.0.0.1:8080")
	msg := "7E0200007B0123456789017FFF000004000000080006EEB6AD02633DF701380003006320070719235901040000000B02020016030200210402002C051E37373700000000000000000000000000000000000000000000000000000011010012064D0000004D4D1307000000580058582504000000632A02000A2B040000001430011E3101283301207A7E"
	data, _ := hex.DecodeString(msg)
	_, _ = conn.Write(data)

	ticker := time.NewTicker(3 * time.Second)
	for range ticker.C {
		_, _ = conn.Write(data)
	}
}
