package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x0805 struct {
	BaseHandle
	// RespondSerialNumber 应答消息流水号
	RespondSerialNumber uint16 `json:"respondSerialNumber"`
	// Result 结果
	Result byte `json:"result"`
	// MultimediaIDNumber 多媒体个数
	MultimediaIDNumber uint16 `json:"multimediaIDNumber"`
	// MultimediaIDList 多媒体ID列表
	MultimediaIDList []uint32 `json:"multimediaIDList"`
}

func (t *T0x0805) Protocol() consts.JT808CommandType {
	return consts.T0805CameraShootImmediately
}

func (t *T0x0805) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 5 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.RespondSerialNumber = binary.BigEndian.Uint16(body[0:2])
	t.Result = body[2]
	t.MultimediaIDNumber = binary.BigEndian.Uint16(body[3:5])
	if len(body) != 5+int(t.MultimediaIDNumber)*4 {
		return protocol.ErrBodyLengthInconsistency
	}
	for i := 0; i < int(t.MultimediaIDNumber); i++ {
		start := 5 + i*4
		end := start + 4
		t.MultimediaIDList = append(t.MultimediaIDList, binary.BigEndian.Uint32(body[start:end]))
	}
	return nil
}

func (t *T0x0805) Encode() []byte {
	data := make([]byte, 5)
	binary.BigEndian.PutUint16(data[0:2], t.RespondSerialNumber)
	data[2] = t.Result
	binary.BigEndian.PutUint16(data[3:5], t.MultimediaIDNumber)
	for i := 0; i < len(t.MultimediaIDList); i++ {
		data = binary.BigEndian.AppendUint32(data, t.MultimediaIDList[i])
	}
	return data
}

func (t *T0x0805) HasReply() bool {
	return false
}

func (t *T0x0805) String() string {
	ids := "\t多媒体ID列表:"
	for _, v := range t.MultimediaIDList {
		ids += fmt.Sprintf("\n\t\t[%04x] 多媒体ID:[%d]", v, v)
	}
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", t.Protocol(), t.Encode()),
		fmt.Sprintf("\t[%04x] 应答消息流水号:[%d]", t.RespondSerialNumber, t.RespondSerialNumber),
		fmt.Sprintf("\t[%02x] 结果:[%d]", t.Result, t.Result),
		fmt.Sprintf("\t[%04x] 多媒体个数:[%d]", t.MultimediaIDNumber, t.MultimediaIDNumber),
		ids,
		"}",
	}, "\n")
}
