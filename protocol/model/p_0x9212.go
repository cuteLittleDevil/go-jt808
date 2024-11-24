package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type (
	P0x9212 struct {
		BaseHandle
		// FileNameLen 文件名称长度
		FileNameLen byte `json:"fileNameLen"`
		// FileName 文件名称
		FileName string `json:"fileName"`
		// FileType 文件类型 0x00-图片 0x01-音频 0x02-视频 0x03-文本 0x04-其他
		FileType byte `json:"fileType"`
		// UploadResult 上传结果 0-完成 1-需要补传
		UploadResult byte `json:"uploadResult"`
		// RetransmitPacketNumber  补传数据包数量 无补传时为0
		RetransmitPacketNumber byte `json:"retransmitPacketNumber"`
		// P0x9212RetransmitPacketList 补传数据包列表
		P0x9212RetransmitPacketList []P0x9212RetransmitPacket `json:"p0X9212RetransmitPacketList"`
	}

	P0x9212RetransmitPacket struct {
		// DataOffset 数据偏移量
		DataOffset uint32 `json:"dataOffset"`
		// DataLength 数据长度
		DataLength uint32 `json:"dataLength"`
	}
)

func (p *P0x9212) Protocol() consts.JT808CommandType {
	return consts.P9212FileUploadCompleteRespond
}

func (p *P0x9212) ReplyProtocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (p *P0x9212) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 4 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.FileNameLen = body[0]
	l := int(p.FileNameLen)
	if len(body) < 4+l {
		return protocol.ErrBodyLengthInconsistency
	}
	p.FileName = string(body[1 : 1+l])
	p.FileType = body[1+l]
	p.UploadResult = body[2+l]
	p.RetransmitPacketNumber = body[3+l]
	if len(body) != 4+l+8*int(p.RetransmitPacketNumber) {
		return protocol.ErrBodyLengthInconsistency
	}
	for i := 0; i < int(p.RetransmitPacketNumber); i++ {
		p.P0x9212RetransmitPacketList = append(p.P0x9212RetransmitPacketList, P0x9212RetransmitPacket{
			DataOffset: binary.BigEndian.Uint32(body[4+l+2*i:]),
			DataLength: binary.BigEndian.Uint32(body[4+l+2*i+4:]),
		})
	}
	return nil
}

func (p *P0x9212) Encode() []byte {
	data := make([]byte, 1, 10)
	data[0] = p.FileNameLen
	data = append(data, p.FileName...)
	data = append(data, p.FileType)
	data = append(data, p.UploadResult)
	data = append(data, p.RetransmitPacketNumber)
	for _, v := range p.P0x9212RetransmitPacketList {
		data = binary.BigEndian.AppendUint32(data, v.DataOffset)
		data = binary.BigEndian.AppendUint32(data, v.DataLength)
	}
	return data
}

func (p *P0x9212) HasReply() bool {
	return false
}

func (p *P0x9212) String() string {
	str := fmt.Sprintf("\t[%x] 补传数据包数量:[%d] 无补传时为0", p.RetransmitPacketNumber, p.RetransmitPacketNumber)
	for _, v := range p.P0x9212RetransmitPacketList {
		str += fmt.Sprintf("\n\t\t[%08x] 数据偏移量:[%d]\n", v.DataOffset, v.DataOffset)
		str += fmt.Sprintf("\t\t[%08x] 数据长度:[%d]", v.DataLength, v.DataLength)
	}
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%02x] 文件名称长度:[%d]", p.FileNameLen, p.FileNameLen),
		fmt.Sprintf("\t[%x] 文件名称:[%s]", p.FileName, p.FileName),
		fmt.Sprintf("\t[%02x] 文件类型:[%d] 0x00-图片 0x01-音频 0x02-视频 0x03-文本 0x04-其他", p.FileType, p.FileType),
		fmt.Sprintf("\t[%x] 上传结果:[%d] 0-完成 1-需要补传", p.UploadResult, p.UploadResult),
		str,
		"}",
	}, "\n")
}
