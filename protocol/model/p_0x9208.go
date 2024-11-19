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
		// TerminalID 终端ID
		TerminalID string `json:"terminalID"`
		// Time 时间 bcd[6]
		Time string `json:"time"`
		// SerialNumber 序号
		SerialNumber byte `json:"serialNumber"`
		// AttachNumber 附件数量
		AttachNumber byte `json:"attachNumber"`
		// AlarmReserve 预留
		AlarmReserve byte `json:"alarmReserve"`
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
	if len(body) < 53 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.ServerIPLen = body[0]
	k := int(p.ServerIPLen)
	if k+53 > len(body) {
		return protocol.ErrBodyLengthInconsistency
	}
	p.ServerAddr = string(body[1 : 1+k])
	p.TcpPort = binary.BigEndian.Uint16(body[1+k : 1+k+2])
	p.UdpPort = binary.BigEndian.Uint16(body[3+k : 3+k+2])
	p.P9208AlarmSign.parse(body[5+k : 5+k+16])
	p.AlarmID = string(body[21+k : 21+k+32])
	p.Reserve = body[53+k:]
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
		"\t报警标识{",
		fmt.Sprintf("\t\t [%014x]终端ID:[%s]", p.TerminalID, p.TerminalID),
		fmt.Sprintf("\t\t [%012x]时间:[%s]", utils.Time2BCD(p.Time), p.Time),
		fmt.Sprintf("\t\t [%02x]序号:[%d]", p.SerialNumber, p.SerialNumber),
		fmt.Sprintf("\t\t [%02x]附件数量:[%d]", p.AttachNumber, p.AttachNumber),
		fmt.Sprintf("\t\t [%02x]预留:[%x]", p.P9208AlarmSign.AlarmReserve, p.P9208AlarmSign.AlarmReserve),
		"\t}",
		fmt.Sprintf("\t [%064x]告警ID:[%s]", p.AlarmID, p.AlarmID),
		fmt.Sprintf("\t [%032x]预留:[%x]", p.Reserve, p.Reserve),
		"}",
	}, "\n")
}

func (p *P9208AlarmSign) parse(data []byte) {
	p.TerminalID = string(data[:7])
	p.Time = utils.BCD2Time(data[7 : 7+6])
	p.SerialNumber = data[13]
	p.AttachNumber = data[14]
	p.AlarmReserve = data[15]
}

func (p *P9208AlarmSign) encode() []byte {
	data := make([]byte, 16)
	copy(data[0:7], p.TerminalID)
	copy(data[7:13], utils.Time2BCD(p.Time))
	data[13] = p.SerialNumber
	data[14] = p.AttachNumber
	data[15] = p.AlarmReserve
	return data
}
