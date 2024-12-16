package shared

import (
	"encoding/json"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/google/uuid"
	"time"
)

const (
	InitSubjectPrefix           = "init"
	LeaveSubjectPrefix          = "leave"
	NotSupportedSubjectPrefix   = "not-supported"
	ReadSubjectPrefix           = "read"
	WriteSubjectPrefix          = "write"
	NoticeSubjectPrefix         = "notice"
	NoticeCompleteSubjectPrefix = "notice-complete"
)

const (
	OnInit = iota + 1
	OnLeave
	OnNotSupported
	OnRead
	OnWrite
	OnNotice
	OnNoticeComplete
)

type (
	EventData struct {
		ID           string          `json:"id"`
		Type         int             `json:"type"`
		Key          string          `json:"key"`
		Message      service.Message `json:"message"`
		Subject      string          `json:"subject"`
		Notice       Notice          `json:"notice"`
		ReplySubject string          `json:"replySubject"`
	}
	Notice struct {
		Key              string                  `json:"key"`
		Command          consts.JT808CommandType `json:"command"`
		Body             []byte                  `json:"body"`
		OverTimeDuration time.Duration           `json:"overTimeDuration"`
	}
)

type EventDataOption struct {
	F func(o *EventData)
}

func NewEventData(ID string, Type int, key string, opts ...EventDataOption) *EventData {
	tmp := &EventData{
		ID:   ID,
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
		o.Message = msg
		o.Subject = o.createSubject(uint16(msg.Command))
	}}
}

func WithNotice(n Notice) EventDataOption {
	return EventDataOption{F: func(o *EventData) {
		o.Notice = n
		o.ReplySubject = o.createSubject(uuid.New().String())
	}}
}

func (d *EventData) ToBytes() []byte {
	b, _ := json.Marshal(d)
	return b
}

func (d *EventData) Parse(data []byte) error {
	return json.Unmarshal(data, d)
}

func (d *EventData) createSubject(data any) string {
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
	case OnNotice:
		prefix = NoticeSubjectPrefix
	case OnNoticeComplete:
		prefix = NoticeCompleteSubjectPrefix
	}
	// 固定前缀-服务ID-手机号-自定义数据 （自定义数据 1-读到消息用指令 2-通知完成用uuid)
	return fmt.Sprintf("%s.%s.%s.%v", prefix, d.ID, d.Key, data)
}
