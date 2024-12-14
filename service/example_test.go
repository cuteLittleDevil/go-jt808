package service

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

func Example() {
	goJt808 := New(
		// 服务的地址
		WithHostPorts("0.0.0.0:808"),
		// 服务的协议
		WithNetwork("tcp"),
		// 自定义终端key
		WithKeyFunc(func(message *Message) (string, bool) {
			return message.JTMessage.Header.TerminalPhoneNo, true
		}),
		// 是否过滤分包 默认过滤 这样读写事件就是完整的包
		WithHasSubcontract(true),
		// 自定义终端事件 终端进入 离开 读写报文事件
		WithCustomTerminalEventer(func() TerminalEventer {
			return &meTerminal{}
		}),
		// 自定义报文解析回复
		WithCustomHandleFunc(func() map[consts.JT808CommandType]Handler {
			return map[consts.JT808CommandType]Handler{
				consts.T0200LocationReport: &meLocation{},
			}
		}),
	)
	go goJt808.Run()

	// Output:
}

type meTerminal struct{}

func (m *meTerminal) OnJoinEvent(msg *Message, key string, err error) {
	fmt.Println("加入", key, err, fmt.Sprintf("%x", msg.ExtensionFields.TerminalData))
}

func (m *meTerminal) OnLeaveEvent(key string) {
	fmt.Println("离开", key)
}

func (m *meTerminal) OnNotSupportedEvent(msg *Message) {
	fmt.Println("未实现的指令", fmt.Sprintf("%x", msg.ExtensionFields.TerminalData))
}

func (m *meTerminal) OnReadExecutionEvent(msg *Message) {
	fmt.Println(fmt.Sprintf("读取: %s: [%x]", msg.Command, msg.ExtensionFields.TerminalData))
}

func (m *meTerminal) OnWriteExecutionEvent(msg Message) {
	fmt.Println(fmt.Sprintf("回复: %s: [%x]", msg.Command, msg.ExtensionFields.TerminalData))
}

type meLocation struct {
	model.T0x0200
}

func (ml *meLocation) OnReadExecutionEvent(msg *Message) {
	_ = ml.Parse(msg.JTMessage)
	fmt.Println("读到位置信息", ml.String())
}

func (ml *meLocation) OnWriteExecutionEvent(msg Message) {
	fmt.Println(fmt.Sprintf("读取位置的回复: %s: [%x]", msg.Command, msg.ExtensionFields.TerminalData))
}
