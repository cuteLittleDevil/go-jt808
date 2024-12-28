package model

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type (
	P0x8300 struct {
		BaseHandle
		// Flag 文本标志 见表38 标志BIT位含义 [0: 1-紧急] [1: 保留] [2: 1-终端显示器显示] [3: 1-终端TTS播读]
		// 4[1-广告屏显示] 5[0-中心导航信息 1-CAN故障码信息] 6-7保留
		Flag byte `json:"textFlag"`
		// Text 文本信息 最长为1024字节(平台限制1000字节) GBK编码发送给终端
		Text string `json:"text"`
		// P0x8300TextFlagDetails 文本标志详情
		P0x8300TextFlagDetails
	}

	P0x8300TextFlagDetails struct {
		// Urgent 紧急
		Urgent bool `json:"urgent"`
		// Bit1Reserve 保留
		Bit1Reserve bool `json:"bit1Reserve"`
		// Display 终端显示器显示
		Display bool `json:"display"`
		// TTS 终端TTS播读
		TTS bool `json:"tts"`
		// AdvertisingScreen 广告屏显示
		AdvertisingScreen bool `json:"advertisingScreen"`
		// InfoCategory 信息类别 0-中心导航信息 1-CAN故障码信息
		InfoCategory int `json:"can"`
		// Bit6Reserve 保留
		Bit6Reserve bool `json:"bit6Reserve"`
		// Bit7Reserve 保留
		Bit7Reserve bool `json:"bit7Reserve"`
	}
)

func (p *P0x8300) Protocol() consts.JT808CommandType {
	return consts.P8300TextInfoDistribution
}

func (p *P0x8300) ReplyProtocol() consts.JT808CommandType {
	return consts.T0001GeneralRespond
}

func (p *P0x8300) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 2 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.Flag = body[0]
	p.Text = string(utils.GBK2UTF8(body[1:]))
	p.P0x8300TextFlagDetails.parse(p.Flag)
	return nil
}

func (p *P0x8300) Encode() []byte {
	if p.Flag == 0 {
		p.Flag = p.toFlag()
	}
	data := make([]byte, 1, 1024)
	data[0] = p.Flag
	text := []byte(p.Text)
	gbk := utils.UTF82GBK(text)
	if len(gbk) > 1000 {
		gbk = gbk[:1000]
	}
	data = append(data, gbk...)
	return data
}

func (p *P0x8300) HasReply() bool {
	return false
}

func (p *P0x8300) String() string {
	text := p.Text
	if len([]byte(text)) > 1000 {
		text = string([]byte(text)[:1000])
	}
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%02x] 文本标志:[%d]", p.Flag, p.Flag),
		p.P0x8300TextFlagDetails.String(),
		fmt.Sprintf("\t[%x] 有效文本:[%s]", text, text),
		"}",
	}, "\n")
}

func (d *P0x8300TextFlagDetails) toFlag() byte {
	flag := byte(0)
	if d.Urgent {
		flag |= 1 << 0
	}
	if d.Bit1Reserve {
		flag |= 1 << 1
	}
	if d.Display {
		flag |= 1 << 2
	}
	if d.TTS {
		flag |= 1 << 3
	}
	if d.AdvertisingScreen {
		flag |= 1 << 4
	}
	if d.InfoCategory == 1 {
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

func (d *P0x8300TextFlagDetails) parse(textFlag byte) {
	data := fmt.Sprintf("%.8b", textFlag)
	if data[7] == '1' {
		d.Urgent = true
	}
	if data[6] == '1' {
		d.Bit1Reserve = true
	}
	if data[5] == '1' {
		d.Display = true
	}
	if data[4] == '1' {
		d.TTS = true
	}
	if data[3] == '1' {
		d.AdvertisingScreen = true
	}
	if data[2] == '1' {
		d.InfoCategory = 1
	}
	if data[1] == '1' {
		d.Bit6Reserve = true
	}
	if data[0] == '1' {
		d.Bit7Reserve = true
	}
}

func (d *P0x8300TextFlagDetails) String() string {
	return strings.Join([]string{
		fmt.Sprintf("\t\t[bit0]紧急:[%t]", d.Urgent),
		fmt.Sprintf("\t\t[bit1]保留:[%t]", d.Bit1Reserve),
		fmt.Sprintf("\t\t[bit2]终端显示器显示:[%t]", d.Display),
		fmt.Sprintf("\t\t[bit3]终端TTS播读:[%t]", d.TTS),
		fmt.Sprintf("\t\t[bit4]广告屏显示:[%t]", d.AdvertisingScreen),
		fmt.Sprintf("\t\t[bit5]信息类别:[%d] 0-中心导航信息 1-CAN故障码信息", d.InfoCategory),
		fmt.Sprintf("\t\t[bit6]保留:[%t]", d.Bit6Reserve),
		fmt.Sprintf("\t\t[bit7]保留:[%t]", d.Bit7Reserve),
	}, "\n")
}
