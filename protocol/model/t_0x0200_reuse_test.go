package model

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

func TestT0x0200_ReuseClearsFlagsAndAdditions(t *testing.T) {
	var loc T0x0200

	// 第 1 包：所有报警位、状态位为 1，且携带 0x01 里程 + 0x11 超速报警(AreaID=0x12345678)
	body1 := make([]byte, 28, 42)
	binary.BigEndian.PutUint32(body1[0:4], 0xFFFFFFFF)
	binary.BigEndian.PutUint32(body1[4:8], (1<<23)-1)
	body1 = append(body1,
		0x01, 0x04, 0x00, 0x00, 0x00, 0x64, // 里程 100
		0x11, 0x05, 0x01, 0x12, 0x34, 0x56, 0x78, // 超速：圆形区域 0x12345678
	)
	if err := loc.Parse(&jt808.JTMessage{Body: body1}); err != nil {
		t.Fatalf("first Parse() error = %v", err)
	}
	if !loc.AlarmSignDetails.EmergencyAlarm || !loc.AlarmSignDetails.OverSpeed || !loc.StatusSignDetails.ACC || loc.StatusSignDetails.Cargo != 3 {
		t.Fatalf("first Parse() flags not set: alarm=%+v status=%+v", loc.AlarmSignDetails, loc.StatusSignDetails)
	}
	if got := loc.Additions[consts.A0x11OverSpeedAlarm].Content.OverSpeedAlarm.AreaID; got != 0x12345678 {
		t.Fatalf("0x11 AreaID = 0x%08x, want 0x12345678", got)
	}

	// 第 2 包：同一实例复用解析——所有标志位清零，且仅携带 0x30 信号强度
	body2 := make([]byte, 28, 31)
	body2 = append(body2, 0x30, 0x01, 0x1F)
	if err := loc.Parse(&jt808.JTMessage{Body: body2}); err != nil {
		t.Fatalf("second Parse() error = %v", err)
	}
	if loc.AlarmSignDetails.EmergencyAlarm || loc.AlarmSignDetails.OverSpeed || loc.AlarmSignDetails.LaneOpeningAlarm {
		t.Fatalf("second Parse() stale alarm flags: %+v", loc.AlarmSignDetails)
	}
	if loc.StatusSignDetails.ACC || loc.StatusSignDetails.Location || loc.StatusSignDetails.Cargo != 0 || loc.StatusSignDetails.VehicleRunning {
		t.Fatalf("second Parse() stale status flags: %+v", loc.StatusSignDetails)
	}
	if _, ok := loc.Additions[consts.A0x01Mile]; ok {
		t.Fatal("second Parse() retained stale 0x01 addition")
	}
	if _, ok := loc.Additions[consts.A0x11OverSpeedAlarm]; ok {
		t.Fatal("second Parse() retained stale 0x11 addition")
	}
	if got := loc.Additions[consts.A0x30WIFISignalStrength].Content.WIFISignalStrength; got != 0x1F {
		t.Fatalf("0x30 WIFISignalStrength = %d, want 31", got)
	}

	// 第 3 包：无附加信息（len == 28），旧 Additions 也必须清空
	body3 := make([]byte, 28)
	if err := loc.Parse(&jt808.JTMessage{Body: body3}); err != nil {
		t.Fatalf("third Parse() error = %v", err)
	}
	if len(loc.Additions) != 0 {
		t.Fatalf("third Parse() expected empty Additions, got %d items", len(loc.Additions))
	}
}

func TestT0x0200Addition_MalformedNoPanic(t *testing.T) {
	tests := []struct {
		name string
		add  []byte
	}{
		{name: "0x11_nonZeroLocationType_len1", add: []byte{0x11, 0x01, 0x02}},
		{name: "0x31_len0", add: []byte{0x31, 0x00}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := append(make([]byte, 28), tt.add...)
			var loc T0x0200
			if err := loc.Parse(&jt808.JTMessage{Body: body}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
				t.Fatalf("Parse() err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
			}
		})
	}
}

func TestModels_ReuseResetsSlicesAndBoundsCheck(t *testing.T) {
	// 1. T0x0704 复用不累积 + 畸形 Num 不 panic
	t0704 := &T0x0704{}
	itemBody := make([]byte, 28)
	batch := []byte{0x00, 0x01, 0x00, 0x00, 0x1C}
	batch = append(batch, itemBody...)
	if err := t0704.Parse(&jt808.JTMessage{Body: batch}); err != nil {
		t.Fatalf("T0x0704 first Parse() error = %v", err)
	}
	if err := t0704.Parse(&jt808.JTMessage{Body: batch}); err != nil {
		t.Fatalf("T0x0704 second Parse() error = %v", err)
	}
	if len(t0704.Items) != 1 {
		t.Fatalf("T0x0704 Items len = %d after reuse, want 1", len(t0704.Items))
	}
	badBatch := append([]byte{0x00, 0x02, 0x00, 0x00, 0x1C}, itemBody...)
	if err := t0704.Parse(&jt808.JTMessage{Body: badBatch}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("T0x0704 bad Num err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}

	// 2. T0x0805 复用不累积
	t0805 := &T0x0805{}
	b0805 := []byte{0x00, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x09}
	_ = t0805.Parse(&jt808.JTMessage{Body: b0805})
	_ = t0805.Parse(&jt808.JTMessage{Body: b0805})
	if len(t0805.MultimediaIDList) != 1 {
		t.Fatalf("T0x0805 MultimediaIDList len = %d after reuse, want 1", len(t0805.MultimediaIDList))
	}

	// 3. T0x1205 复用不累积
	t1205 := &T0x1205{}
	b1205 := make([]byte, 6+28)
	binary.BigEndian.PutUint32(b1205[2:6], 1)
	_ = t1205.Parse(&jt808.JTMessage{Body: b1205})
	_ = t1205.Parse(&jt808.JTMessage{Body: b1205})
	if len(t1205.AudioVideoResourceList) != 1 {
		t.Fatalf("T0x1205 AudioVideoResourceList len = %d after reuse, want 1", len(t1205.AudioVideoResourceList))
	}

	// 4. T0x1210 复用不累积 + 畸形长度不 panic
	t1210 := &T0x1210{P9208AlarmSign: P9208AlarmSign{ActiveSafetyType: consts.ActiveSafetyJS}}
	// 苏标头: 7 + 16 + 32 + 1 + 1 = 57 字节
	b1210 := make([]byte, 57)
	b1210[56] = 1 // AttachCount = 1
	b1210 = append(b1210, 0x03, 'a', '.', 'j', 0x00, 0x00, 0x00, 0x10)
	if err := t1210.Parse(&jt808.JTMessage{Body: b1210}); err != nil {
		t.Fatalf("T0x1210 first Parse() error = %v", err)
	}
	if err := t1210.Parse(&jt808.JTMessage{Body: b1210}); err != nil {
		t.Fatalf("T0x1210 second Parse() error = %v", err)
	}
	if len(t1210.T0x1210AlarmItemList) != 1 {
		t.Fatalf("T0x1210 AlarmItemList len = %d after reuse, want 1", len(t1210.T0x1210AlarmItemList))
	}
	// 2 个附件但第 1 个占满剩余字节，使第 2 轮 start == len(body)
	bad1210 := make([]byte, 57)
	bad1210[56] = 2
	bad1210 = append(bad1210, 0x07, '1', '2', '3', '4', '5', '6', '7', 0x00, 0x00, 0x00, 0x10)
	if err := t1210.Parse(&jt808.JTMessage{Body: bad1210}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("T0x1210 truncated err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}
}

func BenchmarkT0x0200Parse(b *testing.B) {
	raw, _ := hex.DecodeString("7e020000800123456789017fff000004000000080006eeb6ad02633df701380003006320070719235901040000000b02020016030200210402002c051e3737370000000000000000000000000000000000000000000000000000001105420000004212064d0000004d4d1307000000580058582504000000632a02000a2b040000001430011e31012806020001927e")
	jtMsg := jt808.NewJTMessage()
	if err := jtMsg.Decode(raw); err != nil {
		b.Fatal(err)
	}
	var loc T0x0200
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = loc.Parse(jtMsg)
	}
}
