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

type T0x0100 struct {
	BaseHandle
	// ProvinceID 省域 ID
	// 标示终端安装车辆所在的省域，0 保留，由平台取默
	// 认值。省域 ID 采用 GB/T 2260 中规定的行政区划代
	// 码六位中前两位
	ProvinceID uint16 `json:"provinceId"`
	// CityID 市县域 ID
	// 标示终端安装车辆所在的市域和县域，0 保留，由平
	// 台取默认值。市县域 ID 采用 GB/T 2260 中规定的行
	// 政区划代码六位中后四位。
	CityID uint16 `json:"cityId"`
	// ManufacturerID 制造商 ID
	// 2013版本 5 个字节，终端制造商编码
	// 2019版本 11 个字节，终端制造商编码
	ManufacturerID string `json:"manufacturerId"`
	// TerminalModel 终端型号
	// 2011版本   8个字节  ，此终端型号由制造商自行定义，位数不足时，后补“0X00”
	// 2013版本   20 个字节，此终端型号由制造商自行定义，位数不足时，后补“0X00”。
	// 2019版本   30 个字节，此终端型号由制造商自行定义，位数不足时，后补“0X00”。
	TerminalModel string `json:"terminalModel"`
	// TerminalID 终端 ID
	// 2013版本  7个字节，由大写字母和数字组成，此终端 ID 由制造商自行定义，位数不足时，后补“0X00”。
	// 2019版本  30个字节，由大写字母和数字组成，此终端 ID 由制造商自行定义，位数不足时，后补“0X00”。
	TerminalID string `json:"terminalID"`

	// PlateColor 车牌颜色
	// 2013版本 7个字节 按照JT415-2006定义，5.4.12节，0=未上牌，1=蓝，2=黄，3=黑，4=白，9=其他
	// 2019版本 30个字节 按照JT697.7-2014定义，5.6节，0=为上牌，1=蓝，2=黄，3=黑，4=白，5=绿，9=其他
	PlateColor byte `json:"plateColor"`
	// LicensePlateNumber 车辆标识
	// 车牌颜色为 0 时，表示车辆 VIN；
	// 否则，表示公安交通管理部门颁发的机动车号牌。
	LicensePlateNumber string `json:"licensePlateNumber"`
	// Version 版本 1-2011 2-2013 3-2019
	Version consts.ProtocolVersionType `json:"version"`
}

func (t *T0x0100) Protocol() consts.JT808CommandType {
	return consts.T0100Register
}

func (t *T0x0100) ReplyProtocol() consts.JT808CommandType {
	return consts.P8100RegisterRespond
}

func (t *T0x0100) Parse(jtMsg *jt808.JTMessage) error {
	var ( // 默认按照2011版本
		mLen   = 5
		tLen   = 8
		tIDLen = 7
	)
	body := jtMsg.Body
	switch jtMsg.Header.ProtocolVersion {
	case consts.JT808Protocol2019:
		mLen = 11
		tLen = 30
		tIDLen = 30
		t.Version = consts.JT808Protocol2019
	case consts.JT808Protocol2013, consts.JT808Protocol2011:
		// 根据body长度判断是不是2011版本的 2013版本的车牌颜色是36
		if len(body) > 36 {
			t.Version = consts.JT808Protocol2013
			mLen = 5
			tLen = 20
			tIDLen = 7
		} else {
			t.Version = consts.JT808Protocol2011
		}
	}

	if t.Version == consts.JT808Protocol2011 && len(body) < 25 {
		return protocol.ErrBodyLengthInconsistency
	} else if t.Version == consts.JT808Protocol2019 && len(body) < 76 {
		return protocol.ErrBodyLengthInconsistency
	}

	t.ProvinceID = binary.BigEndian.Uint16(body[:2])
	t.CityID = binary.BigEndian.Uint16(body[2:4])

	cutset := "\x00"
	t.ManufacturerID = string(bytes.TrimRight(body[4:4+mLen], cutset))
	t.TerminalModel = string(bytes.TrimRight(body[4+mLen:4+mLen+tLen], cutset))
	t.TerminalID = string(bytes.TrimRight(body[4+mLen+tLen:4+mLen+tLen+tIDLen], cutset))
	t.PlateColor = body[4+mLen+tLen+tIDLen]
	utf8Data := utils.GBK2UTF8(body[4+mLen+tLen+tIDLen+1:])
	t.LicensePlateNumber = string(utf8Data)
	return nil
}

func (t *T0x0100) ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error) {
	// 不限制 默认鉴权码用手机号
	p8100 := &P0x8100{
		RespondSerialNumber: jtMsg.Header.SerialNumber,
		Result:              0,
		AuthCode:            jtMsg.Header.TerminalPhoneNo,
	}
	return p8100.Encode(), nil
}

func (t *T0x0100) Encode() []byte {
	mLen, tLen, tIDLen := t.protocolDiff()
	data := make([]byte, 4, 4+mLen+tLen+tIDLen)
	binary.BigEndian.PutUint16(data[:2], t.ProvinceID)
	binary.BigEndian.PutUint16(data[2:4], t.CityID)
	data = append(data, utils.String2FillingBytes(t.ManufacturerID, mLen)...)
	data = append(data, utils.String2FillingBytes(t.TerminalModel, tLen)...)
	data = append(data, utils.String2FillingBytes(t.TerminalID, tIDLen)...)
	data = append(data, t.PlateColor)
	gbkData := utils.UTF82GBK([]byte(t.LicensePlateNumber))
	data = append(data, gbkData...)
	return data
}

func (t *T0x0100) String() string {
	str := "数据体对象:{\n"
	data := t.Encode()
	str += fmt.Sprintf("\t%s:[%x]", consts.T0100Register, data)
	mLen, tLen, tIDLen := t.protocolDiff()
	f := func(arg string, size int, remark string) string {
		format := "\t[%0" + fmt.Sprintf("%d", mLen) + "x] " + remark + "(%d):[%s]"
		return fmt.Sprintf(format, utils.String2FillingBytes(arg, size), size, arg)
	}
	return strings.Join([]string{
		str,
		fmt.Sprintf("\t[%04x] 省域ID:[%d]", t.ProvinceID, t.ProvinceID),
		fmt.Sprintf("\t[%04x] 市县域ID:[%d]", t.CityID, t.CityID),
		f(t.ManufacturerID, mLen, "制造商ID"),
		f(t.TerminalModel, tLen, "终端型号"),
		f(t.TerminalID, tIDLen, "终端ID"),
		fmt.Sprintf("\t[%02x] 车牌颜色:[%d]", t.PlateColor, t.PlateColor),
		fmt.Sprintf("\t[%x] 车牌号:[%s]", data[4+mLen+tLen+tIDLen+1:], t.LicensePlateNumber),
		"}",
	}, "\n")
}

func (t *T0x0100) protocolDiff() (int, int, int) {
	var ( // 默认按照2011版本
		mLen   = 5
		tLen   = 8
		tIDLen = 7
	)
	if t.Version == consts.JT808Protocol2013 {
		mLen, tLen, tIDLen = 5, 20, 7
	} else if t.Version == consts.JT808Protocol2019 {
		mLen, tLen, tIDLen = 11, 30, 30
	}
	return mLen, tLen, tIDLen
}
