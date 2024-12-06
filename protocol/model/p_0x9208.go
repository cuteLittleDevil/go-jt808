package model

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type (
	P0x9208 struct {
		BaseHandle
		// ServerIPLen 服务器IP地址长度
		ServerIPLen byte `json:"serverIPLen"`
		// ServerAddr 服务器IP地址
		ServerAddr string `json:"serverAddr"`
		// TcpPort tcp端口
		TcpPort uint16 `json:"tcpPort"`
		// UdpPort udp端口
		UdpPort        uint16 `json:"udpPort"`
		P9208AlarmSign `json:"p9208AlarmSign"`
		// AlarmID 报警编号 byte[32] 平台给报警分配的唯一编号
		AlarmID string `json:"alarmID"`
		// Reserve 预留 16位
		Reserve []byte `json:"reserve"`
	}

	P9208AlarmSign struct {
		// TerminalID 终端ID 苏标7 黑标30 广东标30 湖南标7 四川标30
		TerminalID string `json:"terminalID"`
		// Time 时间 bcd[6]
		Time string `json:"time"`
		// SerialNumber 序号
		SerialNumber byte `json:"serialNumber"`
		// AttachNumber 附件数量
		AttachNumber byte `json:"attachNumber"`
		// AlarmReserve 预留 苏标1 黑标0 广东标2 湖南1 四川标1
		AlarmReserve []byte `json:"alarmReserve"`
		// ActiveSafetyType 主动安全告警类型
		consts.ActiveSafetyType `json:"activeSafetyType"`
	}
)

func (p *P0x9208) Protocol() consts.JT808CommandType {
	return consts.P9208AlarmAttachUpload
}

func (p *P0x9208) ReplyProtocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (p *P0x9208) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	sign := 1 + 2 + 2 + p.P9208AlarmSign.getAlarmSignLen() + 32
	if len(body) < sign {
		return protocol.ErrBodyLengthInconsistency
	}
	p.ServerIPLen = body[0]
	k := int(p.ServerIPLen)
	if sign+k > len(body) {
		return protocol.ErrBodyLengthInconsistency
	}
	p.ServerAddr = string(body[1 : 1+k])
	p.TcpPort = binary.BigEndian.Uint16(body[1+k : 1+k+2])
	p.UdpPort = binary.BigEndian.Uint16(body[3+k : 3+k+2])
	p.P9208AlarmSign.parse(body[5+k : 5+k+16])
	p.AlarmID = string(bytes.Trim(body[sign+k-32:sign+k], "\x00"))
	p.Reserve = body[sign+k:]
	return nil
}

func (p *P0x9208) Encode() []byte {
	data := make([]byte, 0, 53+int(p.ServerIPLen))
	data = append(data, p.ServerIPLen)
	data = append(data, []byte(p.ServerAddr)...)
	data = binary.BigEndian.AppendUint16(data, p.TcpPort)
	data = binary.BigEndian.AppendUint16(data, p.UdpPort)
	data = append(data, p.P9208AlarmSign.encode()...)
	data = append(data, utils.String2FillingBytes(p.AlarmID, 32)...)
	data = append(data, p.Reserve...)
	return data
}

func (p *P0x9208) HasReply() bool {
	return false
}

func (p *P0x9208) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t [%02x]服务器IP地址长度:[%d]", p.ServerIPLen, p.ServerIPLen),
		fmt.Sprintf("\t [%x]服务器IP地址:[%s]", p.ServerAddr, p.ServerAddr),
		fmt.Sprintf("\t [%04x]TCP端口:[%d]", p.TcpPort, p.TcpPort),
		fmt.Sprintf("\t [%04x]UDP端口:[%d]", p.UdpPort, p.UdpPort),
		p.P9208AlarmSign.String(),
		fmt.Sprintf("\t [%064x]告警ID:[%s]", p.AlarmID, p.AlarmID),
		fmt.Sprintf("\t [%032x]预留:[%x]", p.Reserve, p.Reserve),
		"}",
	}, "\n")
}

func (p *P9208AlarmSign) parse(data []byte) {
	idLen := p.getTerminalIDLen()
	p.TerminalID = string(bytes.Trim(data[:idLen], "\x00"))
	p.Time = utils.BCD2Time(data[idLen : idLen+6])
	p.SerialNumber = data[idLen+6]
	p.AttachNumber = data[idLen+7]
	if len(data) >= idLen+8 {
		p.AlarmReserve = data[idLen+8:]
	}
}

func (p *P9208AlarmSign) encode() []byte {
	data := make([]byte, 0, 16)
	data = append(data, utils.String2FillingBytes(p.TerminalID, p.getTerminalIDLen())...)
	data = append(data, utils.Time2BCD(p.Time)...)
	data = append(data, p.SerialNumber)
	data = append(data, p.AttachNumber)
	if len(p.AlarmReserve) > 0 {
		data = append(data, p.AlarmReserve...)
	}
	if num := p.getAlarmSignLen() - len(data); num > 0 {
		p.AlarmReserve = make([]byte, num)
		data = append(data, p.AlarmReserve...) // 忘记写预留的情况 帮忙补0
	}
	return data
}

func (p *P9208AlarmSign) getTerminalIDLen() int {
	switch p.ActiveSafetyType {
	case consts.ActiveSafetyJS:
		return 7
	case consts.ActiveSafetyHLJ:
		return 30
	case consts.ActiveSafetyGD:
		return 30
	case consts.ActiveSafetyHN:
		return 7
	case consts.ActiveSafetySC:
		return 30
	default:
	}
	return 7
}

func (p *P9208AlarmSign) getAlarmSignLen() int {
	switch p.ActiveSafetyType {
	case consts.ActiveSafetyJS:
		return 16
	case consts.ActiveSafetyHLJ:
		return 38
	case consts.ActiveSafetyGD:
		return 40
	case consts.ActiveSafetyHN:
		return 32
	case consts.ActiveSafetySC:
		return 39
	default:
	}
	return 16
}

func (p *P9208AlarmSign) String() string {
	return strings.Join([]string{
		fmt.Sprintf("\t报警标识 [%s]{ 默认使用苏标", p.ActiveSafetyType.String()),
		fmt.Sprintf("\t\t [%014x]终端ID:[%s]", p.TerminalID, p.TerminalID),
		fmt.Sprintf("\t\t [%012x]时间:[%s]", utils.Time2BCD(p.Time), p.Time),
		fmt.Sprintf("\t\t [%02x]序号:[%d]", p.SerialNumber, p.SerialNumber),
		fmt.Sprintf("\t\t [%02x]附件数量:[%d]", p.AttachNumber, p.AttachNumber),
		fmt.Sprintf("\t\t [%x]预留", p.AlarmReserve),
		"\t}",
	}, "\n")
}
