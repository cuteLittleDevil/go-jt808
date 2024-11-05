package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x1003 struct {
	BaseHandle
	// EnterAudioEncoding 输入音频编码方式 见表12
	EnterAudioEncoding byte `json:"enterAudioEncoding"`
	// EnterAudioChannelsNumber 输入音频声道数
	EnterAudioChannelsNumber byte `json:"enterAudioChannelsNumber"`
	// EnterAudioSampleRate 输入音频采样率 0-8kHz 1-22.05 kHz 2-44.1 kHz 3-48 kHz
	EnterAudioSampleRate byte `json:"enterAudioSampleRate"`
	// EnterAudioSampleDigits 输入音频采样位数 0-8位 1-16位 2-32位
	EnterAudioSampleDigits byte `json:"enterAudioSampleDigits"`
	// AudioFrameLength 音频帧长度 范围1-4294967295
	AudioFrameLength uint16 `json:"audioFrameLength"`
	// HasSupportedAudioOutput 是否支持音频输出 0-不支持 1-支持
	HasSupportedAudioOutput byte `json:"hasSupportedAudioOutput"`
	// VideoEncoding 视频编码方式 见表19
	VideoEncoding byte `json:"videoEncoding"`
	// TerminalSupportedMaxNumberOfAudioPhysicalChannels 终端支持的最大音频物理通道数量
	TerminalSupportedMaxNumberOfAudioPhysicalChannels byte `json:"terminalSupportedMaxNumberOfAudioPhysicalChannels"`
	// TerminalSupportedMaxNumberOfVideoPhysicalChannels  终端支持的最大视频物理通道数量
	TerminalSupportedMaxNumberOfVideoPhysicalChannels byte `json:"terminalSupportedMaxNumberOfVideoPhysicalChannels"`
}

func (t *T0x1003) Protocol() consts.JT808CommandType {
	return consts.T1003UploadAudioVideoAttr
}

func (t *T0x1003) ReplyProtocol() consts.JT808CommandType {
	return 0
}

func (t *T0x1003) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) != 10 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.EnterAudioEncoding = body[0]
	t.EnterAudioChannelsNumber = body[1]
	t.EnterAudioSampleRate = body[2]
	t.EnterAudioSampleDigits = body[3]
	t.AudioFrameLength = binary.BigEndian.Uint16(body[4:6])
	t.HasSupportedAudioOutput = body[6]
	t.VideoEncoding = body[7]
	t.TerminalSupportedMaxNumberOfAudioPhysicalChannels = body[8]
	t.TerminalSupportedMaxNumberOfVideoPhysicalChannels = body[9]
	return nil
}

func (t *T0x1003) Encode() []byte {
	data := make([]byte, 10)
	data[0] = t.EnterAudioEncoding
	data[1] = t.EnterAudioChannelsNumber
	data[2] = t.EnterAudioSampleRate
	data[3] = t.EnterAudioSampleDigits
	binary.BigEndian.PutUint16(data[4:6], t.AudioFrameLength)
	data[6] = t.HasSupportedAudioOutput
	data[7] = t.VideoEncoding
	data[8] = t.TerminalSupportedMaxNumberOfAudioPhysicalChannels
	data[9] = t.TerminalSupportedMaxNumberOfVideoPhysicalChannels
	return data
}

func (t *T0x1003) ReplyBody(_ *jt808.JTMessage) ([]byte, error) {
	return nil, nil
}

func (t *T0x1003) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", t.Protocol(), t.Encode()),
		fmt.Sprintf("\t[%02x] 输入音频编码方式:[%d]", t.EnterAudioEncoding, t.EnterAudioEncoding),
		fmt.Sprintf("\t[%02x] 输入音频声道数:[%d]", t.EnterAudioChannelsNumber, t.EnterAudioChannelsNumber),
		fmt.Sprintf("\t[%02x] 输入音频采样率:[%d] 0-8 1-22.05 2-44.1 3-48 单位kHz", t.EnterAudioSampleRate, t.EnterAudioSampleRate),
		fmt.Sprintf("\t[%02x] 输入音频采样位数:[%d] 0-8位 1-16位 2-32位", t.EnterAudioSampleDigits, t.EnterAudioSampleDigits),
		fmt.Sprintf("\t[%04x] 音频帧长度:[%d]", t.AudioFrameLength, t.AudioFrameLength),
		fmt.Sprintf("\t[%02x] 是否支持音频输出:[%d] 0-不支持 1-支持", t.HasSupportedAudioOutput, t.HasSupportedAudioOutput),
		fmt.Sprintf("\t[%02x] 视频编码方式:[%d]", t.VideoEncoding, t.VideoEncoding),
		fmt.Sprintf("\t[%02x] 终端支持的最大音频物理通道数量:[%d]", t.TerminalSupportedMaxNumberOfAudioPhysicalChannels, t.TerminalSupportedMaxNumberOfAudioPhysicalChannels),
		fmt.Sprintf("\t[%02x] 终端支持的最大视频物理通道数量:[%d]", t.TerminalSupportedMaxNumberOfVideoPhysicalChannels, t.TerminalSupportedMaxNumberOfVideoPhysicalChannels),
		"}",
	}, "\n")
}
