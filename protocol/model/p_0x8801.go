package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x8801 struct {
	BaseHandle
	// ChannelID 通道ID
	ChannelID byte `json:"channelID"`
	// ShootCommand 拍摄命令 0 表示停止拍摄；0xFFFF 表示录像；其它表示拍照张数
	ShootCommand uint16 `json:"shootCommand"`
	// PhotoIntervalOrVideoTime 拍照间隔/录像时间 单位秒0表示按最小间隔拍照或一直录像
	PhotoIntervalOrVideoTime uint16 `json:"photoIntervalOrVideoTime"`
	// SaveFlag 保存标志 1-保存 0-实时上传
	SaveFlag byte `json:"saveFlag"`
	// Resolution 分辨率 0x01-320*240 0x02-640*480 0x03-800*600 0x04-1024*768
	// 0x05-176*144[Qcif] 0x06-352*288[Cif] 0x07-704*288[HALF D1] 0x08-704*576[D1]
	Resolution byte `json:"resolution"`
	// VideoQuality  图像/视频质量 1-10 1-代表质量损失最小 10-表示压缩比最大
	VideoQuality byte `json:"videoQuality"`
	// Intensity 亮度 0-255
	Intensity byte `json:"intensity"`
	// Contrast 对比度 0-127
	Contrast byte `json:"contrast"`
	// Saturation 饱和度 0-127
	Saturation byte `json:"saturation"`
	// Chroma 色度 0-255
	Chroma byte `json:"chroma"`
}

func (p *P0x8801) Protocol() consts.JT808CommandType {
	return consts.P8801CameraShootImmediateCommand
}

func (p *P0x8801) ReplyProtocol() consts.JT808CommandType {
	return consts.T0805CameraShootImmediately
}

func (p *P0x8801) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) != 12 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.ChannelID = body[0]
	p.ShootCommand = binary.BigEndian.Uint16(body[1:3])
	p.PhotoIntervalOrVideoTime = binary.BigEndian.Uint16(body[3:5])
	p.SaveFlag = body[5]
	p.Resolution = body[6]
	p.VideoQuality = body[7]
	p.Intensity = body[8]
	p.Contrast = body[9]
	p.Saturation = body[10]
	p.Chroma = body[11]
	return nil
}

func (p *P0x8801) Encode() []byte {
	data := make([]byte, 12)
	data[0] = p.ChannelID
	binary.BigEndian.PutUint16(data[1:3], p.ShootCommand)
	binary.BigEndian.PutUint16(data[3:5], p.PhotoIntervalOrVideoTime)
	data[5] = p.SaveFlag
	data[6] = p.Resolution
	data[7] = p.VideoQuality
	data[8] = p.Intensity
	data[9] = p.Contrast
	data[10] = p.Saturation
	data[11] = p.Chroma
	return data
}

func (p *P0x8801) HasReply() bool {
	return false
}

func (p *P0x8801) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%02x] 通道ID:[%d]", p.ChannelID, p.ChannelID),
		fmt.Sprintf("\t[%04x] 拍摄命令:[%d] 0-表示停止拍摄 0xFFFF-表示录像 其它表示拍照张数", p.ShootCommand, p.ShootCommand),
		fmt.Sprintf("\t[%04x] 拍照间隔/录像时间:[%d] 单位秒0表示按最小间隔拍照或一直录像", p.PhotoIntervalOrVideoTime, p.PhotoIntervalOrVideoTime),
		fmt.Sprintf("\t[%02x] 保存标志:[%d]  1-保存 0-实时上传", p.SaveFlag, p.SaveFlag),
		fmt.Sprintf("\t[%02x] 分辨率:[%d] 0x01-320*240 0x02-640*480 0x03-800*600 0x04-1024*768", p.Resolution, p.Resolution),
		fmt.Sprintf("\t[%02x] 图像/视频质量:[%d] 1-10 1-代表质量损失最小 10-表示压缩比最大", p.VideoQuality, p.VideoQuality),
		fmt.Sprintf("\t[%02x] 亮度:[%d] 0-255", p.Intensity, p.Intensity),
		fmt.Sprintf("\t[%02x] 对比度:[%d] 0-127", p.Contrast, p.Contrast),
		fmt.Sprintf("\t[%02x] 饱和度:[%d] 0-127", p.Saturation, p.Saturation),
		fmt.Sprintf("\t[%02x] 色度:[%d] 0-255", p.Chroma, p.Chroma),
		"}",
	}, "\n")
}
