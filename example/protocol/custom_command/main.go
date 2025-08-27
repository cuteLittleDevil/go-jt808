package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
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

var (
	address string
	phone   string
	goJt808 *service.GoJT808
)

const (
	CustomPlatformCommand = 0x6666
	CustomTerminalCommand = 0x6667
)

func init() {
	flag.StringVar(&address, "address", "0.0.0.0:808", "监听的地址")
	flag.StringVar(&phone, "phone", "1001", "测试用的手机号")
	flag.Parse()
	fmt.Println("监听的地址:", address, "测试用的手机号:", phone)
	goJt808 = service.New(
		service.WithHostPorts(address),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			// 如果fork修改 service服务 connection.go的onActiveRespondEvent 匹配的话就是成功的
			// 自定义的话都是匹配不成功 就会超时
			platform := &CustomPlatform{}
			return map[consts.JT808CommandType]service.Handler{
				CustomPlatformCommand: platform,
				CustomTerminalCommand: &CustomTerminalReply{Handle: platform},

				// 这是默认添加匹配的指令 就是成功的情况
				consts.P9101RealTimeAudioVideoRequest: &me9101{},
				consts.T0001GeneralRespond:            &me0001{},
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
	go client(phone, "127.0.0.1:808") // 模拟一个设备连接

	go func() {
		time.Sleep(3 * time.Second)
		platform := CustomPlatform{}
		send(platform.Protocol(), platform.Encode())
		time.Sleep(1 * time.Second)

		p9101 := model.P0x9101{
			ServerIPLen:  byte(len(address)),
			ServerIPAddr: address,
			TcpPort:      1078,
			UdpPort:      0,
			ChannelNo:    1,
			DataType:     0, //  0-音视频 1-视频 2-双向对讲 3-监听 4-中心广播 5-透传
			StreamType:   0,
		}
		send(p9101.Protocol(), p9101.Encode())

	}()
	goJt808.Run()
}

func send(command consts.JT808CommandType, body []byte) {
	msg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              phone,
		Command:          command,
		Body:             body,
		OverTimeDuration: 3 * time.Second,
	})
	// 如果fork修改 service服务 connection.go的onActiveRespondEvent 匹配的话就是成功的
	// 自定义的话都是匹配不成功 就会超时
	if err := msg.ExtensionFields.Err; err == nil {
		fmt.Println("结果", msg.JTMessage.Header.String())
	} else if errors.Is(err, service.ErrWriteDataOverTime) {
		fmt.Println("超时", err)
	} else {
		fmt.Println("其他异常", err)
	}
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

	_, _ = conn.Write(register)
	_, _ = conn.Write(location)

	data := make([]byte, 1023)
	for {
		if n, _ := conn.Read(data); n > 0 {
			msg := fmt.Sprintf("%x", data[:n])
			if strings.HasPrefix(msg, "7e6666") {
				jtMsg := jt808.NewJTMessage()
				_ = jtMsg.Decode(data[:n])
				custom := &CustomTerminalReply{
					BaseHandle: model.BaseHandle{},
					Data:       jtMsg.Header.SerialNumber,
				}
				jtMsg.Header.ReplyID = uint16(custom.Protocol())
				jtMsg.Header.PlatformSerialNumber = 3
				_, _ = conn.Write(jtMsg.Header.Encode(custom.Encode()))
			} else if strings.HasPrefix(msg, "7e9101") {
				jtMsg := jt808.NewJTMessage()
				_ = jtMsg.Decode(data[:n])
				t0x0001 := model.T0x0001{
					BaseHandle:   model.BaseHandle{},
					SerialNumber: jtMsg.Header.SerialNumber,
					ID:           jtMsg.Header.ID,
					Result:       0,
				}
				jtMsg.Header.ReplyID = uint16(t0x0001.Protocol())
				jtMsg.Header.PlatformSerialNumber = 4
				_, _ = conn.Write(jtMsg.Header.Encode(t0x0001.Encode()))
			}
		}
	}
}

type me9101 struct {
	model.P0x9101
}

func (m *me9101) OnReadExecutionEvent(msg *service.Message) {
	fmt.Println("------- 9101 read event")
}

func (m *me9101) OnWriteExecutionEvent(msg service.Message) {
	fmt.Println("------- 9101 write event")
}

type me0001 struct {
	model.T0x0001
}

func (m *me0001) OnReadExecutionEvent(msg *service.Message) {
	fmt.Println("------- 0001 read event")
}

func (m *me0001) OnWriteExecutionEvent(msg service.Message) {
	fmt.Println("------- 0001 write event")
}
