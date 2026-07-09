package jt808

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
)

// assertExtractedFrame 校验 FrameReader 切出的帧形态：首尾必须是 0x7e.
func assertExtractedFrame(t *testing.T, frame []byte) {
	t.Helper()
	if len(frame) < 2 {
		t.Fatalf("frame too short: len=%d data=%x", len(frame), frame)
	}
	if frame[0] != FrameSign || frame[len(frame)-1] != FrameSign {
		t.Fatalf("frame must start/end with 0x7e: %x", frame)
	}
}

// assertPendingNoCompleteFrame 粘包缓冲区里不应存在 PopFrame 能取走的完整帧.
// PopFrame 要求 len>2，故 7e7e 这类仅两字节的数据会留在缓冲区，属正常情况.
func assertPendingNoCompleteFrame(t *testing.T, pending []byte) {
	t.Helper()
	if len(pending) <= 2 {
		return
	}
	if pending[0] != FrameSign {
		return
	}
	if bytes.IndexByte(pending[1:], FrameSign) >= 0 {
		t.Fatalf("pending contains complete frame but was not popped: %x", pending)
	}
}

// unpackLikeService 模拟 service 的 unpack：切帧后立即 Decode，解码失败则停止.
func unpackLikeService(data []byte) (frames [][]byte, decodeErr error, pending int) {
	r := NewFrameReader()
	if frame, ok := r.FeedSingleComplete(data); ok {
		if err := NewJTMessage().Decode(frame); err != nil {
			return nil, err, r.Pending()
		}
		return [][]byte{frame}, nil, r.Pending()
	}
	r.Append(data)
	for {
		frame, ok := r.PopFrame()
		if !ok {
			break
		}
		if err := NewJTMessage().Decode(frame); err != nil {
			return frames, err, r.Pending()
		}
		frames = append(frames, frame)
	}
	return frames, nil, r.Pending()
}

func TestTryPopFrame(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantFrame string
		wantRest  string
		wantOK    bool
	}{
		{
			name:      "完整单帧",
			data:      "7e000200000123456789017fff0a7e",
			wantFrame: "7e000200000123456789017fff0a7e",
			wantRest:  "",
			wantOK:    true,
		},
		{
			name:      "完整帧后有剩余",
			data:      "7e000200000123456789017fff0a7e7e8001",
			wantFrame: "7e000200000123456789017fff0a7e",
			wantRest:  "7e8001",
			wantOK:    true,
		},
		{
			name:      "不完整帧",
			data:      "7e000200000123456789017fff0a",
			wantFrame: "",
			wantRest:  "7e000200000123456789017fff0a",
			wantOK:    false,
		},
		{
			name:      "空数据",
			data:      "",
			wantFrame: "",
			wantRest:  "",
			wantOK:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := hex.DecodeString(tt.data)
			frame, rest, ok := TryPopFrame(data)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if fmt.Sprintf("%x", frame) != tt.wantFrame {
				t.Fatalf("frame = %x, want %s", frame, tt.wantFrame)
			}
			if fmt.Sprintf("%x", rest) != tt.wantRest {
				t.Fatalf("rest = %x, want %s", rest, tt.wantRest)
			}
		})
	}
}

func TestFrameReader_ReadFrames(t *testing.T) {
	type want struct {
		frames         []string
		historyDataLen int
	}
	tests := []struct {
		name string
		args string
		want want
	}{
		{
			name: "单个-完整包",
			args: "7e0100002d0144199999990001000b0065373034343358485830303030320000000000000000000000006964303030303301d4c1413138383838927e",
			want: want{
				frames: []string{
					"7e0100002d0144199999990001000b0065373034343358485830303030320000000000000000000000006964303030303301d4c1413138383838927e",
				},
				historyDataLen: 0,
			},
		},
		{
			name: "单个-不完整包",
			args: "7e0100002d0144199999990001000b00653",
			want: want{
				frames:         nil,
				historyDataLen: 17,
			},
		},
		{
			name: "多个-完整包",
			args: "7e000200000123456789017fff0a7e7e0100002d0144199999990001000b0065373034343358485830303030320000000000000000000000006964303030303301d4c1413138383838927e",
			want: want{
				frames: []string{
					"7e000200000123456789017fff0a7e",
					"7e0100002d0144199999990001000b0065373034343358485830303030320000000000000000000000006964303030303301d4c1413138383838927e",
				},
				historyDataLen: 0,
			},
		},
		{
			name: "多个-不完整包",
			args: "7e000200000144199999990007c07e7e800100000144199999990002467e7e0100002d0144199999990001000b00653",
			want: want{
				frames: []string{
					"7e000200000144199999990007c07e",
					"7e800100000144199999990002467e",
				},
				historyDataLen: 17,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewFrameReader()
			argData, _ := hex.DecodeString(tt.args)
			gotFrames := r.ReadFrames(argData)
			if len(gotFrames) != len(tt.want.frames) {
				t.Fatalf("ReadFrames() got len = %v, want %v", len(gotFrames), len(tt.want.frames))
			}
			if r.Pending() != tt.want.historyDataLen {
				t.Fatalf("Pending() = %v, want %v", r.Pending(), tt.want.historyDataLen)
			}
			for k, frame := range gotFrames {
				str := fmt.Sprintf("%x", frame)
				if !reflect.DeepEqual(str, tt.want.frames[k]) {
					t.Fatalf("ReadFrames() got = %s\n want %s", str, tt.want.frames[k])
				}
			}
		})
	}
}

func TestFrameReader_Clear(t *testing.T) {
	r := NewFrameReader()
	incomplete, _ := hex.DecodeString("7e0100002d0144199999990001000b00653")
	r.Append(incomplete)
	if r.Pending() != len(incomplete) {
		t.Fatalf("Pending() before Clear = %v, want %v", r.Pending(), len(incomplete))
	}

	r.Clear()

	if r.Pending() != 0 {
		t.Fatalf("Pending() after Clear = %v, want 0", r.Pending())
	}
	if frame, ok := r.PopFrame(); ok {
		t.Fatalf("PopFrame() after Clear = (%x, true), want (_, false)", frame)
	}

	// Clear 后应能继续正常粘包
	complete, _ := hex.DecodeString("7e000200000123456789017fff0a7e")
	r.Append(complete)
	frame, ok := r.PopFrame()
	if !ok {
		t.Fatal("PopFrame() after Clear and Append = (_, false), want frame")
	}
	if fmt.Sprintf("%x", frame) != "7e000200000123456789017fff0a7e" {
		t.Fatalf("PopFrame() = %x, want 7e000200000123456789017fff0a7e", frame)
	}
	if r.Pending() != 0 {
		t.Fatalf("Pending() after PopFrame = %v, want 0", r.Pending())
	}
}

func TestFrameReader_Append_empty(t *testing.T) {
	r := NewFrameReader()

	r.Append(nil)
	if r.Pending() != 0 {
		t.Fatalf("Append(nil) Pending = %v, want 0", r.Pending())
	}

	r.Append([]byte{})
	if r.Pending() != 0 {
		t.Fatalf("Append([]byte{}) Pending = %v, want 0", r.Pending())
	}

	incomplete, _ := hex.DecodeString("7e0100002d0144199999990001000b00653")
	r.Append(incomplete)
	pendingBefore := r.Pending()

	r.Append(nil)
	r.Append([]byte{})
	if r.Pending() != pendingBefore {
		t.Fatalf("Append empty after incomplete Pending = %v, want %v", r.Pending(), pendingBefore)
	}

	// 空追加不应影响后续正常粘包（与上方 incomplete 拼接后为完整帧）
	suffix, _ := hex.DecodeString("373034343358485830303030320000000000000000000000006964303030303301d4c1413138383838927e")
	r.Append(suffix)
	frame, ok := r.PopFrame()
	if !ok {
		t.Fatal("PopFrame() after split append = (_, false), want frame")
	}
	want := "7e0100002d0144199999990001000b0065373034343358485830303030320000000000000000000000006964303030303301d4c1413138383838927e"
	if fmt.Sprintf("%x", frame) != want {
		t.Fatalf("PopFrame() = %x\n want %s", frame, want)
	}
}

func TestFrameReader_incorrectData(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantFrames  int
		wantPending int
		wantDecode  bool // 是否期望 Decode 报错
	}{
		{
			name:        "空数据",
			input:       "",
			wantFrames:  0,
			wantPending: 0,
		},
		{
			name:        "仅有起始符",
			input:       "7e",
			wantFrames:  0,
			wantPending: 1,
		},
		{
			name:        "仅有结束符",
			input:       "007e",
			wantFrames:  0,
			wantPending: 2,
		},
		{
			name:        "随机垃圾数据",
			input:       "deadbeefcafe",
			wantFrames:  0,
			wantPending: 6,
		},
		{
			name:        "中间含7e但无帧头",
			input:       "01027e0304",
			wantFrames:  0,
			wantPending: 5,
		},
		{
			name:        "尾部两个7e但无帧头-不走快速路径",
			input:       "307e7e",
			wantFrames:  0,
			wantPending: 3,
		},
		{
			name:        "两个7e但内容非法无法Decode",
			input:       "7e000200000123456789017fff7e",
			wantFrames:  0,
			wantPending: 0,
			wantDecode:  true,
		},
		{
			name:        "合法帧后拼接垃圾",
			input:       "7e000200000123456789017fff0a7edeadbeef",
			wantFrames:  1,
			wantPending: 4,
		},
		{
			name:        "第一帧非法第二帧合法-中途Decode失败",
			input:       "7e000200000123456789017fff7e7e000200000123456789017fff0a7e",
			wantFrames:  0,
			wantPending: 15,
			wantDecode:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := hex.DecodeString(tt.input)
			frames, err, pending := unpackLikeService(data)
			if tt.wantDecode {
				if err == nil {
					t.Fatal("unpackLikeService() err = nil, want decode error")
				}
			} else if err != nil {
				t.Fatalf("unpackLikeService() unexpected err = %v", err)
			}
			if len(frames) != tt.wantFrames {
				t.Fatalf("frames len = %d, want %d", len(frames), tt.wantFrames)
			}
			if pending != tt.wantPending {
				t.Fatalf("pending = %d, want %d", pending, tt.wantPending)
			}
			for _, frame := range frames {
				assertExtractedFrame(t, frame)
			}
		})
	}
}

func FuzzFrameReader_ReadFrames(f *testing.F) {
	seeds := [][]byte{
		nil,
		{FrameSign},
		{FrameSign, FrameSign},
		{0x00, 0x01, 0x02},
		{0x01, FrameSign, 0x03, 0x04},
	}
	if b, err := hex.DecodeString("7e000200000123456789017fff0a7e"); err == nil {
		seeds = append(seeds, b)
	}
	if b, err := hex.DecodeString("7e000200000123456789017fff7e"); err == nil {
		seeds = append(seeds, b)
	}
	if b, err := hex.DecodeString("7e0100002d0144199999990001000b00653"); err == nil {
		seeds = append(seeds, b)
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		r := NewFrameReader()
		frames := r.ReadFrames(data)

		for _, frame := range frames {
			assertExtractedFrame(t, frame)
		}

		pending := r.historyData
		assertPendingNoCompleteFrame(t, pending)

		if r.Pending() < 0 {
			t.Fatalf("Pending() = %d, want >= 0", r.Pending())
		}

		// 再追加空数据不应改变状态
		before := r.Pending()
		r.Append(nil)
		r.Append([]byte{})
		if r.Pending() != before {
			t.Fatalf("Pending after empty Append = %d, want %d", r.Pending(), before)
		}
	})
}

func FuzzFrameReader_unpack(f *testing.F) {
	seeds := [][]byte{
		nil,
		{FrameSign, 0x00, FrameSign},
	}
	if b, err := hex.DecodeString("7e000200000123456789017fff0a7e"); err == nil {
		seeds = append(seeds, b)
	}
	if b, err := hex.DecodeString("7e000200000123456789017fff7e"); err == nil {
		seeds = append(seeds, b)
	}
	if b, err := hex.DecodeString("7e000200000123456789017fff7e7e000200000123456789017fff0a7e"); err == nil {
		seeds = append(seeds, b)
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		frames, err, _ := unpackLikeService(data)

		for _, frame := range frames {
			assertExtractedFrame(t, frame)
			// 能走到这里的帧必须 Decode 成功
			if decodeErr := NewJTMessage().Decode(frame); decodeErr != nil {
				t.Fatalf("decoded frame should be valid: %v frame=%x", decodeErr, frame)
			}
		}

		// 解码失败后不应 panic；未消费数据留在 reader，其中可能还有后续完整帧
		if err != nil && len(frames) == 0 {
			r := NewFrameReader()
			r.Append(data)
			for {
				frame, ok := r.PopFrame()
				if !ok {
					break
				}
				if decodeErr := NewJTMessage().Decode(frame); decodeErr != nil {
					if r.Pending() < 0 {
						t.Fatalf("Pending() after decode error = %d, want >= 0", r.Pending())
					}
					return
				}
			}
		}
	})
}
