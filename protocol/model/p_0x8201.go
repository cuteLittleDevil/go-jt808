package model

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x8201 struct {
	BaseHandle
}

func (p *P0x8201) Protocol() consts.JT808CommandType {
	return consts.P8201QueryLocation
}

func (p *P0x8201) ReplyProtocol() consts.JT808CommandType {
	return consts.T0201QueryLocation
}

func (p *P0x8201) Parse(_ *jt808.JTMessage) error {
	return nil
}

func (p *P0x8201) Encode() []byte {
	return nil
}

func (p *P0x8201) HasReply() bool {
	return false
}

func (p *P0x8201) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		"}",
	}, "\n")
}
