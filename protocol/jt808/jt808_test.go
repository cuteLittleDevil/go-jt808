package jt808

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"os"
	"reflect"
	"testing"
)

func TestJTMessageDecode(t *testing.T) {
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
			want: "7e0002000001234567890100008a7e",
		},
		{
			name: "2019版本",
			args: "7e0002400001000000000172998417380000027e",
			want: "7e0002400001000000000172998417380000027e",
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

func TestHeader_EncodePackets(t *testing.T) {
	type want struct {
		decodeCount int
		encodeCount int
	}
	tests := []struct {
		name    string
		bodyLen int
		want    want
	}{
		{
			name:    "1. 单包_小于1000字节",
			bodyLen: 256,
			want: want{
				decodeCount: 1,
				encodeCount: 1,
			},
		},
		{
			name:    "2. 边界_刚好1000字节",
			bodyLen: 1000,
			want: want{
				decodeCount: 1,
				encodeCount: 1,
			},
		},
		{
			name:    "3. 两包_1001字节",
			bodyLen: 1001,
			want: want{
				decodeCount: 2,
				encodeCount: 2,
			},
		},
		{
			name:    "4. 多包_2500字节",
			bodyLen: 2500,
			want: want{
				decodeCount: 3,
				encodeCount: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jtMsg := NewJTMessage()
			data, _ := hex.DecodeString("7e0002400001000000000172998417380000027e")
			_ = jtMsg.Decode(data)

			body := make([]byte, tt.bodyLen)
			for i := range body {
				body[i] = byte(i % 255)
			}

			packets := jtMsg.Header.EncodePackets(body)
			if len(packets) != tt.want.encodeCount {
				t.Fatalf("EncodePackets() 返回包数量 = %d, want %d", len(packets), tt.want.decodeCount)
			}

			// 验证每个包都能正常解码 + 分包信息正确
			var reconstructed []byte
			for i, pkt := range packets {
				msg := NewJTMessage()
				if err := msg.Decode(pkt); err != nil {
					t.Fatalf("第 %d 个包解码失败: %v", i+1, err)
				}

				reconstructed = append(reconstructed, msg.Body...)

				// 检查分包标识
				if tt.want.decodeCount == 1 {
					if msg.Header.Property.PacketFragmented != 0 {
						t.Errorf("单包情况下 PacketFragmented 应为 0，实际为 %d", msg.Header.Property.PacketFragmented)
					}
					if msg.Header.SubPackageSum != 0 || msg.Header.SubPackageNo != 0 {
						t.Errorf("单包情况下 SubPackageSum/No 应为 0")
					}
				} else {
					if msg.Header.Property.PacketFragmented != 1 {
						t.Errorf("多包情况下 PacketFragmented 应为 1")
					}
					if msg.Header.SubPackageSum != uint16(tt.want.decodeCount) {
						t.Errorf("第[%d]包 SubPackageSum = %d, want %d", i+1, msg.Header.SubPackageSum, tt.want.decodeCount)
					}
					expectedNo := uint16(i + 1)
					if msg.Header.SubPackageNo != expectedNo {
						t.Errorf("第[%d]包 SubPackageNo = %d, want %d", i+1, msg.Header.SubPackageNo, expectedNo)
					}
				}

				// 检查流水号偏移是否正确 (PlatformSerialNumber + num-1)
				expectedSerial := jtMsg.Header.PlatformSerialNumber + uint16(i)
				if msg.Header.SerialNumber != expectedSerial {
					t.Errorf("第[%d]包 SerialNumber = %d, want %d", i+1, msg.Header.SerialNumber, expectedSerial)
				}
			}

			// 所有包解码后 body 能还原
			if !bytes.Equal(reconstructed, body) {
				t.Errorf("分包后重新拼接的 body 与原始 body 不一致！\n got=[%x] \n want[%x]", reconstructed, body)
			}
		})
	}
}

func TestEncodeSubpackage(t *testing.T) {
	type args struct {
		path    string
		command uint16
		seq     uint16
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "0x1205 指令分包",
			args: args{
				path:    "./testdata/0x1205_src.txt",
				command: uint16(consts.T1205UploadAudioVideoResourceList),
				seq:     17148,
			},
			want: "./testdata/0x1205_dst.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := "7e0002400001000000000172998417380000027e"
			jtMsg := NewJTMessage()
			data, _ := hex.DecodeString(msg)
			_ = jtMsg.Decode(data)

			src, err := os.ReadFile(tt.args.path)
			if err != nil {
				t.Fatal(err)
			}
			srcData, _ := hex.DecodeString(string(src))
			jtMsg.Header.ReplyID = tt.args.command
			jtMsg.Header.PlatformSerialNumber = tt.args.seq
			got := jtMsg.Header.Encode(srcData)
			dst, err := os.ReadFile(tt.want)
			if err != nil {
				t.Fatal(err)
			}
			dst = bytes.ReplaceAll(dst, []byte("\n"), []byte(""))
			dstData, _ := hex.DecodeString(string(dst))
			if !reflect.DeepEqual(got, dstData) {
				t.Errorf("Encode() \ngot = %x\nwant= %x", got, dstData)
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
