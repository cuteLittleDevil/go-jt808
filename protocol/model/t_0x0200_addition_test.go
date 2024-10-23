package model

import (
	"encoding/hex"
	"errors"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"io"
	"os"
	"testing"
)

func TestLocationAddition(t *testing.T) {
	type args struct {
		msg        string
		customFunc func(id uint8, content []byte) (AdditionContent, bool)
	}
	type want struct {
		path string
		err  error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "包含全部已知附加信息的",
			args: args{
				msg:        "7e020000800123456789017fff000004000000080006eeb6ad02633df701380003006320070719235901040000000b02020016030200210402002c051e3737370000000000000000000000000000000000000000000000000000001105420000004212064d0000004d4d1307000000580058582504000000632a02000a2b040000001430011e31012806020001927e",
				customFunc: nil,
			},
			want: want{
				path: "./testdata/0x0200_addition_1.txt",
				err:  nil,
			},
		},
		{
			name: "增加0x33自定义未知附加信息",
			args: args{
				msg: "7E0200007B0123456789017FFF000004000000080006EEB6AD02633DF701380003006320070719235901040000000B02020016030200210402002C051E37373700000000000000000000000000000000000000000000000000000011010012064D0000004D4D1307000000580058582504000000632A02000A2B040000001430011E3101283301207A7E",
				customFunc: func(id uint8, content []byte) (AdditionContent, bool) {
					if id == 0x33 {
						return AdditionContent{Data: content}, true
					}
					return AdditionContent{}, false
				},
			},
			want: want{
				path: "./testdata/0x0200_addition_2.txt",
				err:  nil,
			},
		},
		{
			name: "错误的数据 和协议规定的长度不符合",
			args: args{
				msg:        "7E0200007A0123456789017FFF000004000000080006EEB6AD02633DF701380003006320070719235901040000000B02020016030200210402002C051E373737000000000000000000000000000000000000000000000000000000110012064D0000004D4D1307000000580058582504000000632A02000A2B040000001430011E3101283301207A7E",
				customFunc: nil,
			},
			want: want{
				path: "",
				err:  protocol.ErrBodyLengthInconsistency,
			},
		},
		{
			name: "错误的数据 body达不到数据预期长度",
			args: args{
				msg:        "7E0200007F0123456789017FFF000004000000080006EEB6AD02633DF701380003006320070719235901040000000B02020016030200210402002C051E3737370000000000000000000000000000000000000000000000000000001105420000004212064D0000004D4D1307000000580058582504000000632A02000A2B040000001430011E310128330301597E",
				customFunc: nil,
			},
			want: want{
				path: "",
				err:  protocol.ErrBodyLengthInconsistency,
			},
		},
		{
			name: "错误的数据 缺失数据 如只有id 没有剩下的长度数据",
			args: args{
				msg:        "7E0200007D010123456789017FFF000004000000080006EEB6AD02633DF701380003006320070719235901040000000B02020016030200210402002C051E3737370000000000000000000000000000000000000000000000000000001105420000004212064D0000004D4D13070000005800585825040000FFFF2A02000B2B040000001430011E310128333B7E",
				customFunc: nil,
			},
			want: want{
				path: "",
				err:  protocol.ErrBodyLengthInconsistency,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.args.msg
			data, _ := hex.DecodeString(msg)
			jtMsg := jt808.NewJTMessage()
			if err := jtMsg.Decode(data); err != nil {
				t.Errorf("LocationAddition = %v", err)
				return
			}
			var t0x0200 T0x0200
			if tt.args.customFunc != nil {
				t0x0200.CustomAdditionContentFunc = tt.args.customFunc
			}
			if err := t0x0200.Parse(jtMsg); err != nil {
				if !errors.Is(err, tt.want.err) {
					t.Errorf("LocationAddition = %v, want %v", err, tt.want.err)
				}
				return
			}
			got := t0x0200.T0x0200AdditionDetails.String()
			txt := tt.want.path
			f, err := os.Open(txt)
			if err != nil {
				_ = os.WriteFile(txt, []byte(got), os.ModePerm)
			}
			if wantData, _ := io.ReadAll(f); string(wantData) != got {
				_ = os.WriteFile(txt+".tmp", []byte(got), os.ModePerm)
				t.Errorf("LocationAddition =\n%s\n want %s", got, string(wantData))
				return
			}
		})
	}
}
