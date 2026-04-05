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
		// 设置终端读取超时. 默认不设置.
		service.WithTerminalTimeout(30*time.Second, func(t service.TerminalTimeout) {
			fmt.Println(fmt.Sprintf("key=[%s] addr[%s] 首次报文时间[%s] 最后一次报文时间[%s] 运行时间[%v]",
				t.Key, t.Address, t.FirstPacketTime.Format(time.DateTime),
				t.LastPacketTime.Format(time.DateTime), time.Since(t.ConnectionStartTime)))
		}),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				// 入门例子参考 https://github.com/cuteLittleDevil/go-jt808/blob/main/example/web/service/main.go
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
	register := t.CreateDefaultCommandData(consts.T0100Register)
	location := t.CreateDefaultCommandData(consts.T0200LocationReport)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	time.Sleep(1 * time.Second)
	_, _ = conn.Write(register)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	overtime := time.Now().Add(33 * time.Second)
	for range ticker.C {
		if time.Now().Before(overtime) {
			_, _ = conn.Write(location)
		}
	}
}
