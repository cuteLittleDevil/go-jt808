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

type P0x9201 struct {
	BaseHandle
	// ServerIPLen 服务器IP地址长度
	ServerIPLen byte `json:"serverIPLen"`
	// ServerIPAddr 服务器IP地址
	ServerIPAddr string `json:"serverIPAddr"`
	// TcpPort 视频服务器TCP端口号，不使用TCP协议传输时保持默认值0即可（TCP和UDP二选一，当TCP和UDP均非默认值时一般以TCP为准）
	TcpPort uint16 `json:"tcpPort"`
	// UdpPort 视频服务器UDP端口号，不使用UDP协议传输时保持默认值0即可（TCP和UDP二选一，当TCP和UDP均非默认值时一般以TCP为准）
	UdpPort uint16 `json:"udpPort"`
	// ChannelNo 逻辑通道号
	ChannelNo byte `json:"channelNo"`
	// MediaType 音视频类型(媒体类型) 0-音频和视频 1-音频 2-视频 3-音频或视频
	MediaType byte `json:"mediaType"`
	// StreamType 码流类型 0-主或子码流 1-主码流 2-子码流
	StreamType byte `json:"streamType"`
	// MemoryType 存储器类型 0-主或灾备存储器 1-主存储器 2-灾备存储器
	MemoryType byte `json:"memoryType"`
	// PlaybackWay 回放方式 0-正常 1-快进 2-关键帧快退回放 3-关键帧播放 4-单帧上传
	PlaybackWay byte `json:"playbackWay"`
	// PlaySpeed 快进或快退倍数 为1和2时，此字段有效，否则置0 0-无效 1-1倍 2-2倍 3-4倍 4-8倍5-16倍
	PlaySpeed byte `json:"playSpeed"`
	// StartTime 开始时间 YY-MM-DD-HH-MM-SS，回放方式为4时 该字段表示单帧上传时间
	StartTime string `json:"startTime"`
	// EndTime 结束时间 YY-MM-DD-HH-MM-SS，回放方 该字段表示单帧上传时间
	EndTime string `json:"endTime"`
}

func (p *P0x9201) Protocol() consts.JT808CommandType {
	return consts.P9201SendVideoRecordRequest
}

func (p *P0x9201) ReplyProtocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (p *P0x9201) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 1 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.ServerIPLen = body[0]
	if len(body) != 1+int(p.ServerIPLen)+2+2+1+1+1+1+1+1+6+6 {
		return protocol.ErrBodyLengthInconsistency
	}
	n := int(p.ServerIPLen)
	p.ServerIPAddr = string(body[1 : n+1])
	p.TcpPort = binary.BigEndian.Uint16(body[n+1:])
	p.UdpPort = binary.BigEndian.Uint16(body[n+3:])
	p.ChannelNo = body[n+5]
	p.MediaType = body[n+6]
	p.StreamType = body[n+7]
	p.MemoryType = body[n+8]
	p.PlaybackWay = body[n+9]
	p.PlaySpeed = body[n+10]
	p.StartTime = utils.BCD2Time(body[n+11 : n+11+6])
	p.EndTime = utils.BCD2Time(body[n+11+6 : n+11+6+6])
	return nil
}

func (p *P0x9201) Encode() []byte {
	data := make([]byte, 0, 30)
	data = append(data, p.ServerIPLen)
	data = append(data, []byte(p.ServerIPAddr)...)
	data = binary.BigEndian.AppendUint16(data, p.TcpPort)
	data = binary.BigEndian.AppendUint16(data, p.UdpPort)
	data = append(data, p.ChannelNo)
	data = append(data, p.MediaType)
	data = append(data, p.StreamType)
	data = append(data, p.MemoryType)
	data = append(data, p.PlaybackWay)
	data = append(data, p.PlaySpeed)
	data = append(data, utils.Time2BCD(p.StartTime)...)
	data = append(data, utils.Time2BCD(p.EndTime)...)
	return data
}

func (p *P0x9201) HasReply() bool {
	return false
}

func (p *P0x9201) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%02x] IP长度:[%d]", p.ServerIPLen, p.ServerIPLen),
		fmt.Sprintf("\t[%x] IP地址:[%s]", p.ServerIPAddr, p.ServerIPAddr),
		fmt.Sprintf("\t[%04x] TCP 端口:[%d]", p.TcpPort, p.TcpPort),
		fmt.Sprintf("\t[%04x] UDP 端口:[%d]", p.UdpPort, p.UdpPort),
		fmt.Sprintf("\t[%02x] 通道号:[%d]", p.ChannelNo, p.ChannelNo),
		fmt.Sprintf("\t[%02x] 音视频类型(媒体类型):[%d] 0-音频和视频 1-音频 2-视频 3-音频或视频", p.MediaType, p.MediaType),
		fmt.Sprintf("\t[%02x] 码流类型:[%d] 0-主或子码流 1-主码流 2-子码流", p.StreamType, p.StreamType),
		fmt.Sprintf("\t[%02x] 存储器类型:[%d] 0-主或灾备存储器 1-主存储器 2-灾备存储器", p.MemoryType, p.MemoryType),
		fmt.Sprintf("\t[%02x] 回放方式:[%d] 0-正常 1-快进 2-关键帧快退回放 3-关键帧播放 4-单帧上传", p.PlaybackWay, p.PlaybackWay),
		fmt.Sprintf("\t[%02x] 快进或快退倍数:[%d] 为1和2时，此字段有效，否则置0 0-无效 1-1倍 2-2倍 3-4倍 4-8倍5-16倍", p.PlaySpeed, p.PlaySpeed),
		fmt.Sprintf("\t[%012x] 开始时间:[%s]", utils.Time2BCD(p.StartTime), p.StartTime),
		fmt.Sprintf("\t[%012x] 结束时间:[%s]", utils.Time2BCD(p.EndTime), p.EndTime),
		"}",
	}, "\n")
}
