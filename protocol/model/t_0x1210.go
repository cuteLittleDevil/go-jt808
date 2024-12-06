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
		// TerminalID 终端ID byte[7] 苏标-终端7 黑标-0 广东标-终端30 湖南标-终端7 四川标-终端30
		TerminalID string `json:"terminalID"`
		// P9208AlarmSign 报警标识号 苏标-16 黑标-38 广东标-40 湖南标-16 四川标-39
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
	idLen := t.P9208AlarmSign.getTerminalIDLen()
	alarmSignLen := t.P9208AlarmSign.getAlarmSignLen()
	if t.P9208AlarmSign.ActiveSafetyType == consts.ActiveSafetyHLJ { // 黑标的情况是没有终端ID
		idLen = 0
	}
	if len(body) < idLen+alarmSignLen+32+1+1 {
		return protocol.ErrBodyLengthInconsistency
	}
	cursor := idLen
	if idLen > 0 {
		t.TerminalID = string(bytes.Trim(body[0:cursor], "\x00"))
	}
	t.P9208AlarmSign.parse(body[cursor : cursor+alarmSignLen])
	cursor += alarmSignLen
	t.AlarmID = string(bytes.Trim(body[cursor:cursor+32], "\x00"))
	cursor += 32
	t.InfoType = body[cursor]
	t.AttachCount = body[cursor+1]
	cursor += 2
	if len(body) < cursor+int(t.AttachCount)*(1+1+4) {
		return protocol.ErrBodyLengthInconsistency
	}
	start := cursor
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
	data := make([]byte, 0, 60)
	if t.P9208AlarmSign.ActiveSafetyType != consts.ActiveSafetyHLJ {
		data = append(data, utils.String2FillingBytes(t.TerminalID, t.P9208AlarmSign.getTerminalIDLen())...)
	}
	data = append(data, t.P9208AlarmSign.encode()...)
	data = append(data, utils.String2FillingBytes(t.AlarmID, 32)...)
	data = append(data, t.InfoType)
	data = append(data, t.AttachCount)
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
		fmt.Sprintf("数据体对象:{"),
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
