package model

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x0200 struct {
	BaseHandle
	// T0x0200LocationItem 位置等信息
	T0x0200LocationItem
	// T0x0200AdditionDetails 附加信息
	T0x0200AdditionDetails
}

func (t *T0x0200) Protocol() consts.JT808CommandType {
	return consts.T0200LocationReport
}

func (t *T0x0200) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if err := t.T0x0200LocationItem.parse(body); err != nil {
		return err
	}
	if len(body) > 28 {
		return t.T0x0200AdditionDetails.parse(body[28:])
	}
	return nil
}

func (t *T0x0200) Encode() []byte {
	return t.T0x0200LocationItem.encode()
}

func (t *T0x0200) String() string {
	body := t.Encode()
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", t.Protocol(), body),
		t.T0x0200LocationItem.String(),
		"}",
	}, "\n")
}
