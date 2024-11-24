package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x1211 struct {
	BaseHandle
	// FileNameLen 文件名称长度
	FileNameLen byte `json:"fileNameLen"`
	// FileName 文件名称
	FileName string `json:"fileName"`
	// FileType 文件类型 0x00-图片 0x01-音频 0x02-视频 0x03-文本 0x04-其他
	FileType byte `json:"fileType"`
	// FileSize 当前文件大小 单位byte
	FileSize uint32 `json:"fileSize"`
}

func (t *T0x1211) Protocol() consts.JT808CommandType {
	return consts.T1211FileInfoUpload
}

func (t *T0x1211) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 6 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.FileNameLen = body[0]
	l := int(t.FileNameLen)
	if len(body) != 6+l {
		return protocol.ErrBodyLengthInconsistency
	}
	t.FileName = string(body[1 : 1+l])
	t.FileType = body[1+l]
	t.FileSize = binary.BigEndian.Uint32(body[2+l:])
	return nil
}

func (t *T0x1211) Encode() []byte {
	data := make([]byte, 1, 15)
	data[0] = t.FileNameLen
	data = append(data, t.FileName...)
	data = append(data, t.FileType)
	data = binary.BigEndian.AppendUint32(data, t.FileSize)
	return data
}

func (t *T0x1211) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", t.Protocol(), t.Encode()),
		fmt.Sprintf("\t[%02x] 文件名称长度:[%d]", t.FileNameLen, t.FileNameLen),
		fmt.Sprintf("\t[%x] 文件名称:[%s]", t.FileName, t.FileName),
		fmt.Sprintf("\t[%02x] 文件类型:[%d] 0x00-图片 0x01-音频 0x02-视频 0x03-文本 0x04-其他", t.FileType, t.FileType),
		fmt.Sprintf("\t[%02x] 当前文件大小:[%d] 单位byte", t.FileSize, t.FileSize),
		"}",
	}, "\n")
}
