package jt808

import (
	"encoding/hex"
	"errors"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"reflect"
	"testing"
)

func Test_unescape(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    string
		wantErr error
	}{
		{
			name: "不需要反转义的",
			args: "7e000140050100000000017299841738ffff007b01c803b57e",
			want: "000140050100000000017299841738ffff007b01c803b5",
		},
		{
			name: "需要反转义的",
			args: "7e0200401c01000000000172998417380000000004000000080007203b7d0202633df7013800030063200707192359c17e",
			want: "0200401c01000000000172998417380000000004000000080007203b7e02633df7013800030063200707192359c1",
		},
		{
			name:    "错误的数据 内容带7e并且后面不是01或者02",
			args:    "7e02007d037e",
			wantErr: protocol.ErrUnqualifiedData,
		},
		{
			name:    "错误的数据",
			args:    "7e02",
			wantErr: protocol.ErrUnqualifiedData,
		},
		{
			name: "全部需要转义的",
			args: "7e02007d0100107d027e",
			want: "02007d00107e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg, _ := hex.DecodeString(tt.args)
			got, err := unescape(arg)
			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("unescape() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			want, _ := hex.DecodeString(tt.want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("unescape2() = %x\n want %x", got, want)
			}
		})
	}
}

func Benchmark_unescape(b *testing.B) {
	// 假设10个数据里面 7个不需要转义 2个转义1次 1个转义2次
	escapeZero := "7e000140050100000000017299841738ffff007b01c803b57e"
	escapeOne := "7e080040080100000000017299841738ffff0000007b000007017d017e"
	escapeTwo := "7e08007d0240080100000000017299841738ffff0000007b000007017d017e"
	datas := make([][]byte, 0, 10)
	for i := 0; i < 7; i++ {
		arg, _ := hex.DecodeString(escapeZero)
		datas = append(datas, arg)
	}
	for i := 0; i < 2; i++ {
		arg, _ := hex.DecodeString(escapeOne)
		datas = append(datas, arg)
	}
	for i := 0; i < 1; i++ {
		arg, _ := hex.DecodeString(escapeTwo)
		datas = append(datas, arg)
	}
	for i := 0; i < b.N; i++ {
		for _, data := range datas {
			_, _ = unescape(data)
		}
	}
}

func Test_escape(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "不需要转义的",
			args: "000140050100000000017299841738ffff007b01c803b5",
			want: "7e000140050100000000017299841738ffff007b01c803b57e",
		},
		{
			name: "需要转义的",
			args: "0800407e08010000000001727d99841738ffff0000007b000007017d",
			want: "7e0800407d0208010000000001727d0199841738ffff0000007b000007017d017e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg, _ := hex.DecodeString(tt.args)
			want, _ := hex.DecodeString(tt.want)
			if got := escape(arg); !reflect.DeepEqual(got, want) {
				t.Errorf("unescape2() = %x\n want %x", got, want)
			}
		})
	}
}

func Benchmark_escape(b *testing.B) {
	// 假设10个数据里面 7个不需要转义 2个转义1次 1个转义2次
	escapeZero := "000140050100000000017299841738ffff007b01c803b5"
	escapeOne := "080040080100000000017299841738ffff0000007b000007017d01"
	escapeTwo := "08007d0240080100000000017299841738ffff0000007b000007017d01"
	datas := make([][]byte, 0, 10)
	for i := 0; i < 7; i++ {
		arg, _ := hex.DecodeString(escapeZero)
		datas = append(datas, arg)
	}
	for i := 0; i < 2; i++ {
		arg, _ := hex.DecodeString(escapeOne)
		datas = append(datas, arg)
	}
	for i := 0; i < 1; i++ {
		arg, _ := hex.DecodeString(escapeTwo)
		datas = append(datas, arg)
	}
	for i := 0; i < b.N; i++ {
		for _, data := range datas {
			_ = escape(data)
		}
	}
}
