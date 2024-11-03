package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
)

var ErrLen2Short = errors.New("data len is too short")

const (
	PTG711A = uint8(6)
	PTG711U = uint8(7)
	PTH264  = uint8(98)
	PTH265  = uint8(99)
)

type DataType uint8

const (
	DataTypeI DataType = iota
	DataTypeP
	DataTypeB
	DataTypeA
	DataTypePenetrate
)

func (d DataType) String() string {
	switch d {
	case DataTypeI:
		return "视频I祯"
	case DataTypeP:
		return "视频P帧"
	case DataTypeB:
		return "视频B帧"
	case DataTypeA:
		return "音频帧"
	case DataTypePenetrate:
		return "透传数据"
	}
	return fmt.Sprintf("未知类型:%d", uint8(d))
}

type SubcontractType uint8

const (
	SubcontractTypeAtomic SubcontractType = iota
	SubcontractTypeFirst
	SubcontractTypeLast
	SubcontractTypeMiddle
)

func (s SubcontractType) String() string {
	switch s {
	case SubcontractTypeAtomic:
		return "原子祯"
	case SubcontractTypeFirst:
		return "第一祯"
	case SubcontractTypeLast:
		return "最后祯"
	case SubcontractTypeMiddle:
		return "中间祯"
	}
	return fmt.Sprintf("未知类型:%d", uint8(s))
}

type (
	Packet struct {
		ID string // 帧头标识
		Flag
		Seq          uint16 // 包序号 初始为0，每发送一个RTP数据包，序列号加1
		Sim          string // 终端设备SIM卡号 bcd[6]
		LogicChannel uint8  // 逻辑通道号
		DataType            // 0000：数据I祯，0001：视频P帧，0010：视频B帧，0011：音频帧，0100：透传数据
		SubcontractType
		Timestamp int64 // 标识此RTP数据包当前祯的相对时间，单位毫秒（ms）。当数据类型为0100时，则没有该字段，
		// RTP协议里规定这个数字是以任意值开始，然后按毫秒的时间间隔递增即可，
		//千万不要认为它是常规的时间戳的定义，目前有碰到个别厂家提供的时间戳有问题，一直不变。
		LastIFrameInterval uint16 // 该祯与上一个关键祯之间的时间间隔，单位毫秒（ms），当数据类型为非视频祯时，则没有该字段
		LastFrameInterval  uint16 // 该祯与上一个祯之间的时间间隔，单位毫秒（ms），当数据类型为非视频祯时，则没有该字段
		DataLen            uint16 // 后续数据体长度，不含此字段
		Data               []byte
		customAttributes   // 非包内容 用于方便计算
	}

	Flag struct {
		V  uint8 // 2 BITS 固定为2
		P  uint8 // 1 BIT 固定为0
		X  uint8 // 1 BIT RTP头是否需要扩展位，固定为0
		CC uint8 // 4 BITS 固定为1
		M  uint8 // 1 BITS 标志位，确定是否完整数据帧的边界，因为数据体的最大长度是950字节，而一个视频I帧通道要远远超过950字节，所以视频的一个帧通常会分包
		PT uint8 // 7bits 负载类型，原文档里的这里的参考表是错误的，实际上是参考文档的表12，此文章的附录里也有
	}

	customAttributes struct {
		hasParseTimestamp    bool
		hasParseLastInterval bool
		headData             []byte
	}
)

func NewPacket() *Packet {
	return &Packet{}
}

// Parse 解析数据
// 1 解析正常 返回多余的数据包
// 2 数据不够解析失败 返回错误
func (p *Packet) Parse(data []byte) (remainingData []byte, err error) {
	const minDataLen = 16 //
	if len(data) < minDataLen {
		return nil, ErrLen2Short
	}
	// 解析头部 获取剩余数据长度
	buf := bytes.NewBuffer(data[:minDataLen])
	fs := []func(buf *bytes.Buffer) error{
		p.parseID,
		p.parseFlag,
		p.parseSeq,
		p.parseSim,
		p.parseLogicChannel,
		p.parseType,
	}
	for _, f := range fs {
		if err := f(buf); err != nil {
			return nil, err
		}
	}
	dataHeadLen := p.calculateDataHeadLen()
	if minDataLen+dataHeadLen > len(data) {
		return nil, ErrLen2Short
	}
	buf = bytes.NewBuffer(data[minDataLen : minDataLen+dataHeadLen])
	p.customAttributes.headData = data[:minDataLen+dataHeadLen]
	if err := p.parseTimeAndDataLen(buf); err != nil {
		return nil, err
	}

	intactLen := minDataLen + dataHeadLen + int(p.DataLen)
	if len(data) < intactLen {
		return nil, ErrLen2Short
	}
	p.Data = data[minDataLen+dataHeadLen : intactLen]
	return data[intactLen:], nil
}

func (p *Packet) parseID(buf *bytes.Buffer) error {
	const defaultID = "01cd" // 1078协议固定
	var identification [4]byte
	if err := binary.Read(buf, binary.BigEndian, &identification); err != nil {
		return err
	}
	id := string(identification[:])
	if id != defaultID {
		panic("invalid identification")
	}
	p.ID = id
	return nil
}

func (p *Packet) parseFlag(buf *bytes.Buffer) error {
	var attribute uint16
	if err := binary.Read(buf, binary.BigEndian, &attribute); err != nil {
		return err
	}
	p.Flag = Flag{
		V:  uint8(attribute >> 14),
		P:  uint8(attribute << 2 >> 15),
		X:  uint8(attribute << 3 >> 15),
		CC: uint8(attribute << 4 >> 12),
		M:  uint8(attribute << 8 >> 15),
		PT: uint8(attribute & 0b1111111),
	}
	return nil
}

func (p *Packet) parseSeq(buf *bytes.Buffer) error {
	return binary.Read(buf, binary.BigEndian, &p.Seq)
}

func (p *Packet) parseSim(buf *bytes.Buffer) error {
	var sims [6]byte
	if err := binary.Read(buf, binary.BigEndian, &sims); err != nil {
		return err
	}
	p.Sim = utils.Bcd2Dec(sims[:])
	return nil
}

func (p *Packet) parseLogicChannel(buf *bytes.Buffer) error {
	return binary.Read(buf, binary.BigEndian, &p.LogicChannel)
}

func (p *Packet) parseType(buf *bytes.Buffer) error {
	var flag uint8
	if err := binary.Read(buf, binary.BigEndian, &flag); err != nil {
		return err
	}
	dataType := DataTypeI
	switch (flag >> 4) & 0x0f {
	case 0x00:
		dataType = DataTypeI
	case 0x01:
		dataType = DataTypeP
	case 0x02:
		dataType = DataTypeB
	case 0x03:
		dataType = DataTypeA
	case 0x04:
		dataType = DataTypePenetrate
	default:
	}

	subPackType := SubcontractTypeAtomic
	switch flag & 0x0f {
	case 0x00:
		subPackType = SubcontractTypeAtomic
	case 0x01:
		subPackType = SubcontractTypeFirst
	case 0x02:
		subPackType = SubcontractTypeLast
	case 0x03:
		subPackType = SubcontractTypeMiddle
	default:
	}
	p.DataType = dataType
	p.SubcontractType = subPackType
	p.hasParseTimestamp = true
	if p.DataType == DataTypePenetrate {
		// 当数据类型为0100(透传)时，则没有时间戳字段
		p.hasParseTimestamp = false
	}
	p.hasParseLastInterval = false
	if p.DataType == DataTypeI || p.DataType == DataTypeP || p.DataType == DataTypeB {
		// 当数据类型为视频帧时则有时间戳
		p.hasParseLastInterval = true
	}
	return nil
}

func (p *Packet) calculateDataHeadLen() int {
	tmp := 2
	if p.hasParseTimestamp {
		tmp += 8
	}
	if p.hasParseLastInterval {
		tmp += 4
	}
	return tmp
}

func (p *Packet) parseTimeAndDataLen(buf *bytes.Buffer) error {
	if p.hasParseTimestamp {
		if err := binary.Read(buf, binary.BigEndian, &p.Timestamp); err != nil {
			return err
		}
	}
	if p.hasParseLastInterval {
		if err := binary.Read(buf, binary.BigEndian, &p.LastIFrameInterval); err != nil {
			return err
		}
		if err := binary.Read(buf, binary.BigEndian, &p.LastFrameInterval); err != nil {
			return err
		}
	}
	return binary.Read(buf, binary.BigEndian, &p.DataLen)
}
