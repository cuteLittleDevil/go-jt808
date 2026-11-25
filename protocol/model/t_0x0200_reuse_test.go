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

func TestProtocolModelFixes(t *testing.T) {
	// 1. P0x9212 多补传包解析步长(8字节) + 复用重置
	src9212 := &P0x9212{
		FileNameLen:            5,
		FileName:               "a.jpg",
		FileType:               0,
		UploadResult:           1,
		RetransmitPacketNumber: 3,
		P0x9212RetransmitPacketList: []P0x9212RetransmitPacket{
			{DataOffset: 0, DataLength: 1024},
			{DataOffset: 4096, DataLength: 2048},
			{DataOffset: 10000, DataLength: 512},
		},
	}
	var got9212 P0x9212
	for round := 0; round < 2; round++ {
		if err := got9212.Parse(&jt808.JTMessage{Body: src9212.Encode()}); err != nil {
			t.Fatalf("P0x9212 round %d Parse error = %v", round, err)
		}
		if len(got9212.P0x9212RetransmitPacketList) != 3 {
			t.Fatalf("P0x9212 round %d len = %d, want 3", round, len(got9212.P0x9212RetransmitPacketList))
		}
		for i, want := range src9212.P0x9212RetransmitPacketList {
			if got9212.P0x9212RetransmitPacketList[i] != want {
				t.Fatalf("P0x9212 packet[%d] = %+v, want %+v", i, got9212.P0x9212RetransmitPacketList[i], want)
			}
		}
	}

	// 2. P0x9208 多地方标准(GD/HLJ/HN/SC)往返不 panic
	for _, asType := range []consts.ActiveSafetyType{
		consts.ActiveSafetyJS,
		consts.ActiveSafetyHLJ,
		consts.ActiveSafetyGD,
		consts.ActiveSafetyHN,
		consts.ActiveSafetySC,
	} {
		src := &P0x9208{
			ServerIPLen: 9,
			ServerAddr:  "127.0.0.1",
			TcpPort:     8080,
			UdpPort:     8081,
			P9208AlarmSign: P9208AlarmSign{
				TerminalID:       "id00001",
				Time:             "2024-11-27 15:42:17",
				SerialNumber:     3,
				AttachNumber:     2,
				ActiveSafetyType: asType,
			},
			AlarmID: "alarm-001",
			Reserve: make([]byte, 16),
		}
		got := &P0x9208{P9208AlarmSign: P9208AlarmSign{ActiveSafetyType: asType}}
		if err := got.Parse(&jt808.JTMessage{Body: src.Encode()}); err != nil {
			t.Fatalf("P0x9208[%s] Parse error = %v", asType, err)
		}
		if got.TerminalID != src.TerminalID || got.Time != src.Time || got.SerialNumber != src.SerialNumber ||
			got.AttachNumber != src.AttachNumber || got.AlarmID != src.AlarmID {
			t.Fatalf("P0x9208[%s] mismatch: got=%+v want=%+v", asType, got, src)
		}
	}

	// 3. P0x8302 截断第 2 个答案不 panic + 复用重置 AnswerList 和标志位
	src8302 := &P0x8302{
		Flag:               0xFF,
		QuestionContentLen: 2,
		QuestionContent:    "hi",
		AnswerList: []P0x8302Answer{
			{AnswerID: 1, AnswerContentLen: 2, AnswerContent: "ok"},
		},
	}
	var got8302 P0x8302
	raw8302 := src8302.Encode()
	if err := got8302.Parse(&jt808.JTMessage{Body: raw8302}); err != nil {
		t.Fatalf("P0x8302 first Parse error = %v", err)
	}
	// 复用且标志清零
	raw8302Zero := append([]byte(nil), raw8302...)
	raw8302Zero[0] = 0x00
	if err := got8302.Parse(&jt808.JTMessage{Body: raw8302Zero}); err != nil {
		t.Fatalf("P0x8302 second Parse error = %v", err)
	}
	if len(got8302.AnswerList) != 1 || got8302.Urgent || got8302.TTS || got8302.Bit7Reserve {
		t.Fatalf("P0x8302 stale state after reuse: %+v", got8302)
	}
	// 末尾多 1 字节畸形（不足第二个答案 3 字节头）
	if err := got8302.Parse(&jt808.JTMessage{Body: append(append([]byte(nil), raw8302...), 0x02)}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("P0x8302 truncated 2nd answer err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}

	// 4. P0x8300 复用重置标志位（含 InfoCategory）
	var got8300 P0x8300
	if err := got8300.Parse(&jt808.JTMessage{Body: []byte{0xFF, 'a'}}); err != nil {
		t.Fatalf("P0x8300 first Parse error = %v", err)
	}
	if !got8300.Urgent || got8300.InfoCategory != 1 {
		t.Fatalf("P0x8300 flags not set: %+v", got8300.P0x8300TextFlagDetails)
	}
	if err := got8300.Parse(&jt808.JTMessage{Body: []byte{0x00, 'b'}}); err != nil {
		t.Fatalf("P0x8300 second Parse error = %v", err)
	}
	if got8300.Urgent || got8300.Display || got8300.TTS || got8300.InfoCategory != 0 || got8300.Bit7Reserve {
		t.Fatalf("P0x8300 stale flags after reuse: %+v", got8300.P0x8300TextFlagDetails)
	}

	// 5. P0x8202 支持 2 字节停止跟踪往返 + 长度边界校验
	var got8202 P0x8202
	if err := got8202.Parse(&jt808.JTMessage{Body: (&P0x8202{TimeInterval: 5, TrackValidity: 60}).Encode()}); err != nil {
		t.Fatalf("P0x8202 start Parse error = %v", err)
	}
	stopRaw := (&P0x8202{TimeInterval: 0, TrackValidity: 60}).Encode()
	if len(stopRaw) != 2 {
		t.Fatalf("P0x8202 stop Encode len = %d, want 2", len(stopRaw))
	}
	if err := got8202.Parse(&jt808.JTMessage{Body: stopRaw}); err != nil {
		t.Fatalf("P0x8202 stop Parse error = %v", err)
	}
	if got8202.TimeInterval != 0 || got8202.TrackValidity != 0 {
		t.Fatalf("P0x8202 stop stale fields: %+v", got8202)
	}
	if err := got8202.Parse(&jt808.JTMessage{Body: []byte{0x00, 0x05}}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("P0x8202 non-zero 2-byte err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}

	// 6. P0x8800 支持 4 字节无重传包往返 + 复用重置
	var got8800 P0x8800
	withRetrans := (&P0x8800{MultimediaID: 7, AgainPackageCount: 1, AgainPackageList: []uint16{2}}).Encode()
	_ = got8800.Parse(&jt808.JTMessage{Body: withRetrans})
	_ = got8800.Parse(&jt808.JTMessage{Body: withRetrans})
	if len(got8800.AgainPackageList) != 1 {
		t.Fatalf("P0x8800 AgainPackageList len = %d after reuse, want 1", len(got8800.AgainPackageList))
	}
	noRetrans := (&P0x8800{MultimediaID: 9}).Encode()
	if len(noRetrans) != 4 {
		t.Fatalf("P0x8800 no-retrans Encode len = %d, want 4", len(noRetrans))
	}
	if err := got8800.Parse(&jt808.JTMessage{Body: noRetrans}); err != nil {
		t.Fatalf("P0x8800 4-byte Parse error = %v", err)
	}
	if got8800.MultimediaID != 9 || got8800.AgainPackageCount != 0 || len(got8800.AgainPackageList) != 0 {
		t.Fatalf("P0x8800 4-byte stale state: %+v", got8800)
	}

	// 7. P0x8003 复用重置
	var got8003 P0x8003
	raw8003 := (&P0x8003{OriginalSerialNumber: 1, AgainPackageCount: 1, AgainPackageList: []uint16{3}}).Encode()
	_ = got8003.Parse(&jt808.JTMessage{Body: raw8003})
	_ = got8003.Parse(&jt808.JTMessage{Body: raw8003})
	if len(got8003.AgainPackageList) != 1 {
		t.Fatalf("P0x8003 AgainPackageList len = %d after reuse, want 1", len(got8003.AgainPackageList))
	}

	// 8. TerminalParamDetails 解析 0x018/0x019/0x021 + 复用清空旧字段与 OtherContent
	var p8103 P0x8103
	hookCalled := 0
	p8103.TerminalParamDetails.ParamParseBeforeFunc = func(uint32, []byte) { hookCalled++ }
	body1 := []byte{
		4, // 4 个参数：0x018, 0x019, 0x021, 0x9999(未知)
		0x00, 0x00, 0x00, 0x18, 0x04, 0x00, 0x00, 0x1F, 0x90, // 8080
		0x00, 0x00, 0x00, 0x19, 0x04, 0x00, 0x00, 0x1F, 0x91, // 8081
		0x00, 0x00, 0x00, 0x21, 0x04, 0x00, 0x00, 0x00, 0x01, // 1
		0x00, 0x00, 0x99, 0x99, 0x01, 0xAA,
	}
	if err := p8103.Parse(&jt808.JTMessage{Body: body1}); err != nil {
		t.Fatalf("P0x8103 first Parse error = %v", err)
	}
	if p8103.TerminalParamDetails.T0x018TCPPort.Value != 8080 ||
		p8103.TerminalParamDetails.T0x019UDPPort.Value != 8081 ||
		p8103.TerminalParamDetails.T0x021PositionReportingPlan.Value != 1 ||
		len(p8103.TerminalParamDetails.OtherContent) != 1 {
		t.Fatalf("P0x8103 unexpected parsed params: tcp=%+v udp=%+v plan=%+v other=%v",
			p8103.TerminalParamDetails.T0x018TCPPort,
			p8103.TerminalParamDetails.T0x019UDPPort,
			p8103.TerminalParamDetails.T0x021PositionReportingPlan,
			p8103.TerminalParamDetails.OtherContent)
	}
	if encoded := p8103.Encode(); len(encoded) == 0 {
		t.Fatal("P0x8103 Encode with OtherContent returned empty")
	}
	// 复用解析仅含 0x001 的包，旧的 0x018/0x019/0x021/OtherContent 必须清空，且回调保留
	body2 := []byte{1, 0x00, 0x00, 0x00, 0x01, 0x04, 0x00, 0x00, 0x00, 0x1E}
	if err := p8103.Parse(&jt808.JTMessage{Body: body2}); err != nil {
		t.Fatalf("P0x8103 second Parse error = %v", err)
	}
	if p8103.TerminalParamDetails.T0x001HeartbeatInterval.Value != 30 ||
		p8103.TerminalParamDetails.T0x018TCPPort.ID != 0 ||
		p8103.TerminalParamDetails.T0x021PositionReportingPlan.ID != 0 ||
		len(p8103.TerminalParamDetails.OtherContent) != 0 ||
		hookCalled != 5 {
		t.Fatalf("P0x8103 stale state on reuse: hb=%+v tcp=%+v other=%v hookCalled=%d",
			p8103.TerminalParamDetails.T0x001HeartbeatInterval,
			p8103.TerminalParamDetails.T0x018TCPPort,
			p8103.TerminalParamDetails.OtherContent,
			hookCalled)
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

func TestSecondPassProtocolFixes(t *testing.T) {
	// 1. CustomAdditionContentFunc 支持覆盖标准 ID 非标长度（如厂商自定义 0x06 4 字节），返回 false 则回退标准校验
	var loc T0x0200
	loc.CustomAdditionContentFunc = func(id uint8, content []byte) (AdditionContent, bool) {
		if id == 0x06 && len(content) == 4 {
			return AdditionContent{Data: content, CustomValue: binary.BigEndian.Uint32(content)}, true
		}
		return AdditionContent{}, false
	}
	bodyCustom := append(make([]byte, 28),
		0x06, 0x04, 0x00, 0x00, 0x01, 0x02, // 自定义 4 字节 0x06
		0x30, 0x01, 0x15, // 未被自定义拦截，回退标准解析
		0xEE, 0x01, 0x99, // 未知 ID 回退默认分支
	)
	if err := loc.Parse(&jt808.JTMessage{Body: bodyCustom}); err != nil {
		t.Fatalf("CustomAdditionContentFunc override Parse error = %v", err)
	}
	if got := loc.Additions[consts.JT808LocationAdditionType(0x06)].Content.CustomValue; got != uint32(0x0102) {
		t.Fatalf("custom 0x06 = %v, want 0x0102", got)
	}

	// 2. T0x0100 零值 ProtocolVersion 或 nil Header 短包均不 panic
	var t0100 T0x0100
	if err := t0100.Parse(&jt808.JTMessage{Body: make([]byte, 10)}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("T0x0100 nil Header short body err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}
	if err := t0100.Parse(&jt808.JTMessage{Header: &jt808.Header{}, Body: make([]byte, 10)}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("T0x0100 zero version short body err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}

	// 3. T0x0102 2019 短 IMEI 补零与去尾零 -> 2013 复用清空 -> ReplyBody 无状态
	src0102 := T0x0102{
		Version:         consts.JT808Protocol2019,
		AuthCodeLen:     7,
		AuthCode:        "auth123",
		TerminalIMEI:    "860001",
		SoftwareVersion: "v1.2.3",
	}
	if s := src0102.String(); len(s) == 0 {
		t.Fatal("T0x0102(2019).String() empty")
	}
	var got0102 T0x0102
	msg2019 := &jt808.JTMessage{
		Header: &jt808.Header{ProtocolVersion: consts.JT808Protocol2019, SerialNumber: 5},
		Body:   src0102.Encode(),
	}
	if err := got0102.Parse(msg2019); err != nil {
		t.Fatalf("T0x0102 2019 Parse error = %v", err)
	}
	if got0102.AuthCode != "auth123" || got0102.TerminalIMEI != "860001" || got0102.SoftwareVersion != "v1.2.3" {
		t.Fatalf("T0x0102 2019 mismatch: %+v", got0102)
	}
	// ReplyBody 不修改 receiver
	var stateless0102 T0x0102
	if _, err := stateless0102.ReplyBody(msg2019); err != nil || stateless0102.AuthCode != "" {
		t.Fatalf("T0x0102 ReplyBody mutated receiver or failed: err=%v state=%+v", err, stateless0102)
	}
	// 复用为 2013 模式，2019 专属字段必须清空
	msg2013 := &jt808.JTMessage{
		Header: &jt808.Header{ProtocolVersion: consts.JT808Protocol2013},
		Body:   []byte("k2013"),
	}
	if err := got0102.Parse(msg2013); err != nil {
		t.Fatalf("T0x0102 2013 Parse error = %v", err)
	}
	if got0102.AuthCode != "k2013" || got0102.AuthCodeLen != 0 || got0102.TerminalIMEI != "" || got0102.SoftwareVersion != "" {
		t.Fatalf("T0x0102 2013 stale 2019 fields: %+v", got0102)
	}

	// 4. T0x0801 / T0x1212 ReplyBody 对畸形短包返回错误且不污染 receiver
	var t0801 T0x0801
	if _, err := t0801.ReplyBody(&jt808.JTMessage{Body: []byte{1, 2}}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("T0x0801 short ReplyBody err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}
	var t1212 T0x1212
	if _, err := t1212.ReplyBody(&jt808.JTMessage{Body: []byte{5, 'a'}}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("T0x1212 short ReplyBody err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}

	// 5. T0x0704 Num==0 与尾部多余字节校验
	var t0704 T0x0704
	if err := t0704.Parse(&jt808.JTMessage{Body: []byte{0x00, 0x00, 0x00, 0x00, 0x01}}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("T0x0704 Num==0 err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}
	extraBatch := append([]byte{0x00, 0x01, 0x00, 0x00, 0x1C}, make([]byte, 29)...)
	if err := t0704.Parse(&jt808.JTMessage{Body: extraBatch}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("T0x0704 trailing byte err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}

	// 6. T0x1205 防 32 位乘法溢出校验
	var t1205 T0x1205
	overflow1205 := make([]byte, 6+8) // rem=8，非 28 倍数
	binary.BigEndian.PutUint32(overflow1205[2:6], 0x0924924A)
	if err := t1205.Parse(&jt808.JTMessage{Body: overflow1205}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("T0x1205 overflow err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}
	mismatchCount1205 := make([]byte, 6+28)
	binary.BigEndian.PutUint32(mismatchCount1205[2:6], 2)
	if err := t1205.Parse(&jt808.JTMessage{Body: mismatchCount1205}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("T0x1205 count mismatch err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}

	// 7. T0x1210 黑标复用清空 TerminalID
	var t1210 T0x1210
	t1210.TerminalID = "staleID"
	t1210.P9208AlarmSign.ActiveSafetyType = consts.ActiveSafetyHLJ
	srcHLJ := T0x1210{
		P9208AlarmSign: P9208AlarmSign{
			ActiveSafetyType: consts.ActiveSafetyHLJ,
			TerminalID:       "hlj001",
			Time:             "2024-11-27 15:42:30",
		},
		AlarmID:              "alm1",
		AttachCount:          1,
		T0x1210AlarmItemList: []T0x1210AlarmItem{{FileNameLen: 5, FileName: "v.mp4", FileSize: 128}},
	}
	if err := t1210.Parse(&jt808.JTMessage{Body: srcHLJ.Encode()}); err != nil {
		t.Fatalf("T0x1210 HLJ Parse error = %v", err)
	}
	if t1210.TerminalID != "" || len(t1210.T0x1210AlarmItemList) != 1 || t1210.T0x1210AlarmItemList[0].FileName != "v.mp4" {
		t.Fatalf("T0x1210 HLJ unexpected state: %+v", t1210)
	}

	// 8. P0x8202 超过 6 字节返回错误
	var p8202 P0x8202
	if err := p8202.Parse(&jt808.JTMessage{Body: make([]byte, 7)}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("P0x8202 7-byte err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}

	// 9. TerminalParamDetails 256 个参数但 count=0 不得因 uint8 回绕而误判成功 + P0x9205 空时间条件兼容
	var got8103 P0x8103
	var wrapBuf []byte
	wrapBuf = append(wrapBuf, 0) // count = 0
	for i := 0; i < 256; i++ {
		wrapBuf = append(wrapBuf, 0x00, 0x00, 0x99, byte(i), 0x01, 0x00)
	}
	if err := got8103.Parse(&jt808.JTMessage{Body: wrapBuf}); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Fatalf("P0x8103 256-param count=0 wraparound err = %v, want %v", err, protocol.ErrBodyLengthInconsistency)
	}

	var got9205 P0x9205
	if err := got9205.Parse(&jt808.JTMessage{Body: (&P0x9205{ChannelNo: 1}).Encode()}); err != nil || got9205.ChannelNo != 1 {
		t.Fatalf("manual P0x9205 empty time failed: err=%v got=%+v", err, got9205)
	}

	// 10. T0x0102(2019) AuthCodeLen>=221 与 P0x8302 QuestionContentLen>=254 不触发 uint8 切片索引回绕 panic
	longAuthBody := make([]byte, 1+255+15+20)
	longAuthBody[0] = 255
	if err := got0102.Parse(&jt808.JTMessage{
		Header: &jt808.Header{ProtocolVersion: consts.JT808Protocol2019},
		Body:   longAuthBody,
	}); err != nil || len(got0102.AuthCode) != 255 {
		t.Fatalf("T0x0102 AuthCodeLen=255 failed: err=%v len=%d", err, len(got0102.AuthCode))
	}
	longQuestionBody := make([]byte, 2+255+3)
	longQuestionBody[1] = 255
	var got8302Long P0x8302
	if err := got8302Long.Parse(&jt808.JTMessage{Body: longQuestionBody}); err != nil || len(got8302Long.AnswerList) != 1 {
		t.Fatalf("P0x8302 QuestionContentLen=255 failed: err=%v answers=%d", err, len(got8302Long.AnswerList))
	}
}

func FuzzModelParse(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05})
	f.Add(make([]byte, 36))
	f.Add(make([]byte, 80))

	f.Fuzz(func(t *testing.T, data []byte) {
		for _, ver := range []consts.ProtocolVersionType{
			0,
			consts.JT808Protocol2013,
			consts.JT808Protocol2019,
		} {
			jtMsg := &jt808.JTMessage{
				Header: &jt808.Header{ProtocolVersion: ver},
				Body:   data,
			}
			_ = (&T0x0100{}).Parse(jtMsg)
			_ = (&T0x0102{}).Parse(jtMsg)
			_ = (&T0x0104{}).Parse(jtMsg)
			_ = (&T0x0200{}).Parse(jtMsg)
			_ = (&T0x0704{}).Parse(jtMsg)
			_ = (&T0x0801{}).Parse(jtMsg)
			_ = (&T0x0805{}).Parse(jtMsg)
			_ = (&T0x1205{}).Parse(jtMsg)
			_ = (&T0x1210{}).Parse(jtMsg)
			_ = (&T0x1211{}).Parse(jtMsg)
			_ = (&T0x1212{}).Parse(jtMsg)
			_ = (&P0x8003{}).Parse(jtMsg)
			_ = (&P0x8103{}).Parse(jtMsg)
			_ = (&P0x8202{}).Parse(jtMsg)
			_ = (&P0x8302{}).Parse(jtMsg)
			_ = (&P0x8800{}).Parse(jtMsg)
			_ = (&P0x9101{}).Parse(jtMsg)
			_ = (&P0x9201{}).Parse(jtMsg)
			_ = (&P0x9206{}).Parse(jtMsg)
			_ = (&P0x9208{}).Parse(jtMsg)
			_ = (&P0x9212{}).Parse(jtMsg)
		}
	})
}
