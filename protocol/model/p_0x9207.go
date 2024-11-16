package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x9207 struct {
	BaseHandle
	// RespondSerialNumber 应答流水号 对应的平台文件上传消息的流水号
	RespondSerialNumber uint16 `json:"respondSerialNumber"`
	// UploadControl 上传控制 0-暂停 1-继续 2-取消
	UploadControl byte `json:"uploadControl"`
}

func (p *P0x9207) Protocol() consts.JT808CommandType {
	return consts.P9207FileUploadControl
}

func (p *P0x9207) ReplyProtocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (p *P0x9207) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) != 3 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.RespondSerialNumber = binary.BigEndian.Uint16(body[0:2])
	p.UploadControl = body[2]
	return nil
}

func (p *P0x9207) Encode() []byte {
	data := make([]byte, 3)
	binary.BigEndian.PutUint16(data, p.RespondSerialNumber)
	data[2] = p.UploadControl
	return data
}

func (p *P0x9207) HasReply() bool {
	return false
}

func (p *P0x9207) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%04x] 应答流水号:[%d]", p.RespondSerialNumber, p.RespondSerialNumber),
		fmt.Sprintf("\t[%02x] 上传控制:[%d] 0-暂停 1-继续 2-取消", p.UploadControl, p.UploadControl),
		"}",
	}, "\n")
}
