package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"strings"
)

type T0x0200LocationItem struct {
	// AlarmSign 报警标志 JT808.Protocol.Enums.JT808Alarm
	AlarmSign uint32 `json:"alarmSign"`
	// StatusSign 状态标志 JT808.Protocol.Enums.JT808Status
	StatusSign uint32 `json:"statusSign"`
	// Latitude 纬度 以度为单位的纬度值乘以 10 的 6 次方，精确到百万分之一度
	Latitude uint32 `json:"latitude"`
	// Longitude 经度 以度为单位的经度值乘以 10 的 6 次方，精确到百万分之一度
	Longitude uint32 `json:"longitude"`
	// Altitude 海拔高度 单位为米
	Altitude uint16 `json:"altitude"`
	// Speed 速度 1/10km/h
	Speed uint16 `json:"speed"`
	// Direction 方向 0-359，正北为 0，顺时针
	Direction uint16 `json:"direction"`
	// DateTime 时间 YY-MM-DD-hh-mm-ss（GMT+8 时间，本标准中之后涉及的时间均采用此时区）
	DateTime string `json:"dateTime"`
}

func (tl *T0x0200LocationItem) parse(body []byte) error {
	if len(body) < 28 {
		return protocol.ErrBodyLengthInconsistency
	}
	tl.AlarmSign = binary.BigEndian.Uint32(body[:4])
	tl.StatusSign = binary.BigEndian.Uint32(body[4:8])
	tl.Latitude = binary.BigEndian.Uint32(body[8:12])
	tl.Longitude = binary.BigEndian.Uint32(body[12:16])
	tl.Altitude = binary.BigEndian.Uint16(body[16:18])
	tl.Speed = binary.BigEndian.Uint16(body[18:20])
	tl.Direction = binary.BigEndian.Uint16(body[20:22])
	tl.DateTime = utils.BCD2Time(body[22:28])
	return nil
}

func (tl *T0x0200LocationItem) encode() []byte {
	data := make([]byte, 22, 30)
	binary.BigEndian.PutUint32(data[:4], tl.AlarmSign)
	binary.BigEndian.PutUint32(data[4:8], tl.StatusSign)
	binary.BigEndian.PutUint32(data[8:12], tl.Latitude)
	binary.BigEndian.PutUint32(data[12:16], tl.Longitude)
	binary.BigEndian.PutUint16(data[16:18], tl.Altitude)
	binary.BigEndian.PutUint16(data[18:20], tl.Speed)
	binary.BigEndian.PutUint16(data[20:22], tl.Direction)
	bcdTime := strings.ReplaceAll(tl.DateTime, "-", "")
	bcdTime = strings.ReplaceAll(bcdTime, ":", "")
	bcdTime = strings.ReplaceAll(bcdTime, " ", "")
	if len(bcdTime) == 14 {
		bcdTime = bcdTime[2:]
	}
	data = append(data, utils.Time2BCD(bcdTime)...)
	return data
}

func (tl *T0x0200LocationItem) String() string {
	body := tl.encode()
	return strings.Join([]string{
		fmt.Sprintf("\t[%08x] 报警标志:[%d]", tl.AlarmSign, tl.AlarmSign),
		fmt.Sprintf("\t[%08x] 状态标志:[%d]", tl.StatusSign, tl.StatusSign),
		fmt.Sprintf("\t[%08x] 纬度:[%d]", tl.Latitude, tl.Latitude),
		fmt.Sprintf("\t[%08x] 经度:[%d]", tl.Longitude, tl.Longitude),
		fmt.Sprintf("\t[%04x] 海拔高度:[%d]", tl.Altitude, tl.Altitude),
		fmt.Sprintf("\t[%04x] 速度:[%d]", tl.Speed, tl.Speed),
		fmt.Sprintf("\t[%04x] 方向:[%d]", tl.Direction, tl.Direction),
		fmt.Sprintf("\t[%x] 时间:[%s]", body[22:28], tl.DateTime),
	}, "\n")
}
