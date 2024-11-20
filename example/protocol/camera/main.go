package main

import (
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"os"
	"time"
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
	fmt.Println("监听的地址:", address, "测试用的手机号", phone)
	goJt808 = service.New(
		service.WithHostPorts(address),
		service.WithNetwork("tcp"),
		service.WithCustomTerminalEventer(func() service.TerminalEventer {
			return &meTerminal{}
		}),
		service.WithHasSubcontract(false),
	)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(logger)
}

type meTerminal struct{}

func (m *meTerminal) OnJoinEvent(msg *service.Message, key string, err error) {
	if key == phone {
		fmt.Println("加入", key, err, fmt.Sprintf("%x", msg.ExtensionFields.TerminalData))
		go func() {
			p8801 := model.P0x8801{
				ChannelID:                1,
				ShootCommand:             0xFFFF,
				PhotoIntervalOrVideoTime: 0,
				SaveFlag:                 0,
				Resolution:               1,
				VideoQuality:             10,
				Intensity:                255,
				Contrast:                 100,
				Saturation:               100,
				Chroma:                   100,
			}
			activeMsg := &service.ActiveMessage{
				Key:              key,
				Command:          p8801.Protocol(),
				Body:             p8801.Encode(),
				OverTimeDuration: 3 * time.Second,
			}
			fmt.Println("开始录像", p8801.String())
			replyMsg := goJt808.SendActiveMessage(activeMsg)
			fmt.Println("录像的回复", replyMsg.ExtensionFields.Err)
			time.Sleep(10 * time.Second)
			p8801.ShootCommand = 0
			fmt.Println("停止录像", p8801.String())
			activeMsg.Body = p8801.Encode()
			replyMsg = goJt808.SendActiveMessage(activeMsg)
			fmt.Println("停止录像的回复", replyMsg.ExtensionFields.Err)

		}()
		go func() {
			p8801 := model.P0x8801{
				ChannelID:                1,
				ShootCommand:             10,
				PhotoIntervalOrVideoTime: 1,
				SaveFlag:                 0,
				Resolution:               1,
				VideoQuality:             10,
				Intensity:                255,
				Contrast:                 100,
				Saturation:               100,
				Chroma:                   100,
			}
			activeMsg := &service.ActiveMessage{
				Key:              key,
				Command:          p8801.Protocol(),
				Body:             p8801.Encode(),
				OverTimeDuration: 3 * time.Second,
			}
			fmt.Println("开始循环拍照", p8801.String())
			replyMsg := goJt808.SendActiveMessage(activeMsg)
			fmt.Println("拍照的回复", replyMsg.ExtensionFields.Err)
		}()
	}
}

func (m *meTerminal) OnLeaveEvent(key string) {
	if key == phone {
		fmt.Println("离开", key)
	}
}

func (m *meTerminal) OnNotSupportedEvent(msg *service.Message) {
	if msg.Header.TerminalPhoneNo == phone {
		fmt.Println("未实现的报文", fmt.Sprintf("%x", msg.ExtensionFields.TerminalData))
	}
}

func (m *meTerminal) OnReadExecutionEvent(msg *service.Message) {
	switch msg.Command {
	case consts.T0801MultimediaDataUpload:
		var t0801 model.T0x0801
		_ = t0801.Parse(msg.JTMessage)
		fmt.Println(fmt.Sprintf("传输过程中[%d/%d]", msg.Header.SubPackageNo, msg.Header.SubPackageSum))
		fmt.Println("包情况", t0801.String())
		if msg.ExtensionFields.SubcontractComplete {
			format := ".jpg"
			switch t0801.MultimediaFormatEncode {
			case 0:
				format = ".jpeg"
			case 1:
				format = ".tlf"
			}
			name := fmt.Sprintf("./%d%s", t0801.MultimediaID, format)
			fmt.Println("包完成", name, t0801.String())
			_ = os.WriteFile(name, t0801.MultimediaPackage, os.ModePerm)
		}
	}
}

func (m *meTerminal) OnWriteExecutionEvent(msg service.Message) {
	if msg.Header.TerminalPhoneNo != phone {
		return
	}
	extension := msg.ExtensionFields
	if extension.ActiveSend {
		fmt.Println("平台主动下发的", fmt.Sprintf("%d %x", extension.PlatformSeq, extension.PlatformData))
		fmt.Println("完成回复的", fmt.Sprintf("%d %d", extension.TerminalSeq, len(extension.TerminalData)))
		return
	}
}

func main() {
	goJt808.Run()
}
