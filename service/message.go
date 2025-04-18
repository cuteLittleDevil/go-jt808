package service

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type Message struct {
	*jt808.JTMessage
	Handler `json:"-"`
	// Key 唯一标识符 默认手机号 终端未加入的时候为空
	Key string `json:"key"`
	// Command 当前的指令类型
	Command         consts.JT808CommandType `json:"command"`
	ExtensionFields struct {
		// TerminalSeq 终端流水号
		TerminalSeq uint16 `json:"terminalSeq,omitempty"`
		// PlatformSeq 平台下发的流水号
		PlatformSeq uint16 `json:"platformSeq,omitempty"`
		// TerminalData 终端主动上传的数据 分包合并的情况是全部body合在一起
		TerminalData []byte `json:"terminalData,omitempty"`
		// PlatformData 平台下发的数据
		PlatformData []byte `json:"platformData,omitempty"`
		// ActiveSend 是否是平台主动下发的
		ActiveSend bool `json:"activeSend,omitempty"`
		// SubcontractComplete 分包情况是否最终完成了
		SubcontractComplete bool `json:"subcontractComplete,omitempty"`
		// TerminalCommand 终端的指令
		TerminalCommand consts.JT808CommandType `json:"terminalCommand,omitempty"`
		// PlatformCommand 平台的指令
		PlatformCommand consts.JT808CommandType `json:"platformCommand,omitempty"`
		// CustomData 自定义数据
		CustomData any `json:"customData,omitempty"`
		// Err 异常情况
		Err error `json:"err,omitempty"`
	}
}

func newTerminalMessage(jtMsg *jt808.JTMessage, terminalData []byte) *Message {
	return &Message{
		JTMessage: jtMsg,
		Command:   consts.JT808CommandType(jtMsg.Header.ID),
		ExtensionFields: struct {
			TerminalSeq         uint16                  `json:"terminalSeq,omitempty"`
			PlatformSeq         uint16                  `json:"platformSeq,omitempty"`
			TerminalData        []byte                  `json:"terminalData,omitempty"`
			PlatformData        []byte                  `json:"platformData,omitempty"`
			ActiveSend          bool                    `json:"activeSend,omitempty"`
			SubcontractComplete bool                    `json:"subcontractComplete,omitempty"`
			TerminalCommand     consts.JT808CommandType `json:"terminalCommand,omitempty"`
			PlatformCommand     consts.JT808CommandType `json:"platformCommand,omitempty"`
			CustomData          any                     `json:"customData,omitempty"`
			Err                 error                   `json:"err,omitempty"`
		}{
			TerminalData:    terminalData,
			TerminalCommand: consts.JT808CommandType(jtMsg.Header.ID),
			TerminalSeq:     jtMsg.Header.SerialNumber,
		},
	}
}

func newActiveMessage(seq uint16, command consts.JT808CommandType, platformData []byte, err error) *Message {
	jtMsg := jt808.NewJTMessage()
	_ = jtMsg.Decode(platformData)
	return &Message{
		JTMessage: jtMsg,
		Command:   command,
		ExtensionFields: struct {
			TerminalSeq         uint16                  `json:"terminalSeq,omitempty"`
			PlatformSeq         uint16                  `json:"platformSeq,omitempty"`
			TerminalData        []byte                  `json:"terminalData,omitempty"`
			PlatformData        []byte                  `json:"platformData,omitempty"`
			ActiveSend          bool                    `json:"activeSend,omitempty"`
			SubcontractComplete bool                    `json:"subcontractComplete,omitempty"`
			TerminalCommand     consts.JT808CommandType `json:"terminalCommand,omitempty"`
			PlatformCommand     consts.JT808CommandType `json:"platformCommand,omitempty"`
			CustomData          any                     `json:"customData,omitempty"`
			Err                 error                   `json:"err,omitempty"`
		}{
			PlatformSeq:     seq,
			PlatformData:    platformData,
			ActiveSend:      true,
			PlatformCommand: command,
			Err:             err,
		},
	}
}

func newErrMessage(err error) *Message {
	return &Message{ExtensionFields: struct {
		TerminalSeq         uint16                  `json:"terminalSeq,omitempty"`
		PlatformSeq         uint16                  `json:"platformSeq,omitempty"`
		TerminalData        []byte                  `json:"terminalData,omitempty"`
		PlatformData        []byte                  `json:"platformData,omitempty"`
		ActiveSend          bool                    `json:"activeSend,omitempty"`
		SubcontractComplete bool                    `json:"subcontractComplete,omitempty"`
		TerminalCommand     consts.JT808CommandType `json:"terminalCommand,omitempty"`
		PlatformCommand     consts.JT808CommandType `json:"platformCommand,omitempty"`
		CustomData          any                     `json:"customData,omitempty"`
		Err                 error                   `json:"err,omitempty"`
	}{
		Err: err,
	}}
}

func (msg *Message) hasComplete() bool {
	if msg.JTMessage.Header.SubPackageSum == 0 {
		return true
	}
	return msg.ExtensionFields.SubcontractComplete
}
