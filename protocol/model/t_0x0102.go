package model

import (
	"bytes"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x0102 struct {
	BaseHandle
	// AuthCodeLen 鉴权码长度 2019版本
	AuthCodeLen uint8 `json:"authCodeLen"`
	// AuthCode 鉴权码 终端重连后上报鉴权码
	AuthCode string `json:"authCode"`
	// TerminalIMEI 终端IMEI 2019版本 Byte[15]
	TerminalIMEI string `json:"terminalIMEI"`
	// SoftwareVersion 软件版本 2019版本 Byte[20]
	SoftwareVersion string `json:"softwareVersion"`
	// Version 版本 1-2011 2-2013 3-2019
	Version consts.ProtocolVersionType `json:"version"`
}

func (t *T0x0102) Protocol() consts.JT808CommandType {
	return consts.T0102RegisterAuth
}

func (t *T0x0102) Parse(jtMsg *jt808.JTMessage) error {
	version := consts.JT808Protocol2013
	if jtMsg.Header.ProtocolVersion == consts.JT808Protocol2019 {
		version = consts.JT808Protocol2019
	}
	t.Version = version

	body := jtMsg.Body
	if t.Version == consts.JT808Protocol2019 {
		if len(body) < 1+15+20 {
			return protocol.ErrBodyLengthInconsistency
		}
		t.AuthCodeLen = body[0]
		if len(body) < 1+int(t.AuthCodeLen)+15+20 {
			return protocol.ErrBodyLengthInconsistency
		}
		t.AuthCode = string(body[1 : 1+t.AuthCodeLen])
		t.TerminalIMEI = string(body[1+t.AuthCodeLen : 1+t.AuthCodeLen+15])
		data := body[1+t.AuthCodeLen+15 : 1+t.AuthCodeLen+15+20]
		if index := bytes.IndexByte(data, 0x00); index != -1 {
			data = data[:index]
		}
		t.SoftwareVersion = string(data)
	} else {
		t.AuthCode = string(body)
	}
	return nil
}

func (t *T0x0102) Encode() []byte {
	data := make([]byte, 0, 10)
	if t.Version == consts.JT808Protocol2019 {
		data = append(data, t.AuthCodeLen)
		data = append(data, []byte(t.AuthCode)...)
		data = append(data, []byte(t.TerminalIMEI)...)
		data = append(data, utils.String2FillingBytes(t.SoftwareVersion, 20)...)
	} else {
		data = append(data, []byte(t.AuthCode)...)
	}
	return data
}

func (t *T0x0102) ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error) {
	if err := t.Parse(jtMsg); err != nil {
		return nil, err
	}
	// 鉴权码=手机号
	result := byte(1)
	if jtMsg.Header.TerminalPhoneNo == t.AuthCode {
		result = 0
	}
	head := jtMsg.Header
	// 通用应答
	p8001 := &P0x8001{
		RespondSerialNumber: head.SerialNumber,
		RespondID:           head.ID,
		Result:              result, // 0-成功 1-失败 2-消息有误 3-不支持 4-报警处理确认 默认成功
	}
	return p8001.Encode(), nil
}

func (t *T0x0102) String() string {
	str := "数据体对象:{\n"
	body := t.Encode()
	str += fmt.Sprintf("\t%s:[%x]\n", t.Protocol(), body)
	if t.Version == consts.JT808Protocol2019 {
		str += fmt.Sprintf("\t[%02x] 鉴权码长度:[%d]\n", t.AuthCodeLen, t.AuthCodeLen)
		str += fmt.Sprintf("\t[%x] 鉴权码:[%s]\n", t.AuthCode, t.AuthCode)
		str += fmt.Sprintf("\t[%015x] 终端IMEI:[%s]\n", t.TerminalIMEI, t.TerminalIMEI)
		str += fmt.Sprintf("\t[%020x]软件版本:[%s]\n", body[1+t.AuthCodeLen+15:], t.SoftwareVersion)
	} else {
		str += fmt.Sprintf("\t鉴权码:[%s]\n", t.AuthCode)
	}
	return strings.Join([]string{
		str,
		"}",
	}, "\n")
}
