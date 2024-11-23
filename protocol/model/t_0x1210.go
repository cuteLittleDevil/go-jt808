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
	T0x1210 struct {
		BaseHandle
		// TerminalID 终端ID byte[7]
		TerminalID string `json:"terminalID"`
		// P9208AlarmSign 报警标识号 byte[16]
		P9208AlarmSign `json:"p9208AlarmSign"`
		// AlarmID 平台给报警分配的唯一编号 byte[32]
		AlarmID string `json:"alarmID"`
		// InfoType 信息类型 0x00-正常报警文件信息 0x01-补传报警文件信息
		InfoType byte `json:"infoType"`
		// AttachCount 附件数量
		AttachCount byte `json:"attachCount"`
		// T0x1210AlarmItemList 附件信息列表
		T0x1210AlarmItemList []T0x1210AlarmItem `json:"t1210AlarmItemList"`
	}

	T0x1210AlarmItem struct {
		// FileNameLen 文件名称长度
		FileNameLen byte `json:"fileNameLen"`
		// FileName 文件名称
		FileName string `json:"fileName"`
		// FileSize 当前文件大小 单位byte
		FileSize uint32 `json:"fileSize"`
	}
)

func (t *T0x1210) Protocol() consts.JT808CommandType {
	return consts.T1210AlarmAttachInfoMessage
}

func (t *T0x1210) ReplyProtocol() consts.JT808CommandType {
	return consts.P8001GeneralRespond
}

func (t *T0x1210) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 57 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.TerminalID = string(bytes.Trim(body[0:7], "\x00"))
	t.P9208AlarmSign.parse(body[7:23])
	t.AlarmID = string(bytes.Trim(body[23:55], "\x00"))
	t.InfoType = body[55]
	t.AttachCount = body[56]
	if len(body) < 57+int(t.AttachCount)*(1+1+4) {
		return protocol.ErrBodyLengthInconsistency
	}
	start := 57
	for i := 0; i < int(t.AttachCount); i++ {
		fileNameLen := body[start]
		if len(body) < start+1+int(fileNameLen)+4 {
			return protocol.ErrBodyLengthInconsistency
		}
		fileName := string(body[start+1 : start+1+int(fileNameLen)])
		fileSize := binary.BigEndian.Uint32(body[start+1+int(fileNameLen):])
		start = start + 1 + int(fileNameLen) + 4
		t.T0x1210AlarmItemList = append(t.T0x1210AlarmItemList, T0x1210AlarmItem{
			FileNameLen: fileNameLen,
			FileName:    fileName,
			FileSize:    fileSize,
		})
	}
	return nil
}

func (t *T0x1210) Encode() []byte {
	data := make([]byte, 57, 60)
	copy(data[0:7], utils.String2FillingBytes(t.TerminalID, 7))
	copy(data[7:23], t.P9208AlarmSign.encode())
	copy(data[23:55], utils.String2FillingBytes(t.AlarmID, 32))
	data[55] = t.InfoType
	data[56] = t.AttachCount
	for _, v := range t.T0x1210AlarmItemList {
		data = append(data, v.FileNameLen)
		data = append(data, []byte(v.FileName)...)
		data = binary.BigEndian.AppendUint32(data, v.FileSize)
	}
	return data
}

func (t *T0x1210) String() string {
	items := "\t附件信息列表:\n"
	for _, v := range t.T0x1210AlarmItemList {
		items += "\t{\n"
		items += fmt.Sprintf("\t\t文件名称长度:[%d]\n", v.FileNameLen)
		items += fmt.Sprintf("\t\t文件名称:[%s]\n", v.FileName)
		items += fmt.Sprintf("\t\t当前文件大小:[%d]单位byte\n", v.FileSize)
		items += "\t}\n"
	}
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", t.Protocol(), t.Encode()),
		fmt.Sprintf("\t[%014x] 终端ID:[%s]", t.TerminalID, t.TerminalID),
		t.P9208AlarmSign.String(),
		fmt.Sprintf("\t[%064x] 平台给报警分配的唯一编号:[%s]", t.AlarmID, t.AlarmID),
		fmt.Sprintf("\t[%02x] 信息类型:[%d] 0x00-正常报警文件信息 0x01-补传报警文件信息", t.InfoType, t.InfoType),
		fmt.Sprintf("\t[%02x] 附件数量:[%d]", t.AttachCount, t.AttachCount),
		items,
		"}",
	}, "\n")
}
