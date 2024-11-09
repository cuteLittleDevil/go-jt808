package model

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x8103 struct {
	BaseHandle
	// ParamTotal 参数总数
	ParamTotal uint8 `json:"paramTotal"`
	// TerminalParamDetails 参数项列表 设置终端参数消息体数据格式见表10
	TerminalParamDetails TerminalParamDetails `json:"terminalParamDetails"`
}

func (p *P0x8103) Protocol() consts.JT808CommandType {
	return consts.P8103SetTerminalParams
}

func (p *P0x8103) ReplyProtocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (p *P0x8103) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 1 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.ParamTotal = body[0]
	return p.TerminalParamDetails.parse(p.ParamTotal, body[1:])
}

func (p *P0x8103) Encode() []byte {
	data := make([]byte, 1, 100)
	data[0] = p.ParamTotal
	data = append(data, p.TerminalParamDetails.encode()...)
	return data
}

func (p *P0x8103) HasReply() bool {
	return false
}

func (p *P0x8103) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t [%02x]参数总数:[%d]", p.ParamTotal, p.ParamTotal),
		p.TerminalParamDetails.String(),
		"}",
	}, "\n")
}
