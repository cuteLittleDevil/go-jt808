package model

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x8104 struct {
	BaseHandle
}

func (p *P0x8104) Protocol() consts.JT808CommandType {
	return consts.P8104QueryTerminalParams
}

func (p *P0x8104) ReplyProtocol() consts.JT808CommandType {
	return consts.T0104QueryParameter
}

func (p *P0x8104) Parse(_ *jt808.JTMessage) error {
	return nil
}

func (p *P0x8104) Encode() []byte {
	return nil
}

func (p *P0x8104) ReplyBody(_ *jt808.JTMessage) ([]byte, error) {
	return nil, nil
}

func (p *P0x8104) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		"\t查询终端参数:null",
		"}",
	}, "\n")
}
