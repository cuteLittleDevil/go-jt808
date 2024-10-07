package model

import (
	"encoding/hex"
	"errors"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
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
		body []byte // 用于覆盖率100测试 强制替换了解析正确的body
	}
	tests := []struct {
		name   string
		fields Handler
		args   args
	}{
		{
			name: "T0X0001 终端-通用应答",
			args: args{
				msg:     "7e000100050123456789017fff007b01c803bd7e",
				Handler: &T0x0001{},
				body:    []byte{0, 123, 1, 200},
			},
			fields: &T0x0001{
				SerialNumber: 123,
				ID:           456,
				Result:       3,
			},
		},
		{
			name: "P0X8001 平台-通用应答",
			args: args{
				msg:     "7e8001000501234567890100007fff0002008e7e",
				Handler: &P0x8001{},
				body:    []byte{0, 0, 0, 0},
			},
			fields: &P0x8001{
				RespondSerialNumber: 32767,
				RespondID:           2,
				Result:              0,
			},
		},
		{
			name: "P0X8100 终端-注册消息应答",
			args: args{
				msg:     "7e8100000e01234567890100000000003132333435363738393031377e",
				Handler: &P0x8100{},
				body:    []byte{0, 0},
			},
			fields: &P0x8100{
				RespondSerialNumber: 0,
				Result:              0,
				AuthCode:            "12345678901",
			},
		},
		{
			name: "T0X0002 终端-心跳",
			args: args{
				msg:     "7e0002000001234567890100008a7e",
				Handler: &T0x0002{},
			},
			fields: &T0x0002{},
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
			if tt.args.body != nil {
				jtMsg.Body = tt.args.body
				if err := tt.args.Parse(jtMsg); err != nil {
					if errors.Is(err, protocol.ErrBodyLengthInconsistency) {
						return
					}
					t.Errorf("Parse() error = %v", err)
					return
				}
			}
		})
	}
}
