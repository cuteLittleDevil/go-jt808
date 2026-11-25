package jt808

import (
	"bytes"
	"encoding/binary"
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
			err := jtMsg.Decode(arg)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBodyProperty_Bit15AndEncryptMethod(t *testing.T) {
	tests := []struct {
		name        string
		raw         uint16
		wantBit15   byte
		wantVersion uint8
		wantFrag    uint8
		wantEncrypt uint8
		wantBodyLen uint16
	}{
		{
			name:        "bit15置1且可无损编回",
			raw:         0x800A,
			wantBit15:   1,
			wantBodyLen: 10,
		},
		{
			name:        "加密方式bit11置1(值=2)",
			raw:         0x0805,
			wantEncrypt: 2,
			wantBodyLen: 5,
		},
		{
			name:        "加密方式bit10-12全1(值=7)+2019+分包",
			raw:         0xFC10,
			wantBit15:   1,
			wantVersion: 1,
			wantFrag:    1,
			wantEncrypt: 7,
			wantBodyLen: 16,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b [2]byte
			binary.BigEndian.PutUint16(b[:], tt.raw)

			var p BodyProperty
			p.decode(b[:])
			if p.bit15 != tt.wantBit15 ||
				p.Version != tt.wantVersion ||
				p.PacketFragmented != tt.wantFrag ||
				p.EncryptMethod != tt.wantEncrypt ||
				p.BodyDayaLen != tt.wantBodyLen {
				t.Fatalf("decode(0x%04x) = %+v", tt.raw, p)
			}
			if got := p.encode(); got != tt.raw {
				t.Fatalf("encode() = 0x%04x, want 0x%04x", got, tt.raw)
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

func TestJTMessageDecode_ReuseSubPackageReset(t *testing.T) {
	base := NewJTMessage()
	rawBase, _ := hex.DecodeString("7e0002000001234567890100008a7e")
	if err := base.Decode(rawBase); err != nil {
		t.Fatalf("Decode base failed: %v", err)
	}
	packets := base.Header.EncodePackets(make([]byte, maxBodyLength+1))
	if len(packets) != 2 {
		t.Fatalf("EncodePackets len = %d, want 2", len(packets))
	}

	reused := NewJTMessage()
	if err := reused.Decode(packets[0]); err != nil {
		t.Fatalf("Decode subpackage failed: %v", err)
	}
	if reused.Header.SubPackageSum != 2 || reused.Header.SubPackageNo != 1 || !reused.Header.Property.isSubPackage {
		t.Fatalf("subpackage header unexpected: sum=%d no=%d isSub=%v",
			reused.Header.SubPackageSum, reused.Header.SubPackageNo, reused.Header.Property.isSubPackage)
	}

	if err := reused.Decode(rawBase); err != nil {
		t.Fatalf("Decode non-subpackage on reused JTMessage failed: %v", err)
	}
	if reused.Header.SubPackageSum != 0 || reused.Header.SubPackageNo != 0 || reused.Header.Property.isSubPackage {
		t.Fatalf("reused header stale subpackage state: sum=%d no=%d isSub=%v",
			reused.Header.SubPackageSum, reused.Header.SubPackageNo, reused.Header.Property.isSubPackage)
	}
}

func TestHeader_EncodeDoesNotMutateProperty(t *testing.T) {
	msg := NewJTMessage()
	raw, _ := hex.DecodeString("7e0002000001234567890100008a7e")
	if err := msg.Decode(raw); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	origProp := *msg.Header.Property

	packets := msg.Header.EncodePackets(make([]byte, maxBodyLength+10))
	if len(packets) != 2 {
		t.Fatalf("EncodePackets len = %d, want 2", len(packets))
	}
	if *msg.Header.Property != origProp {
		t.Fatalf("EncodePackets mutated Header.Property: got=%+v want=%+v", *msg.Header.Property, origProp)
	}

	// Property 为 nil 时也不 panic
	nilPropHeader := &Header{
		ID:                 0x0002,
		bcdTerminalPhoneNo: []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0x01},
	}
	if pkts := nilPropHeader.EncodePackets(nil); len(pkts) != 1 {
		t.Fatalf("nil Property EncodePackets len = %d, want 1", len(pkts))
	}
}
