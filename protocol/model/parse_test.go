package model

import (
	"encoding/hex"
	"errors"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"math"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	type Handler interface {
		Parse(*jt808.JTMessage) error
		String() string
	}
	type args struct {
		msg string
		Handler
		bodyLens []int // 用于覆盖率100测试 强制替换了解析正确的body
	}
	tests := []struct {
		name   string
		fields Handler
		args   args
	}{
		{
			name: "T0x0001 终端-通用应答",
			args: args{
				msg:      "7e000100050123456789017fff007b01c803bd7e",
				Handler:  &T0x0001{},
				bodyLens: []int{4},
			},
			fields: &T0x0001{
				SerialNumber: 123,
				ID:           456,
				Result:       3,
			},
		},
		{
			name: "P0x8001 平台-通用应答",
			args: args{
				msg:      "7e8001000501234567890100007fff0002008e7e",
				Handler:  &P0x8001{},
				bodyLens: []int{4},
			},
			fields: &P0x8001{
				RespondSerialNumber: 32767,
				RespondID:           2,
				Result:              0,
			},
		},
		{
			name: "P0x8100 终端-注册消息应答",
			args: args{
				msg:      "7e8100000e01234567890100000000003132333435363738393031377e",
				Handler:  &P0x8100{},
				bodyLens: []int{2},
			},
			fields: &P0x8100{
				RespondSerialNumber: 0,
				Result:              0,
				AuthCode:            "12345678901",
			},
		},
		{
			name: "T0x0002 终端-心跳",
			args: args{
				msg:     "7e0002000001234567890100008a7e",
				Handler: &T0x0002{},
			},
			fields: &T0x0002{},
		},
		{
			name: "T0x0102 注册-鉴权 2013版本",
			args: args{
				msg:     "7e0102000b01234567890100003137323939383431373338b57e",
				Handler: &T0x0102{},
			},
			fields: &T0x0102{
				AuthCodeLen:     0,
				AuthCode:        "17299841738",
				TerminalIMEI:    "",
				SoftwareVersion: "",
				Version:         consts.JT808Protocol2013,
			},
		},
		{
			name: "T0x0102 注册-鉴权 2019版本",
			args: args{
				msg:      "7e0102402f010000000001729984173800000b3137323939383431373338313233343536373839303132333435332e372e31350000000000000000000000000000227e",
				Handler:  &T0x0102{},
				bodyLens: []int{35, 37},
			},
			fields: &T0x0102{
				AuthCodeLen:     uint8(len("17299841738")),
				AuthCode:        "17299841738",
				TerminalIMEI:    "123456789012345",
				SoftwareVersion: "3.7.15",
				Version:         consts.JT808Protocol2019,
			},
		},
		{
			name: "T0x0100 终端注册 2011版本",
			args: args{
				msg:      "7e010000200123456789010000001f007363640000007777772e3830382e3736353433323101b2e24131323334a17e",
				Handler:  &T0x0100{},
				bodyLens: []int{24},
			},
			fields: &T0x0100{
				ProvinceID:         31,
				CityID:             115,
				ManufacturerID:     "cd",
				TerminalModel:      "www.808.",
				TerminalID:         "7654321",
				PlateColor:         1,
				LicensePlateNumber: "测A1234",
				Version:            consts.JT808Protocol2011,
			},
		},
		{
			name: "T0x0100 终端注册 2013版本",
			args: args{
				msg:      "7e0100002c0123456789010000001f007363640000007777772e3830382e636f6d0000000000000000003736353433323101b2e24131323334cc7e",
				Handler:  &T0x0100{},
				bodyLens: []int{36},
			},
			fields: &T0x0100{
				ProvinceID:         31,
				CityID:             115,
				ManufacturerID:     "cd",
				TerminalModel:      "www.808.com",
				TerminalID:         "7654321",
				PlateColor:         1,
				LicensePlateNumber: "测A1234",
				Version:            consts.JT808Protocol2013,
			},
		},
		{
			name: "T0x0100 终端注册 2019版本",
			args: args{
				msg:      "7e0100405301000000000172998417380000001f007363640000000000000000007777772e3830382e636f6d0000000000000000000000000000000000000037363534333231000000000000000000000000000000000000000000000001b2e241313233343b7e",
				Handler:  &T0x0100{},
				bodyLens: []int{75},
			},
			fields: &T0x0100{
				ProvinceID:         31,
				CityID:             115,
				ManufacturerID:     "cd",
				TerminalModel:      "www.808.com",
				TerminalID:         "7654321",
				PlateColor:         1,
				LicensePlateNumber: "测A1234",
				Version:            consts.JT808Protocol2019,
			},
		},
		{
			name: "T0x0200 位置上报",
			args: args{
				msg:      "7e0200001c0123456789010000000004000000080007203b7d0202633df70138000300632410012359591c7e",
				Handler:  &T0x0200{},
				bodyLens: []int{27},
			},
			fields: &T0x0200{
				T0x0200LocationItem: T0x0200LocationItem{
					AlarmSign:  1024,
					StatusSign: 2048,
					Latitude:   119552894,
					Longitude:  40058359,
					Altitude:   312,
					Speed:      3,
					Direction:  99,
					DateTime:   "2024-10-01 23:59:59",
				},
			},
		},
		{
			name: "T0x0704 位置批量上传",
			args: args{
				msg:      "7e0704003f0123456789010000000200001c000004000000080007203b7d0202633df7013800030063241001235959001c000004000000080007203b7d0202633df7013800030063241001235959b67e",
				Handler:  &T0x0704{},
				bodyLens: []int{30, 60, 68},
			},
			fields: &T0x0704{
				Num:          2,
				LocationType: 0,
				Items: []T0x0704LocationItem{
					{
						Len: 28,
						T0x0200LocationItem: T0x0200LocationItem{
							AlarmSign:  1024,
							StatusSign: 2048,
							Latitude:   119552894,
							Longitude:  40058359,
							Altitude:   312,
							Speed:      3,
							Direction:  99,
							DateTime:   "2024-10-01 23:59:59",
						},
					},
					{
						Len: 28,
						T0x0200LocationItem: T0x0200LocationItem{
							AlarmSign:  1024,
							StatusSign: 2048,
							Latitude:   119552894,
							Longitude:  40058359,
							Altitude:   312,
							Speed:      3,
							Direction:  99,
							DateTime:   "2024-10-01 23:59:59",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := hex.DecodeString(tt.args.msg)
			jtMsg := jt808.NewJTMessage()
			if err := jtMsg.Decode(data); err != nil {
				t.Errorf("Decode() error = %v", err)
				return
			}
			if err := tt.args.Parse(jtMsg); err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			//fmt.Println(tt.args.Handler.String())
			if tt.args.Handler.String() != tt.fields.String() {
				t.Errorf("Parse() want: \n%v\nactual:\n%v", tt.args, tt.fields)
				return
			}
			body := jtMsg.Body
			for _, bodyLen := range tt.args.bodyLens {
				jtMsg.Body = body[:bodyLen]
				if err := tt.args.Parse(jtMsg); err != nil {
					if !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
						t.Errorf("Parse() error = %v", err)
						return
					}
				}
			}
		})
	}
}

// 为了覆盖率100%增加的测试 ------------------------------------
func TestT0x0704Parse(t *testing.T) {
	msg := "7e0704003f0123456789010000000200001c000004000000080007203b7d0202633df7013800030063241001235959001c000004000000080007203b7d0202633df7013800030063241001235959b67e"
	data, _ := hex.DecodeString(msg)
	jtMsg := jt808.NewJTMessage()
	_ = jtMsg.Decode(data)
	handler := &T0x0704{}
	// 强制错误情况
	jtMsg.Body = jtMsg.Body[:63]
	jtMsg.Body[4] = 0x00
	if err := handler.Parse(jtMsg); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Errorf("T0x0704 Parse() err[%v]", err)
		return
	}
}

func TestT0x0200LocationItemString(t *testing.T) {
	var t0x0200Item T0x0200LocationItem
	t0x0200Item.AlarmSignDetails.parse(math.MaxUint32)
	alarmSignData, _ := os.ReadFile("./testdata/0x0200_alarm_sign.txt")
	if string(alarmSignData) != t0x0200Item.AlarmSignDetails.String() {
		t.Errorf("want[%s] actual[%s]", string(alarmSignData), t0x0200Item.AlarmSignDetails.String())
		return
	}

	infos := map[uint32]string{
		1<<23 - 1:             "./testdata/0x0200_status_sign_03.txt",
		1<<23 - 1 - 256 - 512: "./testdata/0x0200_status_sign_00.txt",
		1<<23 - 1 - 256:       "./testdata/0x0200_status_sign_01.txt",
		1<<23 - 1 - 512:       "./testdata/0x0200_status_sign_02.txt",
	}
	for statusSign, signPath := range infos {
		var tmp T0x0200LocationItem
		tmp.StatusSignDetails.parse(statusSign)
		statusSignData, _ := os.ReadFile(signPath)
		if string(statusSignData) != tmp.StatusSignDetails.String() {
			t.Errorf("path[%s]\n%s", signPath, tmp.StatusSignDetails.String())
			return
		}
	}
}
