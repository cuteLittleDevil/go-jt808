package model

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x9102 struct {
	BaseHandle
	// ChannelNo 逻辑通道号
	ChannelNo byte `json:"channelNo"`
	// ControlCmd 控制指令
	// 平台可以通过该指令对设备的实时音视频进行控制：
	// 0:关闭音视频传输指令
	// 1:切换码流（增加暂停和继续）
	// 2:暂停该通道所有流的发送
	// 3:恢复暂停前流的发送,与暂停前的流类型一致
	// 4:关闭双向对讲
	ControlCmd byte `json:"controlCmd"`
	// CloseAudioVideoData 关闭音视频类型
	// 0:关闭该通道有关的音视频数据
	// 1:只关闭该通道有关的音频，保留该通道有关的视频
	// 2:只关闭该通道有关的视频，保留该通道有关的音频
	CloseAudioVideoData byte `json:"closeAudioVideoData"`
	// StreamType 切换码流类型
	// 将之前申请的码流切换为新申请的码流，音频与切换前保持一致。
	// 新申请的码流为：
	// 0:主码流
	// 1:子码流
	StreamType byte `json:"streamType"`
}

func (p *P0x9102) Protocol() consts.JT808CommandType {
	return consts.P9102AudioVideoControl
}

func (p *P0x9102) ReplyProtocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (p *P0x9102) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) != 4 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.ChannelNo = body[0]
	p.ControlCmd = body[1]
	p.CloseAudioVideoData = body[2]
	p.StreamType = body[3]
	return nil
}

func (p *P0x9102) Encode() []byte {
	data := make([]byte, 4)
	data[0] = p.ChannelNo
	data[1] = p.ControlCmd
	data[2] = p.CloseAudioVideoData
	data[3] = p.StreamType
	return data
}

func (p *P0x9102) HasReply() bool {
	return false
}

func (p *P0x9102) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%02x]逻辑通道号:[%d]", p.ChannelNo, p.ChannelNo),
		fmt.Sprintf("\t[%02x]控制指令:[%d]", p.ControlCmd, p.ControlCmd),
		fmt.Sprintf("\t[%02x]关闭音视频类型:[%d] 0-关闭音视频 1-关闭音频 2-关闭视频", p.CloseAudioVideoData, p.CloseAudioVideoData),
		fmt.Sprintf("\t[%02x]切换码流类型:[%d] 0-主码流 1-子码流", p.StreamType, p.StreamType),
		"}",
	}, "\n")
}
