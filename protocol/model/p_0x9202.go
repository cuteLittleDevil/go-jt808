package model

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x9202 struct {
	BaseHandle
	// ChannelNo 音视频通道号
	ChannelNo byte `json:"channelNo"`
	// PlayControl 回放控制 0-开始 1-暂停 2-结束 3-快进 4-关键帧快退播放 5-拖动(到指定位置) 6-关键帧播放
	PlayControl byte `json:"playControl"`
	// PlaySpeed 快进或快退倍数 PlayControl=3或4时生效 0-无效 1-1倍 2-2倍 3-4倍 4-8倍 5-16倍
	PlaySpeed byte `json:"playSpeed"`
	// DateTime 拖动回放位置 PlayControl=5时生效 YY-MM-DD-HH-MM-SS
	DateTime string `json:"dateTime"`
}

func (p *P0x9202) Protocol() consts.JT808CommandType {
	return consts.P9202SendVideoRecordControl
}

func (p *P0x9202) ReplyProtocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (p *P0x9202) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) != 9 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.ChannelNo = body[0]
	p.PlayControl = body[1]
	p.PlaySpeed = body[2]
	p.DateTime = utils.BCD2Time(body[3:9])
	return nil
}

func (p *P0x9202) Encode() []byte {
	data := make([]byte, 9)
	data[0] = p.ChannelNo
	data[1] = p.PlayControl
	data[2] = p.PlaySpeed
	copy(data[3:9], utils.Time2BCD(p.DateTime))
	return data
}

func (p *P0x9202) HasReply() bool {
	return false
}

func (p *P0x9202) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t [%02x]音视频通道号:[%d]", p.ChannelNo, p.ChannelNo),
		fmt.Sprintf("\t [%02x]回放控制:[%d] 0-开始 1-暂停 2-结束 3-快进 4-关键帧快退播放 5-拖动(到指定位置) 6-关键帧播放", p.PlayControl, p.PlayControl),
		fmt.Sprintf("\t [%02x]快进或快退倍数:[%d] PlayControl=3或4时生效 0-无效 1-1倍 2-2倍 3-4倍 4-8倍 5-16倍", p.PlaySpeed, p.PlaySpeed),
		fmt.Sprintf("\t [%02x]拖动回放位置:[%s] PlayControl=5时生效 ", utils.Time2BCD(p.DateTime), p.DateTime),
		"}",
	}, "\n")
}
