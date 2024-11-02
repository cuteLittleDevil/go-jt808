package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x1206 struct {
	BaseHandle
	// RespondSerialNumber 应答流水号 对应的平台文件上传消息的流水号
	RespondSerialNumber uint16 `json:"respondSerialNumber"`
	// Result 结果 0-成功 1-失败
	Result byte `json:"result"`
}

func (t *T0x1206) Protocol() consts.JT808CommandType {
	return consts.T1206FileUploadCompleteNotice
}

func (t *T0x1206) ReplyProtocol() consts.JT808CommandType {
	return 0
}

func (t *T0x1206) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) != 3 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.RespondSerialNumber = binary.BigEndian.Uint16(body[0:2])
	t.Result = body[2]
	return nil
}

func (t *T0x1206) Encode() []byte {
	data := make([]byte, 3)
	binary.BigEndian.PutUint16(data[0:2], t.RespondSerialNumber)
	data[2] = t.Result
	return data
}

func (t *T0x1206) HasReply() bool {
	return false
}

func (t *T0x1206) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", t.Protocol(), t.Encode()),
		fmt.Sprintf("\t[%04x] 应答流水号:[%d]", t.RespondSerialNumber, t.RespondSerialNumber),
		fmt.Sprintf("\t[%02x] 结果:[%d] 0-成功 1-失败", t.Result, t.Result),
		"}",
	}, "\n")
}
