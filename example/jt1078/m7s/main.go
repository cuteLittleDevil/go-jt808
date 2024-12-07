package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	_ "github.com/cuteLittleDevil/m7s-jt1078/v5"
	_ "m7s.live/v5/plugin/mp4" // v5版本目前还没有发布 需要拉到本地 使用go work
	_ "m7s.live/v5/plugin/preview"
	"time"
)

var (
	_ip           = "127.0.0.1"
	_realTimePort = 10012
	goJt808       *service.GoJT808
)

func init() {
	flag.StringVar(&_ip, "ip", "127.0.0.1", "实时视频 ip地址")
	flag.Parse()

	goJt808 = service.New(
		service.WithHostPorts("0.0.0.0:8085"),
		service.WithNetwork("tcp"),
		service.WithCustomTerminalEventer(func() service.TerminalEventer {
			return &terminal{}
		}),
	)
	go goJt808.Run()
}

func main() {
	ctx := context.Background()
	// 使用自定义模拟器推流 读取本地文件的
	fmt.Println("preview", "http://127.0.0.1:8088/preview")
	_ = m7s.Run(ctx, "./config.yaml")
}

type terminal struct {
}

func (t *terminal) OnJoinEvent(_ *service.Message, key string, err error) {
	if err == nil {
		go func() {
			time.Sleep(3 * time.Second)
			p9101 := model.P0x9101{
				ServerIPLen:  byte(len(_ip)),
				ServerIPAddr: _ip,
				TcpPort:      uint16(_realTimePort),
				UdpPort:      0,
				ChannelNo:    1,
				DataType:     0,
				StreamType:   1,
			}
			body := p9101.Encode()
			fmt.Println(time.Now().Format(time.DateTime), "发送 9101指令 key=", key, p9101.String())
			activeMsg := service.NewActiveMessage(key, p9101.Protocol(), body, 3*time.Second)
			msg := goJt808.SendActiveMessage(activeMsg)
			var t0x0001 model.T0x0001
			if msg.ExtensionFields.Err != nil {
				panic(msg.ExtensionFields.Err)
			}
			if err := t0x0001.Parse(msg.JTMessage); err != nil {
				panic(err)
			}
			fmt.Println(t0x0001.String())
		}()
	}
}

func (t *terminal) OnLeaveEvent(_ string) {}

func (t *terminal) OnNotSupportedEvent(_ *service.Message) {}

func (t *terminal) OnReadExecutionEvent(_ *service.Message) {}

func (t *terminal) OnWriteExecutionEvent(_ service.Message) {}
