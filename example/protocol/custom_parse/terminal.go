package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type meTerminal struct {
}

func (m *meTerminal) OnJoinEvent(msg *service.Message, key string, err error) {
	fmt.Println("加入", msg.Header.String(), key, err)
}

func (m *meTerminal) OnLeaveEvent(key string) {
	fmt.Println("离开 key=默认手机号", key)
}

func (m *meTerminal) OnNotSupportedEvent(msg *service.Message) {
	fmt.Println(fmt.Sprintf("暂未实现的报文 %x", msg.ExtensionFields.TerminalData))
}

func (m *meTerminal) OnReadExecutionEvent(msg *service.Message) {
	if msg.Command != consts.T0200LocationReport {
		return
	}
	var tmp Location
	_ = tmp.Parse(msg.JTMessage)
	fmt.Println(tmp.T0x0200AdditionDetails.String())
	if v, ok := tmp.Additions[consts.A0x01Mile]; ok {
		fmt.Println(fmt.Sprintf("里程[%d] 自定义辅助里程[%d]", v.Content.Mile, tmp.customMile))
	}
	id := consts.JT808LocationAdditionType(0x33)
	if v, ok := tmp.Additions[id]; ok {
		fmt.Println("自定义未知信息扩展", v.Content.CustomValue, tmp.customValue)
	}
}

func (m *meTerminal) OnWriteExecutionEvent(msg service.Message) {
	extension := msg.ExtensionFields
	fmt.Println(fmt.Sprintf("设备上传的 %x", extension.TerminalData))
	fmt.Println(fmt.Sprintf("平台下发的 %x", extension.PlatformData))
}

type Location struct {
	model.T0x0200
	customMile  int
	customValue uint8
}

func (l *Location) Parse(jtMsg *jt808.JTMessage) error {
	l.T0x0200AdditionDetails.CustomAdditionContentFunc = func(id uint8, content []byte) (model.AdditionContent, bool) {
		if id == uint8(consts.A0x01Mile) {
			l.customMile = 100
		}
		if id == 0x33 {
			value := content[0]
			l.customValue = value
			return model.AdditionContent{
				Data:        content,
				CustomValue: value,
			}, true
		}
		return model.AdditionContent{}, false
	}
	return l.T0x0200.Parse(jtMsg)
}

func (l *Location) OnReadExecutionEvent(_ *service.Message) {

}

func (l *Location) OnWriteExecutionEvent(_ service.Message) {}
