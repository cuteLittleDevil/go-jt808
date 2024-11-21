package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"time"
)

type meTerminal struct {
}

func (m *meTerminal) OnJoinEvent(msg *service.Message, key string, err error) {
	if key == phone {
		fmt.Println("加入", key, err, fmt.Sprintf("%x", msg.ExtensionFields.TerminalData))
		go func() {
			time.Sleep(500 * time.Millisecond) // 保证先回复默认的指令
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
			if replyMsg.ExtensionFields.Err == nil {
				fmt.Println("录像的回复", fmt.Sprintf("%x", replyMsg.ExtensionFields.TerminalData))
			}
			time.Sleep(5 * time.Second)
			p8801.ShootCommand = 0
			fmt.Println("停止录像", p8801.String())
			activeMsg.Body = p8801.Encode()
			replyMsg = goJt808.SendActiveMessage(activeMsg)
			if replyMsg.ExtensionFields.Err == nil {
				fmt.Println("停止录像的回复", fmt.Sprintf("%x", replyMsg.ExtensionFields.TerminalData))
			}
		}()
		go func() {
			time.Sleep(700 * time.Millisecond) // 保证先回复默认的指令
			p8801 := model.P0x8801{
				ChannelID:                1,
				ShootCommand:             3,
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
			if replyMsg.ExtensionFields.Err == nil {
				fmt.Println("拍照的回复", fmt.Sprintf("%x", replyMsg.ExtensionFields.TerminalData))
			}
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

func (m *meTerminal) OnReadExecutionEvent(_ *service.Message) {

}

func (m *meTerminal) OnWriteExecutionEvent(msg service.Message) {
	if msg.Header.TerminalPhoneNo != phone {
		return
	}
	extension := msg.ExtensionFields
	if extension.ActiveSend {
		fmt.Println("平台主动下发的", fmt.Sprintf("%d %x",
			extension.PlatformSeq, extension.PlatformData))
		fmt.Println("完成回复的", fmt.Sprintf("%d %d %s",
			extension.TerminalSeq, len(extension.TerminalData), msg.Command))
		return
	}
}
