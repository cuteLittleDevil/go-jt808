package model

import (
	"encoding/hex"
	"errors"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
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
