package shared

import (
	"encoding/json"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
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

type Data struct {
	ID      string          `json:"id"`
	Type    int             `json:"type"`
	Subject string          `json:"subject"`
	Key     string          `json:"key"`
	Message service.Message `json:"message"`
}

func NewData(ID string, Type int, message service.Message) *Data {
	tmp := &Data{
		ID:      ID,
		Type:    Type,
		Key:     message.Header.TerminalPhoneNo,
		Message: message,
	}
	tmp.Subject = tmp.createSubject()
	return tmp
}

func (d *Data) ToBytes() []byte {
	b, _ := json.Marshal(d)
	return b
}

func (d *Data) Parse(data []byte) error {
	return json.Unmarshal(data, d)
}

func (d *Data) createSubject() string {
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
	return fmt.Sprintf("%s.%s.%d", prefix, d.ID, uint16(d.Message.Command))
}

type Notice struct {
	Key              string                  `json:"key"`
	Command          consts.JT808CommandType `json:"command"`
	Body             []byte                  `json:"body"`
	OverTimeDuration time.Duration           `json:"overTimeDuration"`
	UUID             string                  `json:"uuid"`
}

func (n *Notice) ToBytes() []byte {
	b, _ := json.Marshal(n)
	return b
}

func (n *Notice) Parse(data []byte) error {
	return json.Unmarshal(data, n)
}

func (n *Notice) Subject() string {
	return NoticeSubjectPrefix + "." + n.UUID
}

func (n *Notice) ReplySubject() string {
	return NoticeCompleteSubjectPrefix + "." + n.UUID
}
