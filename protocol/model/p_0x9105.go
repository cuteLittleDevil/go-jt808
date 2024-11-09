package model

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x9105 struct {
	BaseHandle
	// ChannelNo 逻辑通道号
	ChannelNo byte `json:"channelNo"`
	// PackageLossRate 丢包率 当前传输通道的丢包率，数值乘以100之后取整部分
	PackageLossRate byte `json:"packageLossRate"`
}

func (p *P0x9105) Protocol() consts.JT808CommandType {
	return consts.P9105AudioVideoControlStatusNotice
}

func (p *P0x9105) ReplyProtocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (p *P0x9105) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) != 2 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.ChannelNo = body[0]
	p.PackageLossRate = body[1]
	return nil
}

func (p *P0x9105) Encode() []byte {
	data := make([]byte, 2)
	data[0] = p.ChannelNo
	data[1] = p.PackageLossRate
	return data
}

func (p *P0x9105) HasReply() bool {
	return false
}

func (p *P0x9105) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%02x]逻辑通道号:[%d]", p.ChannelNo, p.ChannelNo),
		fmt.Sprintf("\t[%02x]丢包率:[%d]", p.PackageLossRate, p.PackageLossRate),
		"}",
	}, "\n")
}
