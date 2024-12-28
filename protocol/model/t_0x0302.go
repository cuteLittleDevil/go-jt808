package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x0302 struct {
	BaseHandle
	// SerialNumber 应答流水号 对应的平台消息的流水号
	SerialNumber uint16 `json:"serialNumber"`
	// AnswerID 答案ID
	AnswerID byte `json:"id"`
}

func (t *T0x0302) Protocol() consts.JT808CommandType {
	return consts.T0302QuestionAnswer
}

func (t *T0x0302) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) != 3 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.SerialNumber = binary.BigEndian.Uint16(body[:2])
	t.AnswerID = body[2]
	return nil
}

func (t *T0x0302) Encode() []byte {
	data := make([]byte, 3)
	binary.BigEndian.PutUint16(data[:2], t.SerialNumber)
	data[2] = t.AnswerID
	return data
}

func (t *T0x0302) HasReply() bool {
	return false
}

func (t *T0x0302) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", t.Protocol(), t.Encode()),
		fmt.Sprintf("\t[%04x] 应答流水号:[%d]", t.SerialNumber, t.SerialNumber),
		fmt.Sprintf("\t[%04x] 答案ID:[%d]", t.AnswerID, t.AnswerID),
		"}",
	}, "\n")
}
