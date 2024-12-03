package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"os"
	"time"
)

type meTerminal struct {
	file *os.File
}

func (m *meTerminal) OnJoinEvent(_ *service.Message, key string, err error) {
	str := fmt.Sprintf("终端加入: %s err[%v]", key, err)
	m.println(str)
	m.file, _ = os.OpenFile("terminal.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
}

func (m *meTerminal) OnLeaveEvent(key string) {
	str := fmt.Sprintf("终端离开: %s", key)
	m.println(str)
}

func (m *meTerminal) OnNotSupportedEvent(msg *service.Message) {
	str := fmt.Sprintf("暂未实现的报文: %s: [%x]", msg.Command, msg.ExtensionFields.TerminalData)
	m.println(str)
}

func (m *meTerminal) OnReadExecutionEvent(msg *service.Message) {
	str := fmt.Sprintf("读到的报文 %s: [%x]", msg.Command, msg.ExtensionFields.TerminalData)
	m.println(str)
}

func (m *meTerminal) OnWriteExecutionEvent(msg service.Message) {
	str := "回复的报文"
	if msg.ExtensionFields.ActiveSend {
		str = "主动发送的报文"
	}
	str += fmt.Sprintf(" %s: [%x]", msg.ExtensionFields.PlatformCommand, msg.ExtensionFields.PlatformData)
	m.println(str)

	if msg.Command == consts.T0200LocationReport {
		go m.onAlarmEvent(msg.JTMessage)
	}
}

func (m *meTerminal) println(str string) {
	str += "\n"
	fmt.Println(str)
	if m.file != nil {
		_, _ = m.file.WriteString(str)
		_ = m.file.Sync()
	}
}

func (m *meTerminal) onAlarmEvent(jtMsg *jt808.JTMessage) {
	key := jtMsg.Header.TerminalPhoneNo
	var location meLocation
	if location.Parse(jtMsg) == nil {
		if location.T0x0200AdditionExtension0x64.ParseSuccess {
			str := fmt.Sprintf("0x64附件信息: %s", location.T0x0200AdditionExtension0x64.String())
			m.println(str)
			m.send9208("me_64", key, location.T0x0200AdditionExtension0x64.P9208AlarmSign)
		}

		if location.T0x0200AdditionExtension0x65.ParseSuccess {
			str := fmt.Sprintf("0x65附件信息: %s", location.T0x0200AdditionExtension0x65.String())
			m.println(str)
			m.send9208("me_65", key, location.T0x0200AdditionExtension0x65.P9208AlarmSign)
		}

		if location.T0x0200AdditionExtension0x66.ParseSuccess {
			str := fmt.Sprintf("0x66附件信息: %s", location.T0x0200AdditionExtension0x66.String())
			m.println(str)
			m.send9208("me_66", key, location.T0x0200AdditionExtension0x66.P9208AlarmSign)
		}

		if location.T0x0200AdditionExtension0x67.ParseSuccess {
			str := fmt.Sprintf("0x67附件信息: %s", location.T0x0200AdditionExtension0x67.String())
			m.println(str)
			m.send9208("me_67", key, location.T0x0200AdditionExtension0x67.P9208AlarmSign)
		}

		if location.T0x0200AdditionExtension0x70.ParseSuccess {
			str := fmt.Sprintf("0x70附件信息: %s", location.T0x0200AdditionExtension0x70.String())
			m.println(str)
			m.send9208("me_70", key, location.T0x0200AdditionExtension0x70.P9208AlarmSign)
		}
	}
}

func (m *meTerminal) send9208(alarmID string, key string, p9208AlarmSign model.P9208AlarmSign) {
	const (
		attachIP   = "127.0.0.1"
		attachPort = 10001
	)
	p9208 := model.P0x9208{
		ServerIPLen:    byte(len(attachIP)),
		ServerAddr:     attachIP,
		TcpPort:        uint16(attachPort),
		UdpPort:        0,
		P9208AlarmSign: p9208AlarmSign,
		AlarmID:        alarmID,
		Reserve:        make([]byte, 16),
	}
	replyMsg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              key,
		Command:          p9208.Protocol(),
		Body:             p9208.Encode(),
		OverTimeDuration: 5 * time.Second,
	})
	if replyMsg.ExtensionFields.Err == nil {
		m.println(fmt.Sprintf("9208回复成功 %x", replyMsg.ExtensionFields.TerminalData))
	} else {
		m.println(fmt.Sprintf("9208回复失败 %s", replyMsg.ExtensionFields.Err))
	}
}