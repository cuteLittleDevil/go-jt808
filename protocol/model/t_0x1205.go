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
	T0x1205 struct {
		BaseHandle
		// SerialNumber 流水号
		SerialNumber uint16 `json:"serialNumber"`
		// AudioVideoResourceTotal 音视频资源总数
		AudioVideoResourceTotal uint32 `json:"audioVideoResourceTotal"`
		// AudioVideoResourceList 音视频资源列表
		AudioVideoResourceList []T0x1205AudioVideoResource `json:"audioVideoResourceList"`
	}

	T0x1205AudioVideoResource struct {
		// ChannelNo 逻辑通道号 按照 JT/T 1076—2016中的表2
		ChannelNo byte `json:"logicChannelNo"`
		// StartTime 开始时间 YY-MM-DD-HH-MM-SS
		StartTime string `json:"startTime"`
		// EndTime 结束时间 YY-MM-DD-HH-MM-SS
		EndTime string `json:"endTime"`
		// AlarmFlag 报警标志 bit0 ~ bit31 按照 JT/T 808—2011 的表18 bit32~bi63见表13
		AlarmFlag uint64 `json:"alarmFlag"`
		// AudioVideoResourceType 音视频资源类型 0-音视频 1-音频 2-视频
		AudioVideoResourceType byte `json:"audioVideoResourceType"`
		// StreamType 码流类型 1-主码流 2-子码流
		StreamType byte `json:"streamType"`
		// MemoryType 存储器类型 1-主存储器 2-灾备存储器
		MemoryType byte `json:"memoryType"`
		// FileSizeByte 文件大小 单位字节(BYTE)
		FileSizeByte uint32 `json:"fileSize"`
	}
)

func (t *T0x1205) Protocol() consts.JT808CommandType {
	return consts.T1205UploadAudioVideoResourceList
}

func (t *T0x1205) ReplyProtocol() consts.JT808CommandType {
	return 0
}

func (t *T0x1205) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 6 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.SerialNumber = binary.BigEndian.Uint16(body[0:2])
	t.AudioVideoResourceTotal = binary.BigEndian.Uint32(body[2:6])
	if len(body) != 6+int(t.AudioVideoResourceTotal)*28 {
		return protocol.ErrBodyLengthInconsistency
	}
	start, end := 6, 6+28
	for i := 0; i < int(t.AudioVideoResourceTotal); i++ {
		curData := body[start:end]
		t.AudioVideoResourceList = append(t.AudioVideoResourceList, T0x1205AudioVideoResource{
			ChannelNo:              curData[0],
			StartTime:              utils.BCD2Time(curData[1:7]),
			EndTime:                utils.BCD2Time(curData[7:13]),
			AlarmFlag:              binary.BigEndian.Uint64(curData[13:21]),
			AudioVideoResourceType: curData[21],
			StreamType:             curData[22],
			MemoryType:             curData[23],
			FileSizeByte:           binary.BigEndian.Uint32(curData[24:28]),
		})
		start = end
		end = start + 28
	}
	return nil
}

func (t *T0x1205) Encode() []byte {
	data := make([]byte, 6, 100)
	binary.BigEndian.PutUint16(data[:2], t.SerialNumber)
	binary.BigEndian.PutUint32(data[2:6], t.AudioVideoResourceTotal)
	for _, v := range t.AudioVideoResourceList {
		data = append(data, v.ChannelNo)
		data = append(data, utils.Time2BCD(v.StartTime)...)
		data = append(data, utils.Time2BCD(v.EndTime)...)
		binary.BigEndian.AppendUint64(data, v.AlarmFlag)
		data = append(data, v.AudioVideoResourceType)
		data = append(data, v.StreamType)
		data = append(data, v.MemoryType)
		binary.BigEndian.AppendUint32(data, v.FileSizeByte)
	}
	return data
}

func (t *T0x1205) HasReply() bool {
	return false
}

func (t *T0x1205) String() string {
	str := ""
	for _, v := range t.AudioVideoResourceList {
		str += "\t{\n"
		str += fmt.Sprintf("\t\t[%02x] 逻辑通道号:[%d]\n", v.ChannelNo, v.ChannelNo)
		str += fmt.Sprintf("\t\t[%012x] 开始时间:[%s]\n", utils.Time2BCD(v.StartTime), v.StartTime)
		str += fmt.Sprintf("\t\t[%012x] 结束时间:[%s]\n", utils.Time2BCD(v.EndTime), v.EndTime)
		str += fmt.Sprintf("\t\t[%016x] 报警标志:[%d]\n", v.AlarmFlag, v.AlarmFlag)
		str += fmt.Sprintf("\t\t[%02x] 音视频资源类型:[%d] 0-音视频 1-音频 2-视频\n", v.AudioVideoResourceType, v.AudioVideoResourceType)
		str += fmt.Sprintf("\t\t[%02x] 码流类型:[%d] 1-主码流 2-子码流\n", v.StreamType, v.StreamType)
		str += fmt.Sprintf("\t\t[%02x] 存储器类型:[%d] 1-主存储器 2-灾备存储器\n", v.MemoryType, v.MemoryType)
		str += fmt.Sprintf("\t\t[%08x] 文件大小:[%d] 单位字节(BYTE)\n", v.FileSizeByte, v.FileSizeByte)
		str += "\t}\n"
	}
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", t.Protocol(), t.Encode()),
		fmt.Sprintf("\t[%04x]流水号:[%d]", t.SerialNumber, t.SerialNumber),
		fmt.Sprintf("\t[%08x]音视频资源总数:[%d]", t.AudioVideoResourceTotal, t.AudioVideoResourceTotal),
		str,
		"}",
	}, "\n")
}
