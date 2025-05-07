package utils

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
)

func TestBcd2Dec(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "2013版本",
			args: "012345678901",
			want: "12345678901",
		},
		{
			name: "2019版本",
			args: "00000000017299841738",
			want: "17299841738",
		},
		{
			name: "不需要补0的",
			args: "12345678",
			want: "12345678",
		},
		{
			name: "奇数情况",
			args: "123456789",
			want: "12345678",
		},
		{
			name: "全是0",
			args: "00000000",
			want: "00000000",
		},
		{
			name: "字母和数字组合",
			args: "abcdef1234567890ABCDEF",
			want: "abcdef1234567890abcdef",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg, _ := hex.DecodeString(tt.args)
			if got := Bcd2Dec(arg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bcd2Dec() got = %+v \n want %v", got, tt.want)
			}
		})
	}
}

func TestString2Bcd(t *testing.T) {
	tests := []struct {
		name string
		sim  string
		size int
		want []byte
	}{
		{
			name: "jt1078-2016版",
			sim:  "012345678901",
			size: 12,
			want: []byte{1, 35, 69, 103, 137, 1},
		},
		{
			name: "jt1078-2016版 sim卡号位数不够",
			sim:  "12345678901",
			size: 12,
			want: []byte{1, 35, 69, 103, 137, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String2Bcd(tt.sim, tt.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String2Bcd() got = %+v \n want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkBcd2Dec(b *testing.B) {
	bcd2013, _ := hex.DecodeString("012345678901")
	bcd2019, _ := hex.DecodeString("00000000017299841738")

	for i := 0; i < b.N; i++ {
		Bcd2Dec(bcd2013)
		Bcd2Dec(bcd2019)
	}
}

func TestCreateVerifyCode(t *testing.T) {
	tests := []struct {
		name string
		args string
		want byte
	}{
		{
			name: "2013版本",
			args: "000100050123456789017fff007b01c803",
			want: 0xbd,
		},
		{
			name: "2019版本",
			args: "000140050100000000017299841738ffff007b01c803",
			want: 0xb5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg, _ := hex.DecodeString(tt.args)
			if got := CreateVerifyCode(arg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unescape2() = %x\n want %x", got, tt.want)
			}
		})
	}
}

func TestTime2BCD(t *testing.T) {
	time := "202410010000"
	bcd := Time2BCD(time)
	want := []byte{32, 36, 16, 1, 0, 0}
	if !reflect.DeepEqual(bcd, want) {
		t.Errorf("Time2BCD() got = %x\n want %x", bcd, want)
	}
}

func BenchmarkTime2BCD(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Time2BCD("20241001000000")
	}
}

func TestBCD2Time(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "6位的时间",
			args: "241001000000",
			want: "2024-10-01 00:00:00",
		},
		{
			name: "默认时间格式的",
			args: "2024-10-01 00:00:00",
			want: "2024-10-01 00:00:00",
		},
		{
			name: "7位的时间",
			args: "20241001000000",
			want: "20241001000000",
		},
		{
			name: "非正常的时间格式",
			args: "2024100100000",
			want: "02024100100000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bcd := Time2BCD(tt.args)
			if got := BCD2Time(bcd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TestBCD2Time() got = %s\n want %s", got, tt.want)
			}
		})
	}
}

func TestString2FillingBytes(t *testing.T) {
	type args struct {
		text string
		size int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "需要补0的",
			args: args{
				text: "1234",
				size: 5,
			},
			want: []byte{'1', '2', '3', '4', 0},
		},
		{
			name: "去掉多余的",
			args: args{
				text: "12345",
				size: 4,
			},
			want: []byte{'1', '2', '3', '4'},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String2FillingBytes(tt.args.text, tt.args.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String2FillingBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodingConversion(t *testing.T) {
	type args struct {
		data               []byte
		encodingConversion func([]byte) []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "gbk -> utf8",
			args: args{
				data:               []byte{178, 226, 65, 49, 50, 51, 52},
				encodingConversion: GBK2UTF8,
			},
			want: "测A1234",
		},
		{
			name: "utf8 -> gbk",
			args: args{
				data:               []byte("测A1234"),
				encodingConversion: UTF82GBK,
			},
			want: string([]byte{178, 226, 65, 49, 50, 51, 52}),
		},
		{
			name: "utf8 -> gbk 错误的情况",
			args: args{
				data:               []byte{178, 226, 65, 49, 50, 51, 52},
				encodingConversion: UTF82GBK,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.args.encodingConversion(tt.args.data))
			if got := tt.args.encodingConversion(tt.args.data); !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("EncodingConversion() = %s, want %s", string(got), tt.want)
			}
		})
	}
}
