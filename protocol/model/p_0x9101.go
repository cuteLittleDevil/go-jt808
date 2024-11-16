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

type P0x9101 struct {
	BaseHandle
	// ServerIPLen 视频服务器IP地址长度
	ServerIPLen byte `json:"serverIPLen"`
	// ServerIPAddr 视频服务器IP地址
	ServerIPAddr string `json:"serverIPAddr"`
	// TcpPort 视频服务器TCP端口号，不使用TCP协议传输时保持默认值0即可（TCP和UDP二选一，当TCP和UDP均非默认值时一般以TCP为准）
	TcpPort uint16 `json:"tcpPort"`
	// UdpPort 视频服务器UDP端口号，不使用UDP协议传输时保持默认值0即可（TCP和UDP二选一，当TCP和UDP均非默认值时一般以TCP为准）
	UdpPort uint16 `json:"udpPort"`
	// ChannelNo 逻辑通道号
	ChannelNo byte `json:"channelNo"`
	// DataType 数据类型 0-音视频 1-视频 2-双向对讲 3-监听 4-中心广播 5-透传
	DataType byte `json:"dataType"`
	// StreamType 码流类型 0-主码流 1-子码流
	StreamType byte `json:"streamType"`
}

func (p *P0x9101) Protocol() consts.JT808CommandType {
	return consts.P9101RealTimeAudioVideoRequest
}

func (p *P0x9101) ReplyProtocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (p *P0x9101) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 1 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.ServerIPLen = body[0]
	if len(body) != 1+int(p.ServerIPLen)+7 {
		return protocol.ErrBodyLengthInconsistency
	}
	n := int(p.ServerIPLen)
	p.ServerIPAddr = string(body[1 : n+1])
	p.TcpPort = binary.BigEndian.Uint16(body[n+1:])
	p.UdpPort = binary.BigEndian.Uint16(body[n+3:])
	p.ChannelNo = body[n+5]
	p.DataType = body[n+6]
	p.StreamType = body[n+7]
	return nil
}

func (p *P0x9101) Encode() []byte {
	data := make([]byte, 0, 25)
	data = append(data, p.ServerIPLen)
	data = append(data, utils.String2FillingBytes(p.ServerIPAddr, len(p.ServerIPAddr))...)
	data = binary.BigEndian.AppendUint16(data, p.TcpPort)
	data = binary.BigEndian.AppendUint16(data, p.UdpPort)
	data = append(data, p.ChannelNo)
	data = append(data, p.DataType)
	data = append(data, p.StreamType)
	return data
}

func (p *P0x9101) HasReply() bool {
	return false
}

func (p *P0x9101) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%02x] 服务器IP地址长度:[%d]", p.ServerIPLen, p.ServerIPLen),
		fmt.Sprintf("\t[%x] 服务器IP地址:[%s]", p.ServerIPAddr, p.ServerIPAddr),
		fmt.Sprintf("\t[%04x] 服务器视频通道监听端口号(TCP):[%d]", p.TcpPort, p.TcpPort),
		fmt.Sprintf("\t[%04x] 服务器视频通道监听端口号(UDP):[%d]", p.UdpPort, p.UdpPort),
		fmt.Sprintf("\t[%02x] 逻辑通道号:[%d]", p.ChannelNo, p.ChannelNo),
		fmt.Sprintf("\t[%02x] 数据类型:[%d]", p.DataType, p.DataType),
		fmt.Sprintf("\t[%02x] 码流类型:[%d]", p.StreamType, p.StreamType),
		"}",
	}, "\n")
}
