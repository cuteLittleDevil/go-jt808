package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"log/slog"
	"net"
	"os"
	"strings"
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
		service.WithHostPorts("0.0.0.0:808"),
		//service.WithCustomTerminalEventer(func() service.TerminalEventer {
		//	// 自定义终端事件 包括加入 离开 读取报文等
		//}),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				consts.T0200LocationReport: &meLocation{}, // 自定义0x0200位置解析等
			}
		}),
	)
	go client("1001", "127.0.0.1:808") // 模拟一个设备连接
	goJt808.Run()
}

type meLocation struct {
	model.T0x0200
}

func (l *meLocation) OnReadExecutionEvent(msg *service.Message) {
	_ = l.Parse(msg.JTMessage)
	fmt.Println(time.Now().Format(time.DateTime), l.String()) // 打印经纬度等信息
}

func (l *meLocation) OnWriteExecutionEvent(_ service.Message) {}

func (l *meLocation) String() string {
	body := l.T0x0200.Encode()
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", l.Protocol(), body),
		l.T0x0200LocationItem.String(),
		l.AlarmSignDetails.String(),
		l.StatusSignDetails.String(),
		"}",
	}, "\n")
}

func client(phone string, address string) {
	time.Sleep(time.Second)
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, phone))
	location := t.CreateDefaultCommandData(consts.T0200LocationReport)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		_, _ = conn.Write(location)
	}
}
