package jt808

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type (
	Header struct {
		// ID 消息ID
		ID uint16 `json:"ID,omitempty"`
		// Property 消息属性
		Property *BodyProperty `json:"property"`
		// ProtocolVersion 协议版本 1-2011 2-2013 3-2019
		ProtocolVersion consts.ProtocolVersionType `json:"protocolVersion,omitempty"`
		// TerminalPhoneNo 根据安装后终端自身的手机号转换。手机号不足 12 位，则在前补充数字
		// 大陆手机号补充数字 0，港澳台则根据其区号进行位数补充
		TerminalPhoneNo string `json:"terminalPhoneNo,omitempty"`
		// SerialNumber 占用两个字节，为发送信息的序列号，用于接收方检测是否有信息的丢失，
		// 上级平台和下级平台接自己发送数据包的个数计数，互不影响。
		// 程序开始运行时等于零，发送第一帧数据时开始计数，到最大数后自动归零
		SerialNumber uint16 `json:"serialNumber,omitempty"`
		// SubPackageSum 消息总包数 不分包的时候为0
		SubPackageSum uint16 `json:"subPackageSum,omitempty"`
		// SubPackageNo 消息包序号
		SubPackageNo uint16 `json:"subPackageNo,omitempty"`

		// PlatformSerialNumber 平台的流水号
		PlatformSerialNumber uint16 `json:"platformSerialNumber,omitempty"`
		// ReplyID 平台回复的消息ID
		ReplyID uint16 `json:"replyID,omitempty"`

		// headEnd 请求头结束位置
		headEnd int
		// bcdTerminalPhoneNo 设备上传的bcd编码的手机号
		bcdTerminalPhoneNo []byte
	}

	BodyProperty struct {
		// Version 协议版本 jt808协议的
		Version uint8 `json:"version,omitempty"`
		// PacketFragmented 分包标识，1：长消息，有分包；2：无分包
		PacketFragmented uint8 `json:"packetFragmented,omitempty"`
		// EncryptMethod 加密标识，0为不加密
		// 当此三位都为 0，表示消息体不加密；
		// 当第 10 位为 1，表示消息体经过 RSA 算法加密；
		EncryptMethod uint8 `json:"encryptMethod,omitempty"`
		// BodyDayaLen 消息体长度
		BodyDayaLen uint16 `json:"bodyDayaLen,omitempty"`

		// attribute 消息属性的原始数据
		attribute uint16
		// bit15 保留位
		bit15 byte
		// bit14 2013版本的保留 2019版本为1
		bit14 byte
		// 是否分包
		isSubPackage bool
	}

	JTMessage struct {
		// Header 请求头 固定格式的
		Header *Header `json:"header"`
		// VerifyCode 校验码
		VerifyCode byte `json:"-"`
		// Body 有效数据
		Body []byte `json:"body"`
	}
)

func NewJTMessage() *JTMessage {
	return &JTMessage{
		Header: &Header{
			Property: &BodyProperty{},
		},
		VerifyCode: 0,
		Body:       nil,
	}
}

func (j *JTMessage) Decode(data []byte) error {
	escapeData, err := unescape(data)
	if err != nil {
		return err
	}
	if code := utils.CreateVerifyCode(escapeData); code != 0 {
		return protocol.ErrCheckCode
	}
	if err := j.Header.decode(escapeData); err != nil {
		return err
	}
	start := j.Header.headEnd
	end := start + int(j.Header.Property.BodyDayaLen)
	if end+1 != len(escapeData) {
		return protocol.ErrBodyLengthInconsistency
	}
	j.Body = escapeData[start:end]
	j.VerifyCode = escapeData[end]
	return nil
}

func (h *Header) decode(data []byte) error {
	if len(data) < 4 {
		return protocol.ErrHeaderLength2Short
	}
	h.ID = binary.BigEndian.Uint16(data[0:2])
	h.Property.decode(data[2:4])
	var (
		start    = 4
		phoneLen = 6
		version  = consts.JT808Protocol2013 // 默认2013版本
	)
	if h.Property.Version == 1 {
		start = 5
		phoneLen = 10
		version = consts.JT808Protocol2019
	}
	if len(data) < start+phoneLen+2 {
		return protocol.ErrHeaderLength2Short
	}
	h.ProtocolVersion = version
	h.bcdTerminalPhoneNo = data[start : start+phoneLen]
	h.TerminalPhoneNo = utils.Bcd2Dec(h.bcdTerminalPhoneNo)
	h.SerialNumber = binary.BigEndian.Uint16(data[start+phoneLen : start+phoneLen+2])
	end := start + phoneLen + 2
	if h.Property.isSubPackage {
		if len(data) < start+phoneLen+6 {
			return protocol.ErrHeaderLength2Short
		}
		h.SubPackageSum = binary.BigEndian.Uint16(data[start+phoneLen+2 : start+phoneLen+4])
		h.SubPackageNo = binary.BigEndian.Uint16(data[start+phoneLen+4 : start+phoneLen+6])
		end += 4
	}
	h.headEnd = end
	return nil
}

func (h *Header) Encode(body []byte) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, h.ReplyID)           // 写ID 回复消息的ID
	h.Property.BodyDayaLen = uint16(len(body))                   // 消息的长度改为回复的body长度
	_ = binary.Write(buf, binary.BigEndian, h.Property.encode()) // 写消息属性
	if h.ProtocolVersion == consts.JT808Protocol2019 {
		// 2019版本的标识
		buf.WriteByte(0x01)
	}
	buf.Write(h.bcdTerminalPhoneNo)                                 // 写终端手机号
	_ = binary.Write(buf, binary.BigEndian, h.PlatformSerialNumber) // 写流水号 平台回复的流水号
	buf.Write(body)                                                 // 写消息体
	code := utils.CreateVerifyCode(buf.Bytes())                     // 校验码
	buf.WriteByte(code)
	return escape(buf.Bytes()) // 转义
}

func (p *BodyProperty) decode(data []byte) {
	attribute := binary.BigEndian.Uint16(data)
	p.attribute = attribute
	p.bit15 = byte(attribute & 0x8000)      // 第15位 保留
	p.bit14 = byte((attribute >> 14) & 0b1) // 第14位 协议版本 0-2013 1-2019
	p.Version = p.bit14
	p.PacketFragmented = byte((attribute >> 13) & 0b1) // 第13位 分包
	if p.PacketFragmented == 1 {
		p.isSubPackage = true
	}
	p.EncryptMethod = uint8((attribute & 0x400) >> 10) // 第10-12位 加密方式 0-不加密 1-RSA
	p.BodyDayaLen = attribute & 0x3FF                  // 最低10位 消息体长度 3=011 F=1111
}

func (p *BodyProperty) encode() uint16 {
	return (uint16(p.bit15) << 15) | // 第15位 保留
		(uint16(p.Version) << 14) | // 第14位 协议版本 0-2013 1-2019
		(uint16(p.PacketFragmented) << 13) | // 第13位 分包
		(uint16(p.EncryptMethod) << 10) | // 第10位 加密方式
		p.BodyDayaLen // 最低10位 消息体长度
}

func (h *Header) String() string {
	str := ""
	switch h.ProtocolVersion {
	case consts.JT808Protocol2019:
		str += "[01] 协议版本号(2019):[1]\n"
		str += fmt.Sprintf("[%20x] 终端手机号:[%s]", h.bcdTerminalPhoneNo, h.TerminalPhoneNo)
	default:
		str += fmt.Sprintf("[%12x] 终端手机号:[%s]", h.bcdTerminalPhoneNo, h.TerminalPhoneNo)
	}
	return strings.Join([]string{
		fmt.Sprintf("[%04x] 消息ID:[%d] [%s]", h.ID, h.ID, consts.TerminalRequestType(h.ID)),
		h.Property.String(),
		str,
		fmt.Sprintf("[%04x] 消息流水号:[%d]", h.SerialNumber, h.SerialNumber),
	}, "\n")
}

func (p *BodyProperty) String() string {
	version := consts.JT808Protocol2013
	if p.Version == 1 {
		version = consts.JT808Protocol2019
	}
	return strings.Join([]string{
		"消息体属性对象: {",
		fmt.Sprintf("\t[%016b] 消息体属性对象:[%d]", p.attribute, p.attribute),
		fmt.Sprintf("\t版本号:[%s]", version.String()),
		fmt.Sprintf("\t[bit15] [%d]", p.bit15),
		fmt.Sprintf("\t[bit14] 协议版本标识:[%d]", p.bit14),
		fmt.Sprintf("\t[bit13] 是否分包:[%t]", p.isSubPackage),
		fmt.Sprintf("\t[bit10-12] 加密标识:[%d] 0-不加密 1-RSA", p.EncryptMethod),
		fmt.Sprintf("\t[bit0-bit9] 消息体长度:[%d]", p.BodyDayaLen),
		"}",
	}, "\n")
}
