package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x0801 struct {
	BaseHandle
	// MultimediaID  多媒体数据ID 值大于0
	MultimediaID uint32 `json:"multimediaIDNumber"`
	// MultimediaType 多媒体数据类型 0-图像 1-音频 2-视频
	MultimediaType byte `json:"multimediaType"`
	// MultimediaFormatEncode 多媒体格式编码 0-jpeg 1-tlf 2-mp3 4-wav 4-wmv 其他保留
	MultimediaFormatEncode byte `json:"multimediaFormatEncode"`
	// EventItemEncode 事件项编码 0-平台下发指令 1-定时动作 2-抢劫报警触发 3-碰撞侧翻报警触发 4-门开拍照 5-门关拍照
	EventItemEncode byte `json:"eventItemEncode"`
	// ChannelID 通道ID
	ChannelID           byte `json:"channelID"`
	T0x0200LocationItem `json:"t0X0200LocationItem"`
	// MultimediaPackage 多媒体包
	MultimediaPackage []byte `json:"multimediaPackage"`
}

func (t *T0x0801) Protocol() consts.JT808CommandType {
	return consts.T0801MultimediaDataUpload
}

func (t *T0x0801) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 36 {
		return protocol.ErrBodyLengthInconsistency
	}
	t.MultimediaID = binary.BigEndian.Uint32(body[0:4])
	t.MultimediaType = body[4]
	t.MultimediaFormatEncode = body[5]
	t.EventItemEncode = body[6]
	t.ChannelID = body[7]
	_ = t.T0x0200LocationItem.parse(body[8:36])
	t.MultimediaPackage = body[36:]
	return nil
}

func (t *T0x0801) Encode() []byte {
	data := make([]byte, 8, 100)
	binary.BigEndian.PutUint32(data[0:4], t.MultimediaID)
	data[4] = t.MultimediaType
	data[5] = t.MultimediaFormatEncode
	data[6] = t.EventItemEncode
	data[7] = t.ChannelID
	data = append(data, t.T0x0200LocationItem.encode()...)
	data = append(data, t.MultimediaPackage...)
	return data
}

func (t *T0x0801) ReplyBody(_ *jt808.JTMessage) ([]byte, error) {
	return nil, nil
}

func (t *T0x0801) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", t.Protocol(), t.Encode()[:26]),
		fmt.Sprintf("\t[%08x] 多媒体数据ID:[%d]", t.MultimediaID, t.MultimediaID),
		fmt.Sprintf("\t[%02x] 多媒体数据类型:[%d] 0-图像 1-音频 2-视频", t.MultimediaType, t.MultimediaType),
		fmt.Sprintf("\t[%02x] 多媒体格式编码:[%d] 0-jpeg 1-tlf 2-mp3 4-wav 4-wmv", t.MultimediaFormatEncode, t.MultimediaFormatEncode),
		fmt.Sprintf("\t[%02x] 事件项编码:[%d] 0-平台下发指令 1-定时动作 2-抢劫报警触发 3-碰撞侧翻报警触发", t.EventItemEncode, t.EventItemEncode),
		fmt.Sprintf("\t[%02x] 通道ID:[%d]", t.ChannelID, t.ChannelID),
		t.T0x0200LocationItem.String(),
		fmt.Sprintf("\t多媒体包大小:[%d]", t.MultimediaPackage),
		"}",
	}, "\n")
}
