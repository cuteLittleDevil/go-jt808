package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x0104 struct {
	BaseHandle
	// RespondSerialNumber 应答消息流水号
	RespondSerialNumber uint16
	// RespondParamCount 应答参数个数
	RespondParamCount uint8
	// TerminalParamDetails 参数项列表 详情见表12 2019版新增
	TerminalParamDetails
}

func (t *T0x0104) Protocol() consts.JT808CommandType {
	return consts.T0104QueryParameter
}

func (t *T0x0104) ReplyProtocol() consts.JT808CommandType {
	return 0
}

func (t *T0x0104) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 3 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.RespondSerialNumber = binary.BigEndian.Uint16(body[:2])
	t.RespondParamCount = body[2]
	return t.TerminalParamDetails.parse(t.RespondParamCount, body[3:])
}

func (t *T0x0104) HasReply() bool {
	return false
}

func (t *T0x0104) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t[%04x] 应答消息流水号:[%d]", t.RespondSerialNumber, t.RespondSerialNumber),
		fmt.Sprintf("\t[%02x] 应答参数个数:[%d]", t.RespondParamCount, t.RespondParamCount),
		fmt.Sprintf("\t%s:\n", t.Protocol()),
		t.TerminalParamDetails.String(),
		"}",
	}, "\n")
}
