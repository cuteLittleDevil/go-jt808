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
	P0x8302 struct {
		BaseHandle
		// Flag 标志位
		Flag byte `json:"flag"`
		// QuestionContentLen 问题内容长度
		QuestionContentLen byte `json:"questionContentLen"`
		// QuestionContent 问题内容 GBK编码发送给终端
		QuestionContent string `json:"questionContent"`
		// AnswerList 答案列表
		AnswerList []P0x8302Answer `json:"answerList"`
		// P0x8302TextFlagDetails 提问下发标志位定义见表43
		P0x8302TextFlagDetails
	}

	P0x8302Answer struct {
		// AnswerID 答案ID
		AnswerID byte `json:"answerID"`
		// AnswerContentLen 答案内容长度
		AnswerContentLen uint16 `json:"answerContentLen"`
		// AnswerContent 答案内容 GBK编码发送给终端
		AnswerContent string `json:"answerContent"`
	}

	P0x8302TextFlagDetails struct {
		// Urgent 紧急
		Urgent bool `json:"urgent"`
		// Bit1Reserve 保留
		Bit1Reserve bool `json:"bit1Reserve"`
		// Bit2Reserve 保留
		Bit2Reserve bool `json:"bit2Reserve"`
		// TTS 终端TTS播读
		TTS bool `json:"tts"`
		// AdvertisingScreen 广告屏显示
		AdvertisingScreen bool `json:"advertisingScreen"`
		// Bit5Reserve 保留
		Bit5Reserve bool `json:"bit5Reserve"`
		// Bit6Reserve 保留
		Bit6Reserve bool `json:"bit6Reserve"`
		// Bit7Reserve 保留
		Bit7Reserve bool `json:"bit7Reserve"`
	}
)

func (p *P0x8302) Protocol() consts.JT808CommandType {
	return consts.P8302QuestionDistribution
}

func (p *P0x8302) ReplyProtocol() consts.JT808CommandType {
	return consts.T0302QuestionAnswer
}

func (p *P0x8302) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 2 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.Flag = body[0]
	p.QuestionContentLen = body[1]
	if len(body) < 2+int(p.QuestionContentLen)+3 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.QuestionContent = string(utils.GBK2UTF8(body[2 : 2+p.QuestionContentLen]))
	start := 2 + int(p.QuestionContentLen)
	for {
		answer := P0x8302Answer{
			AnswerID:         body[start],
			AnswerContentLen: binary.BigEndian.Uint16(body[start+1 : start+3]),
		}
		end := start + 3 + int(answer.AnswerContentLen)
		if len(body) < end {
			return protocol.ErrBodyLengthInconsistency
		}
		answer.AnswerContent = string(utils.GBK2UTF8(body[start+3 : end]))
		p.AnswerList = append(p.AnswerList, answer)
		if end >= len(body) {
			break
		}
		start = end
	}
	p.P0x8302TextFlagDetails.parse(p.Flag)
	return nil
}

func (p *P0x8302) Encode() []byte {
	data := make([]byte, 2, 20)
	if p.Flag == 0 {
		p.Flag = p.toFlag()
	}
	data[0] = p.Flag
	data[1] = p.QuestionContentLen
	data = append(data, utils.UTF82GBK([]byte(p.QuestionContent))...)
	for _, v := range p.AnswerList {
		data = append(data, v.AnswerID)
		data = binary.BigEndian.AppendUint16(data, v.AnswerContentLen)
		data = append(data, utils.UTF82GBK([]byte(v.AnswerContent))...)
	}
	return data
}

func (p *P0x8302) HasReply() bool {
	return false
}

func (p *P0x8302) String() string {
	str := "\t答案列表:"
	for _, v := range p.AnswerList {
		str += fmt.Sprintf("\n\t\t[%02x] 答案ID:[%d]\n", v.AnswerID, v.AnswerID)
		str += fmt.Sprintf("\t\t[%04x] 答案内容长度:[%d]\n", v.AnswerContentLen, v.AnswerContentLen)
		str += fmt.Sprintf("\t\t[%x] 答案内容:[%s]", v.AnswerContent, v.AnswerContent)
	}
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%02x] 标志位:[%d]", p.Flag, p.Flag),
		p.P0x8302TextFlagDetails.String(),
		fmt.Sprintf("\t[%02x] 问题内容长度:[%d]", p.QuestionContentLen, p.QuestionContentLen),
		fmt.Sprintf("\t[%s] 问题内容:[%s]", p.QuestionContent, p.QuestionContent),
		str,
		"}",
	}, "\n")
}

func (d *P0x8302TextFlagDetails) toFlag() byte {
	flag := byte(0)
	if d.Urgent {
		flag |= 1 << 0
	}
	if d.Bit1Reserve {
		flag |= 1 << 1
	}
	if d.Bit2Reserve {
		flag |= 1 << 2
	}
	if d.TTS {
		flag |= 1 << 3
	}
	if d.AdvertisingScreen {
		flag |= 1 << 4
	}
	if d.Bit5Reserve {
		flag |= 1 << 5
	}
	if d.Bit6Reserve {
		flag |= 1 << 6
	}
	if d.Bit7Reserve {
		flag |= 1 << 7
	}
	return flag
}

func (d *P0x8302TextFlagDetails) parse(flag byte) {
	data := fmt.Sprintf("%.8b", flag)
	if data[7] == '1' {
		d.Urgent = true
	}
	if data[6] == '1' {
		d.Bit1Reserve = true
	}
	if data[5] == '1' {
		d.Bit2Reserve = true
	}
	if data[4] == '1' {
		d.TTS = true
	}
	if data[3] == '1' {
		d.AdvertisingScreen = true
	}
	if data[2] == '1' {
		d.Bit5Reserve = true
	}
	if data[1] == '1' {
		d.Bit6Reserve = true
	}
	if data[0] == '1' {
		d.Bit7Reserve = true
	}
}

func (d *P0x8302TextFlagDetails) String() string {
	return strings.Join([]string{
		fmt.Sprintf("\t\t[bit0]紧急:[%t]", d.Urgent),
		fmt.Sprintf("\t\t[bit1]保留:[%t]", d.Bit1Reserve),
		fmt.Sprintf("\t\t[bit2]保留:[%t]", d.Bit2Reserve),
		fmt.Sprintf("\t\t[bit3]终端TTS播读:[%t]", d.TTS),
		fmt.Sprintf("\t\t[bit4]广告屏显示:[%t]", d.AdvertisingScreen),
		fmt.Sprintf("\t\t[bit5]保留:[%t]", d.Bit5Reserve),
		fmt.Sprintf("\t\t[bit6]保留:[%t]", d.Bit6Reserve),
		fmt.Sprintf("\t\t[bit7]保留:[%t]", d.Bit7Reserve),
	}, "\n")
}
