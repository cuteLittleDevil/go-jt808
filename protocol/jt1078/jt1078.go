package jt1078

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"strings"
)

type (
	Packet struct {
		// ID  帧头标识 固定为0x30 0x31 0x63 0x64 既01cd
		ID string
		// Flag 标志
		Flag
		// Seq 包序号 初始为0，每发送一个RTP数据包，序列号加1
		Seq uint16
		// Sim 终端设备SIM卡号 bcd[6]
		Sim string
		// LogicChannel 逻辑通道号
		LogicChannel uint8
		// DataType 数据帧类型 0000-数据I祯 0001-视频P帧 0010-视频B帧 0011-音频帧 0100-透传数据
		DataType
		// SubcontractType 分包处理标记 0000-原子包 0001-分包处理时的第一个包 0010-分包处理的最后一个包 0011-分包处理时的中间包
		SubcontractType
		// Timestamp 标识此RTP数据包当前祯的相对时间 单位毫秒（ms）当数据类型为0100时，则没有该字段
		// RTP协议里规定这个数字是以任意值开始 然后按毫秒的时间间隔递增即可
		// 千万不要认为它是常规的时间戳的定义 目前有碰到个别厂家提供的时间戳有问题 一直不变
		Timestamp uint64
		// LastIFrameInterval 该祯与上一个关键祯之间的时间间隔，单位毫秒（ms）当数据类型为非视频祯时 则没有该字段
		LastIFrameInterval uint16
		// LastFrameInterval 该祯与上一个祯之间的时间间隔，单位毫秒（ms）当数据类型为非视频祯时 则没有该字段
		LastFrameInterval uint16
		// DataBodyLen 后续数据体长度 不含此字段
		DataBodyLen uint16
		// Body 有效数据 音频数据或透传数据 长度不超过950 byte
		Body []byte
		// customAttributes 自定义信息 非包内容 用于方便计算
		customAttributes
	}

	Flag struct {
		// V 2 BITS 固定为2
		V uint8
		// P 1 BIT 固定为0
		P uint8
		// X 1 BIT RTP头是否需要扩展位，固定为0
		X uint8
		// CC 4 BITS 固定为1
		CC uint8
		// M 1 BITS 标志位，确定是否完整数据帧的边界，因为数据体的最大长度是950字节，而一个视频I帧通道要远远超过950字节，所以视频的一个帧通常会分包
		M uint8
		// PT 7bits 负载类型，原文档里的这里的参考表是错误的，实际上是参考文档的表12，此文章的附录里也有
		PT PTType
	}

	customAttributes struct {
		// videoFrame 是否为视频帧
		videoFrame bool
		// headEnd 头部是到哪里结束的
		headEnd int
	}
)

func NewPacket() *Packet {
	return &Packet{}
}

func (p *Packet) Decode(data []byte) (remainData []byte, err error) {
	if err := p.decodeHead(data); err != nil {
		return data, err
	}
	body := data[p.headEnd:]
	if len(body) < int(p.DataBodyLen) {
		return body, errors.Join(fmt.Errorf("cur body len is [%d]", len(body)), ErrBodyLength2Short)
	}
	p.Body = body[:p.DataBodyLen]
	if len(body) == int(p.DataBodyLen) {
		return nil, nil
	}
	return body[p.DataBodyLen:], nil
}

func (p *Packet) decodeHead(data []byte) error {
	if len(data) < 16 {
		return ErrHeaderLength2Short
	}
	p.ID = string(data[:4])
	if p.ID != "01cd" { // 1078协议固定
		fmt.Println(fmt.Sprintf("%x", data[:16]))
		return errors.Join(fmt.Errorf("id is [%s]", p.ID), ErrUnqualifiedData)
	}

	attr := data[4]
	sign := data[5]
	p.Flag = Flag{
		V:  (attr >> 6) & 0b11,
		P:  (attr >> 5) & 0b1,
		X:  (attr >> 4) & 0b1,
		CC: attr & 0b1111,
		M:  (sign >> 7) & 0b1,
		PT: PTType(sign & 0b1111_111),
	}

	p.Seq = binary.BigEndian.Uint16(data[6:8])
	p.Sim = utils.Bcd2Dec(data[8:14])
	p.LogicChannel = data[14]
	p.DataType = DataType((data[15] >> 4) & 0x0F)
	p.SubcontractType = SubcontractType(data[15] & 0x0F)

	end := 18
	if p.DataType != DataTypePenetrate {
		end += 8
	}
	if p.DataType == DataTypeI || p.DataType == DataTypeP || p.DataType == DataTypeB {
		p.customAttributes.videoFrame = true
		end += 4
	}

	if len(data) < end {
		return ErrHeaderLength2Short
	}
	start := 16
	if p.DataType != DataTypePenetrate {
		p.Timestamp = binary.BigEndian.Uint64(data[16:24])
		start = 24
	}
	if p.customAttributes.videoFrame {
		p.LastIFrameInterval = binary.BigEndian.Uint16(data[start : start+2])
		p.LastFrameInterval = binary.BigEndian.Uint16(data[start+2 : start+4])
		start += 4
	}
	p.DataBodyLen = binary.BigEndian.Uint16(data[start : start+2])
	p.headEnd = start + 2
	return nil
}

func (p *Packet) String() string {
	str := ""
	if p.DataType != DataTypePenetrate {
		str += fmt.Sprintf("\t[%016x] 标识此RTP数据包当前祯的相对时间:[%d]\n", p.Timestamp, p.Timestamp)
	}
	if p.customAttributes.videoFrame {
		str += fmt.Sprintf("\t[%04x] 该帧与上一个关键帧之间的时间间隔:[%d]单位毫秒(ms)\n",
			p.LastIFrameInterval, p.LastIFrameInterval)
		str += fmt.Sprintf("\t[%04x] 该祯与上一个祯之间的时间间隔:[%d]单位毫秒(ms)\n",
			p.LastFrameInterval, p.LastFrameInterval)
	}
	str += fmt.Sprintf("\t[%04x] 后续数据体长度:[%d]不含此字段", p.DataBodyLen, p.DataBodyLen)
	return strings.Join([]string{
		"{",
		fmt.Sprintf("\t[%04x] 固定标志头:[%s]", p.ID, p.ID),
		fmt.Sprintf("\t\tV:[%d] 固定为2", p.Flag.V),
		fmt.Sprintf("\t\tP:[%d] 固定为0", p.Flag.P),
		fmt.Sprintf("\t\tX:[%d] RTP头是否需要扩展位固定为0", p.Flag.X),
		fmt.Sprintf("\t\tCC:[%d] 固定为1", p.Flag.CC),
		fmt.Sprintf("\t\tM:[%d] 确定是否是完整数据帧的边界", p.Flag.M),
		fmt.Sprintf("\t\tPT:[%d] 负载类型[%s]", p.Flag.PT, p.Flag.PT.String()),
		fmt.Sprintf("\t[%04x] 序列号:[%d]", p.Seq, p.Seq),
		fmt.Sprintf("\tSIM卡号:[%s]", p.Sim),
		fmt.Sprintf("\t[%02x] 逻辑通道号:[%d]", p.LogicChannel, p.LogicChannel),
		fmt.Sprintf("\t[%02x] 数据帧类型:[%d] [%s]", uint8(p.DataType), p.DataType, p.DataType.String()),
		fmt.Sprintf("\t[%02x] 分包处理标记:[%d] [%s]", uint8(p.SubcontractType), p.SubcontractType, p.SubcontractType.String()),
		str,
		fmt.Sprintf("\t数据体长度:[%d]", len(p.Body)),
		"}",
	}, "\n")
}
