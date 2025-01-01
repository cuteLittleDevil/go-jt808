package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x0201 struct {
	BaseHandle
	// RespondSerialNumber 应答消息流水号
	RespondSerialNumber uint16
	// T0x0200LocationItem 位置等信息
	T0x0200LocationItem
}

func (t *T0x0201) Protocol() consts.JT808CommandType {
	return consts.T0201QueryLocation
}

func (t *T0x0201) ReplyProtocol() consts.JT808CommandType {
	return 0
}

func (t *T0x0201) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 2 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.RespondSerialNumber = binary.BigEndian.Uint16(body[:2])
	return t.T0x0200LocationItem.parse(body[2:])
}

func (t *T0x0201) Encode() []byte {
	data := make([]byte, 0, 30)
	data = binary.BigEndian.AppendUint16(data, t.RespondSerialNumber)
	data = append(data, t.T0x0200LocationItem.encode()...)
	return data
}

func (t *T0x0201) HasReply() bool {
	return false
}

func (t *T0x0201) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", t.Protocol(), t.Encode()),
		fmt.Sprintf("\t[%02x] 应答消息流水号:[%d]", t.RespondSerialNumber, t.RespondSerialNumber),
		t.T0x0200LocationItem.String(),
		"}",
	}, "\n")
}
