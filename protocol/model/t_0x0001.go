package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x0001 struct {
	BaseHandle
	SerialNumber uint16 `json:"serialNumber"`
	ID           uint16 `json:"id"`
	Result       byte   `json:"result"`
}

func (t *T0x0001) Protocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (t *T0x0001) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) != 5 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.SerialNumber = binary.BigEndian.Uint16(body[:2])
	t.ID = binary.BigEndian.Uint16(body[2:4])
	t.Result = body[4]
	return nil
}

func (t *T0x0001) Encode() []byte {
	data := make([]byte, 5)
	binary.BigEndian.PutUint16(data[:2], t.SerialNumber)
	binary.BigEndian.PutUint16(data[2:4], t.ID)
	data[4] = t.Result
	return data
}

func (t *T0x0001) String() string {
	str := "数据体对象:{\n"
	str += fmt.Sprintf("\t%s:[%10x]", consts.T0001GeneralRespond, t.Encode())
	return strings.Join([]string{
		str,
		fmt.Sprintf("\t[%04x] 应答流水号:[%d]", t.SerialNumber, t.SerialNumber),
		fmt.Sprintf("\t[%04x] 应答消息ID:[%d]", t.ID, t.ID),
		fmt.Sprintf("\t[%02x] 结果:[%d] 0-成功 1-失败 "+
			"2-消息有误 3-不支持 4-报警处理确认", t.Result, t.Result),
		"}",
	}, "\n")
}
