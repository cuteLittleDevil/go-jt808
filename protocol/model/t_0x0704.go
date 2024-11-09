package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type (
	T0x0704 struct {
		BaseHandle
		// Num 数据项个数 必须大于0
		Num uint16 `json:"num"`
		// LocationType 0-正常位置批量汇报 1-盲区补报
		LocationType byte `json:"locationType"`
		// Items 数据项
		Items []T0x0704LocationItem `json:"items"`
	}

	T0x0704LocationItem struct {
		// Len 长度 位置汇报数据体长度
		Len uint16 `json:"len"`
		// T0x0200LocationItem 位置汇报数据体
		T0x0200LocationItem
		// T0x0200AdditionDetails 附加信息
		T0x0200AdditionDetails
	}
)

func (t *T0x0704) Protocol() consts.JT808CommandType {
	return consts.T0704LocationBatchUpload
}

func (t *T0x0704) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 31 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.Num = binary.BigEndian.Uint16(body[:2])
	t.LocationType = body[2]
	start := 3
	for i := 0; i < int(t.Num); i++ {
		var item T0x0704LocationItem
		item.Len = binary.BigEndian.Uint16(body[start : start+2])
		if start+2+int(item.Len) > len(body) {
			return protocol.ErrBodyLengthInconsistency
		}
		curBody := body[start+2 : start+2+int(item.Len)]
		if err := item.T0x0200LocationItem.parse(curBody); err != nil {
			return err
		}
		if len(curBody) > 28 {
			if err := item.T0x0200AdditionDetails.parse(curBody[28:]); err != nil {
				return err
			}
		}
		t.Items = append(t.Items, item)
		start += 2 + int(item.Len)
	}
	return nil
}

func (t *T0x0704) Encode() []byte {
	data := make([]byte, 3, 100)
	binary.BigEndian.PutUint16(data[:2], t.Num)
	data[2] = t.LocationType
	for i := 0; i < len(t.Items); i++ {
		body := t.Items[i].encode()
		binary.BigEndian.AppendUint16(data, uint16(len(body)))
		data = append(data, body...)
	}
	return data
}

func (t *T0x0704) String() string {
	str := "数据体对象:{\n"
	str += fmt.Sprintf("\t%s:[%x]\n", t.Protocol(), t.Encode())
	str += fmt.Sprintf("\t[%04x] 数据项个数:[%d]\n", t.Num, t.Num)
	str += fmt.Sprintf("\t[%02x] 位置汇报类型:[%d] 0-正常位置批量汇报 1-盲区补报\n", t.LocationType, t.LocationType)
	str += fmt.Sprintf("\t位置汇报数据集合: [\n")
	for i := 0; i < len(t.Items); i++ {
		str += "\t{\n"
		itemStr := t.Items[i].T0x0200LocationItem.String()
		str += strings.ReplaceAll(itemStr, "\t", "\t\t")
		str += "\n\t}\n"
	}
	str += "\t]"
	return strings.Join([]string{
		str,
		"}",
	}, "\n")
}
