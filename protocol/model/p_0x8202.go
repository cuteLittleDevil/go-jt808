package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x8202 struct {
	BaseHandle
	// 时间间隔 单位为秒(s) 0则停止跟踪 停止跟踪后无需带后继字段
	TimeInterval uint16 `json:"timeInterval"`
	// 位置跟踪有效期 单位为秒(s) 字段在接收到位置跟踪控制消息后
	// 在有效期截止时间之前 依据消息中的时间间隔发送位置汇报
	TrackValidity uint32 `json:"trackValidity"`
}

func (p *P0x8202) Protocol() consts.JT808CommandType {
	return consts.P8202TmpLocationTrack
}

func (p *P0x8202) ReplyProtocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (p *P0x8202) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 6 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.TimeInterval = binary.BigEndian.Uint16(body[:2])
	p.TrackValidity = binary.BigEndian.Uint32(body[2:])
	return nil
}

func (p *P0x8202) Encode() []byte {
	data := make([]byte, 0, 6)
	data = binary.BigEndian.AppendUint16(data, p.TimeInterval)
	if p.TimeInterval != 0 {
		data = binary.BigEndian.AppendUint32(data, p.TrackValidity)
	}
	return data
}

func (p *P0x8202) HasReply() bool {
	return false
}

func (p *P0x8202) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%04x] 时间间隔:[%d] 单位为秒(s) 0则停止跟踪 停止跟踪后无需带后继字段", p.TimeInterval, p.TimeInterval),
		fmt.Sprintf("\t[%08x] 位置跟踪有效期:[%d]", p.TrackValidity, p.TrackValidity),
		"}",
	}, "\n")
}
