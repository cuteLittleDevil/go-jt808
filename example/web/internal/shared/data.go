package shared

import (
	"encoding/json"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

const (
	InitSubjectPrefix         = "init"
	LeaveSubjectPrefix        = "leave"
	NotSupportedSubjectPrefix = "not-supported"
	ReadSubjectPrefix         = "read"
	WriteSubjectPrefix        = "write"
)

const (
	OnInit = iota + 1
	OnLeave
	OnNotSupported
	OnRead
	OnWrite
)

type (
	EventData struct {
		ID              string           `json:"id"`
		HTTPPrefix      string           `json:"HTTPPrefix"`
		AttachIP        string           `json:"attachIP"`
		AttachPort      int              `json:"attachPort"`
		Type            int              `json:"type"`
		Key             string           `json:"key"`
		JTMessage       *jt808.JTMessage `json:"message"`
		Subject         string           `json:"subject"`
		ExtensionFields struct {
			// TerminalSeq 终端流水号
			TerminalSeq uint16 `json:"terminalSeq,omitempty"`
			// PlatformSeq 平台下发的流水号
			PlatformSeq uint16 `json:"platformSeq,omitempty"`
			// TerminalData 终端主动上传的数据 分包合并的情况是全部body合在一起
			TerminalData []byte `json:"terminalData"`
			// PlatformData 平台下发的数据
			PlatformData []byte `json:"platformData"`
			// ActiveSend 是否是平台主动下发的
			ActiveSend bool `json:"activeSend,omitempty"`
			// SubcontractComplete 分包情况是否最终完成了
			SubcontractComplete bool `json:"subcontractComplete,omitempty"`
			// CurrentCommand 当前的指令
			CurrentCommand consts.JT808CommandType `json:"currentCommand,omitempty"`
			// TerminalCommand 终端的指令
			TerminalCommand consts.JT808CommandType `json:"terminalCommand,omitempty"`
			// PlatformCommand 平台的指令
			PlatformCommand consts.JT808CommandType `json:"platformCommand,omitempty"`
			// Err 异常情况
			Err error `json:"err,omitempty"`
		}
	}
)

type EventDataOption struct {
	F func(o *EventData)
}

func NewEventData(Type int, key string, opts ...EventDataOption) *EventData {
	tmp := &EventData{
		Type: Type,
		Key:  key,
	}
	for _, op := range opts {
		op.F(tmp)
	}
	return tmp
}

func WithMessage(msg service.Message) EventDataOption {
	return EventDataOption{F: func(o *EventData) {
		o.JTMessage = msg.JTMessage
		ex := msg.ExtensionFields
		o.ExtensionFields = struct {
			// TerminalSeq 终端流水号
			TerminalSeq uint16 `json:"terminalSeq,omitempty"`
			// PlatformSeq 平台下发的流水号
			PlatformSeq uint16 `json:"platformSeq,omitempty"`
			// TerminalData 终端主动上传的数据 分包合并的情况是全部body合在一起
			TerminalData []byte `json:"terminalData"`
			// PlatformData 平台下发的数据
			PlatformData []byte `json:"platformData"`
			// ActiveSend 是否是平台主动下发的
			ActiveSend bool `json:"activeSend,omitempty"`
			// SubcontractComplete 分包情况是否最终完成了
			SubcontractComplete bool `json:"subcontractComplete,omitempty"`
			// CurrentCommand 当前的指令
			CurrentCommand consts.JT808CommandType `json:"currentCommand,omitempty"`
			// TerminalCommand 终端的指令
			TerminalCommand consts.JT808CommandType `json:"terminalCommand,omitempty"`
			// PlatformCommand 平台的指令
			PlatformCommand consts.JT808CommandType `json:"platformCommand,omitempty"`
			// Err 异常情况
			Err error `json:"err,omitempty"`
		}{
			TerminalSeq:         ex.TerminalSeq,
			PlatformSeq:         ex.PlatformSeq,
			TerminalData:        ex.TerminalData,
			PlatformData:        ex.PlatformData,
			ActiveSend:          ex.ActiveSend,
			SubcontractComplete: ex.SubcontractComplete,
			CurrentCommand:      msg.Command,
			TerminalCommand:     ex.TerminalCommand,
			PlatformCommand:     ex.PlatformCommand,
			Err:                 ex.Err,
		}
		o.Subject = o.createSubject(uint16(msg.Command))
	}}
}

func WithIDAndHTTPPrefix(id string, httpPrefix string) EventDataOption {
	return EventDataOption{F: func(o *EventData) {
		o.ID = id
		o.HTTPPrefix = httpPrefix
	}}
}

func WithAttachIPAndPort(attachIP string, attachPort int) EventDataOption {
	return EventDataOption{F: func(o *EventData) {
		o.AttachIP = attachIP
		o.AttachPort = attachPort
	}}
}

func (d *EventData) ToBytes() []byte {
	b, _ := json.Marshal(d)
	return b
}

func (d *EventData) Parse(data []byte) error {
	return json.Unmarshal(data, d)
}

func (d *EventData) createSubject(command uint16) string {
	prefix := ""
	switch d.Type {
	case OnInit:
		prefix = InitSubjectPrefix
	case OnLeave:
		prefix = LeaveSubjectPrefix
	case OnNotSupported:
		prefix = NotSupportedSubjectPrefix
	case OnRead:
		prefix = ReadSubjectPrefix
	case OnWrite:
		prefix = WriteSubjectPrefix
	}
	// 固定事件前缀.服务ID.手机号.报文类型
	sim := d.JTMessage.Header.TerminalPhoneNo
	return fmt.Sprintf("%s.%s.%s.%d", prefix, d.ID, sim, command)
}
