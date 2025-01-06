package main

import (
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"log/slog"
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
	var (
		config    string
		videoPort int
		port      int
		ip        string
		phone     string
		dataPath  string
	)
	flag.StringVar(&config, "config", "./conf/lalserver.conf.json", "流媒体服务的配置文件")
	flag.IntVar(&videoPort, "videoPort", 1078, "1078服务端口")
	flag.IntVar(&port, "port", 808, "808服务端口")
	flag.StringVar(&ip, "ip", "", "对外发送的1078的ip")
	flag.StringVar(&phone, "phone", "", "测试的手机号 不为空的时候 开启一个模拟客户端")
	flag.StringVar(&dataPath, "dataPath", "../data/data.txt", "数据存放目录")
	flag.Parse()

	if ip != "" {
		var goJt808 *service.GoJT808
		goJt808 = service.New(
			service.WithHostPorts(fmt.Sprintf("0.0.0.0:%d", port)),
			service.WithNetwork("tcp"),
			service.WithCustomTerminalEventer(func() service.TerminalEventer {
				return &meTerminal{
					goJt808: goJt808,
					ip:      ip,
				}
			}),
		)
		go goJt808.Run()
	}

	if phone != "" {
		go client(port, videoPort, phone, dataPath)
	}

	goJt1078 := newLal1078(fmt.Sprintf("0.0.0.0:%d", videoPort), ip, config)
	goJt1078.run()
}

type meTerminal struct {
	goJt808 *service.GoJT808
	ip      string
}

func (t *meTerminal) OnJoinEvent(_ *service.Message, key string, _ error) {
	fmt.Printf("加入[%s] 3秒后发送9101\n", key)
	go func() {
		time.Sleep(3 * time.Second)
		p9101 := model.P0x9101{
			ServerIPLen:  byte(len(t.ip)),
			ServerIPAddr: t.ip,
			TcpPort:      1078,
			UdpPort:      0,
			ChannelNo:    1,
			DataType:     0, // 音视频
			StreamType:   0, // 主码流
		}
		body := p9101.Encode()
		activeMsg := service.NewActiveMessage(key, p9101.Protocol(), body, 3*time.Second)
		msg := t.goJt808.SendActiveMessage(activeMsg)
		var t0x0001 model.T0x0001
		if msg.ExtensionFields.Err != nil {
			panic(msg.ExtensionFields.Err)
		}
		if err := t0x0001.Parse(msg.JTMessage); err != nil {
			panic(err)
		}
		fmt.Println("下发的数据", fmt.Sprintf("%x", msg.ExtensionFields.PlatformData))
		{
			var m9101 model.P0x9101
			jtMsg := jt808.NewJTMessage()
			_ = jtMsg.Decode(msg.ExtensionFields.PlatformData)
			if err := m9101.Parse(jtMsg); err == nil {
				fmt.Println("平台:", m9101.String())
			}
		}

		fmt.Println("终端:", t0x0001.String())
	}()
}

func (t *meTerminal) OnLeaveEvent(key string) {
	fmt.Println("退出终端", key)
}

func (t *meTerminal) OnNotSupportedEvent(msg *service.Message) {}

func (t *meTerminal) OnReadExecutionEvent(msg *service.Message) {}

func (t *meTerminal) OnWriteExecutionEvent(msg service.Message) {}
