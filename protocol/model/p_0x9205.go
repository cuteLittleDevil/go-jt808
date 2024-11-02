package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x9205 struct {
	BaseHandle
	// ChannelNo 逻辑通道号 按照JT/T 1076—2016中的表2,0表示所有通道
	ChannelNo byte `json:"channelNo"`
	// StartTime YY-MM-DD-HH-MM-SS，全0表示无起始时间条件
	StartTime string `json:"startTime"`
	// EndTime YY-MM-DD-HH-MM-SS，全0表示无终止时间条件
	EndTime string `json:"endTime"`
	// AlarmFlag 告警标志 bit0 ~ bit31 见 JT/T 808—2011 表18 报警
	// bit32~ bi63 见表13；全0表示无报警类型条件
	AlarmFlag uint64 `json:"alarmFlag"`
	// MediaType 媒体类型 0-音视频，1-音频，2-视频，3-视频或音视频
	MediaType byte `json:"mediaType"`
	// StreamType 码流类型 0-所有码流 1-主码流 2-子码流
	StreamType byte `json:"streamType"`
	// StorageType 存储器类型 0-所有存储器 1-主存储器 2-灾备存储器
	StorageType byte `json:"storageType"`
}

func (p *P0x9205) Protocol() consts.JT808CommandType {
	return consts.P9205QueryResourceList
}

func (p *P0x9205) ReplyProtocol() consts.JT808CommandType {
	return consts.T1205UploadAudioVideoResourceList
}

func (p *P0x9205) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) != 24 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.ChannelNo = body[0]
	p.StartTime = utils.BCD2Time(body[1 : 1+6])
	p.EndTime = utils.BCD2Time(body[1+6 : 1+12])
	p.AlarmFlag = binary.BigEndian.Uint64(body[13 : 13+8])
	p.MediaType = body[21]
	p.StreamType = body[22]
	p.StorageType = body[23]
	return nil
}

func (p *P0x9205) Encode() []byte {
	data := make([]byte, 24)
	data[0] = p.ChannelNo
	copy(data[1:7], utils.Time2BCD(p.StartTime))
	copy(data[7:13], utils.Time2BCD(p.EndTime))
	binary.BigEndian.PutUint16(data[13:21], uint16(p.AlarmFlag))
	data[21] = p.MediaType
	data[22] = p.StreamType
	data[23] = p.StorageType
	return data
}

func (p *P0x9205) HasReply() bool {
	return false
}

func (p *P0x9205) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%02x] 逻辑通道号:[%d]", p.ChannelNo, p.ChannelNo),
		fmt.Sprintf("\t[%012x] 开始时间:[%s]", utils.Time2BCD(p.StartTime), p.StartTime),
		fmt.Sprintf("\t[%012x] 结束时间:[%s]", utils.Time2BCD(p.EndTime), p.EndTime),
		fmt.Sprintf("\t[%016x] 告警标志:[%d]", p.AlarmFlag, p.AlarmFlag),
		fmt.Sprintf("\t[%02x] 媒体类型:[%d] 0-音视频，1-音频，2-视频，3-视频或音视频", p.MediaType, p.MediaType),
		fmt.Sprintf("\t[%02x] 码流类型:[%d] 0-所有码流 1-主码流 2-子码流", p.StreamType, p.StreamType),
		fmt.Sprintf("\t[%02x] 存储器类型:[%d] 0-所有存储器 1-主存储器 2-灾备存储器", p.StorageType, p.StorageType),
		"}",
	}, "\n")
}
