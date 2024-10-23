package jt808

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"os"
	"reflect"
	"testing"
)

func TestJTMessage_Decode(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		wantErr error
	}{
		{
			name: "2013版本",
			args: "7e0100002c0123456789010000001f0073797a6800007777772e6a74743830382e636f6d0000000000003736353433323101b2e24131323334ca7e",
		},
		{
			name: "2019版本",
			args: "7e0100405301000000000172998417380000001f0073797a6800000000000000007777772e6a74743830382e636f6d0000000000000000000000000000000037363534333231000000000000000000000000000000000000000000000001b2e241313233343d7e",
		},
		{
			name: "正确的分包数据",
			args: "7E0801200500123456789002DE001A00022808000102537E",
		},
		{
			name: "RSA加密数据",
			args: "7E0801040500123456789002DE001A000221757E", // 模拟生成的 仅标志位=1为RSA
		},
		{
			name: "兼容部分错误情况",
			args: "7e0002000000000000067900007d7e",
		},
		{
			name:    "不完整的数据",
			args:    "7e010040530100",
			wantErr: protocol.ErrUnqualifiedData,
		},
		{
			name:    "错误的数据",
			args:    "7e01017e",
			wantErr: protocol.ErrHeaderLength2Short,
		},
		{
			name: "RSA加密数据",
			args: "7E0801040500123456789002DE001A000221757E", // 模拟生成的 仅标志位=1为RSA
		},
		{
			name:    "校验码错误",
			args:    "7E0801200500123456789002DE001A00022808000102547E",
			wantErr: protocol.ErrCheckCode,
		},
		{
			name:    "body数据和解析头不符合",
			args:    "7E0801200500123456789002DE001A000228080001517E",
			wantErr: protocol.ErrBodyLengthInconsistency,
		},
		{
			name:    "头部情况不足",
			args:    "7e0100002c0123454a7e",
			wantErr: protocol.ErrHeaderLength2Short,
		},
		{
			name:    "头部情况不足 分包情况",
			args:    "7E0801200500123456789002b67E",
			wantErr: protocol.ErrHeaderLength2Short,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jtMsg := NewJTMessage()
			arg, _ := hex.DecodeString(tt.args)
			if err := jtMsg.Decode(arg); err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
		})
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "2013版本",
			args: "7e0002000001234567890100008a7e",
			want: "7e000000000123456789010000887e",
		},
		{
			name: "2019版本",
			args: "7e0002400001000000000172998417380000027e",
			want: "7e0000400001000000000172998417380000007e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jtMsg := NewJTMessage()
			head, _ := hex.DecodeString(tt.args)
			_ = jtMsg.Decode(head)
			data := jtMsg.Header.Encode(nil)
			got := fmt.Sprintf("%x", data)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encode() = %s\n want %s", got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "2013版本",
			args: "7e0002000001234567890100008a7e",
			want: "./testdata/head_2013.txt",
		},
		{
			name: "2019版本",
			args: "7e0002400001000000000172998417380000027e",
			want: "./testdata/head_2019.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jtMsg := NewJTMessage()
			head, _ := hex.DecodeString(tt.args)
			_ = jtMsg.Decode(head)
			got := jtMsg.Header.String()
			txt, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("open file [%s] [%v]", tt.want, err)
				return
			}
			if !reflect.DeepEqual(got, string(txt)) {
				t.Errorf("Encode() = %s\n want %s", got, string(txt))
			}
		})
	}
}
