package model

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x0200 struct {
	BaseHandle
	T0x0200LocationItem
}

func (t *T0x0200) Protocol() uint16 {
	return uint16(consts.T0200LocationReport)
}

func (t *T0x0200) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	return t.parse(body)
}

func (t *T0x0200) Encode() []byte {
	return t.T0x0200LocationItem.encode()
}

func (t *T0x0200) String() string {
	body := t.T0x0200LocationItem.encode()
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", consts.T0200LocationReport, body),
		t.T0x0200LocationItem.String(),
		"}",
	}, "\n")
}
