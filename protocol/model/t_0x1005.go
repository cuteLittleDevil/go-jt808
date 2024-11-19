package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x1005 struct {
	BaseHandle
	// StartTime 开始时间 BCD[6] CMT+8时区(上海时区)
	StartTime string `json:"startTime"`
	// EndTime 结束时间 BCD[6] CMT+8时区(上海时区)
	EndTime string `json:"endTime"`
	// BoardNumber 上车人数 从起始时间到结束时间的上车人数
	BoardNumber uint16 `json:"boardNumber"`
	// AlightNumber 下车人数 从起始时间到结束时间的上车人数
	AlightNumber uint16 `json:"alightNumber"`
}

func (t *T0x1005) Protocol() consts.JT808CommandType {
	return consts.T1005UploadPassengerFlow
}

func (t *T0x1005) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) != 16 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.StartTime = utils.BCD2Time(body[:6])
	t.EndTime = utils.BCD2Time(body[6:12])
	t.BoardNumber = binary.BigEndian.Uint16(body[12:14])
	t.AlightNumber = binary.BigEndian.Uint16(body[14:16])
	return nil
}

func (t *T0x1005) Encode() []byte {
	data := make([]byte, 16)
	startBcdTime := utils.Time2BCD(t.StartTime)
	copy(data[0:6], startBcdTime)
	endBcdTime := utils.Time2BCD(t.EndTime)
	copy(data[6:12], endBcdTime)
	binary.BigEndian.PutUint16(data[12:14], t.BoardNumber)
	binary.BigEndian.PutUint16(data[14:16], t.AlightNumber)
	return data
}

func (t *T0x1005) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", t.Protocol(), t.Encode()),
		fmt.Sprintf("\t[%x] 开始时间:[%s]", utils.Time2BCD(t.StartTime), t.StartTime),
		fmt.Sprintf("\t[%x] 结束时间:[%s]", utils.Time2BCD(t.EndTime), t.EndTime),
		fmt.Sprintf("\t[%04x] 上车人数:[%d]", t.BoardNumber, t.BoardNumber),
		fmt.Sprintf("\t[%04x] 下车人数:[%d]", t.AlightNumber, t.AlightNumber),
		"}",
	}, "\n")
}
