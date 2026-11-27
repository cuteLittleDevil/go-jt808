package stream

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"github.com/pion/rtp"
)

// buildJT1078Frame 构造一个合法的 JT1078 帧用于单测。
func buildJT1078Frame(pt jt1078.PTType, dataType jt1078.DataType, subType jt1078.SubcontractType, ts uint64, body []byte) []byte {
	buf := make([]byte, 0, 30+len(body))
	buf = append(buf, 0x30, 0x31, 0x63, 0x64) // 帧头标识
	buf = append(buf, 0x81)                   // V=2, P=0, X=0, CC=1
	buf = append(buf, byte(pt)&0x7f)          // M=0, PT
	buf = binary.BigEndian.AppendUint16(buf, 1)
	buf = append(buf, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01) // SIM BCD[6]
	buf = append(buf, 1)                                  // LogicChannel
	buf = append(buf, byte(dataType)<<4|byte(subType)&0x0f)
	if dataType != jt1078.DataTypePenetrate {
		buf = binary.BigEndian.AppendUint64(buf, ts)
	}
	if dataType == jt1078.DataTypeI || dataType == jt1078.DataTypeP || dataType == jt1078.DataTypeB {
		buf = binary.BigEndian.AppendUint16(buf, 40) // LastIFrameInterval
		buf = binary.BigEndian.AppendUint16(buf, 40) // LastFrameInterval
	}
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(body)))
	buf = append(buf, body...)
	return buf
}

func TestPackageParse_AtomicFragmentAndFilter(t *testing.T) {
	p := newPackageParse(false)
	defer p.clear()

	// 1. 分包重组：First + Middle + Last
	f1 := buildJT1078Frame(jt1078.PTH264, jt1078.DataTypeI, jt1078.SubcontractTypeFirst, 1000, []byte("part1-"))
	f2 := buildJT1078Frame(jt1078.PTH264, jt1078.DataTypeI, jt1078.SubcontractTypeMiddle, 1000, []byte("part2-"))
	f3 := buildJT1078Frame(jt1078.PTH264, jt1078.DataTypeI, jt1078.SubcontractTypeLast, 1000, []byte("part3"))

	streamData := append(append(f1, f2...), f3...)
	// 故意先喂半截，再喂剩余，验证半包缓存
	var got []*jt1078.Packet
	for pkt, err := range p.parse(streamData[:10]) {
		if err == nil && pkt != nil {
			got = append(got, pkt)
		}
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 packets for partial header, got %d", len(got))
	}
	for pkt, err := range p.parse(streamData[10:]) {
		if err != nil {
			t.Fatalf("parse() error = %v", err)
		}
		got = append(got, pkt)
	}
	if len(got) != 1 || string(got[0].Body) != "part1-part2-part3" {
		t.Fatalf("reassembled body = %q, want %q", got[0].Body, "part1-part2-part3")
	}

	// 2. 开启过滤时跳过前缀脏数据
	pf := newPackageParse(true)
	defer pf.clear()
	dirty := append([]byte{0xde, 0xad, 0xbe, 0xef}, buildJT1078Frame(jt1078.PTG711A, jt1078.DataTypeA, jt1078.SubcontractTypeAtomic, 2000, []byte("audio"))...)
	var filtered []*jt1078.Packet
	for pkt, err := range pf.parse(dirty) {
		if err != nil {
			t.Fatalf("filtered parse() error = %v", err)
		}
		filtered = append(filtered, pkt)
	}
	if len(filtered) != 1 || string(filtered[0].Body) != "audio" {
		t.Fatalf("filtered packets = %+v", filtered)
	}

	// 3. 未开启过滤时遇到非法数据返回错误
	pStrict := newPackageParse(false)
	var gotErr error
	for _, err := range pStrict.parse([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}) {
		gotErr = err
	}
	if gotErr == nil {
		t.Fatal("expected error for unqualified data when hasFilterPacket=false")
	}
}

func TestPackageParse_InterleavedAudioVideoNoAliasOverwrite(t *testing.T) {
	// 复现单次 TCP 粘包中夹杂 [视频First][音频Atomic][视频Last]：
	// 验证视频分包 append 不会沿 historyData 的 cap 覆写已 yield 的音频帧 Body。
	videoPart1 := bytes.Repeat([]byte{0x11}, 100)
	audioBody := bytes.Repeat([]byte{0xAA}, 160)
	videoPart2 := bytes.Repeat([]byte{0x22}, 500)

	fVideoFirst := buildJT1078Frame(jt1078.PTH264, jt1078.DataTypeI, jt1078.SubcontractTypeFirst, 1000, videoPart1)
	fAudio := buildJT1078Frame(jt1078.PTG711A, jt1078.DataTypeA, jt1078.SubcontractTypeAtomic, 1020, audioBody)
	fVideoLast := buildJT1078Frame(jt1078.PTH264, jt1078.DataTypeI, jt1078.SubcontractTypeLast, 1000, videoPart2)

	// 拼到同一个底层数组，模拟一次 conn.Read 粘包读出
	batch := make([]byte, 0, len(fVideoFirst)+len(fAudio)+len(fVideoLast))
	batch = append(batch, fVideoFirst...)
	batch = append(batch, fAudio...)
	batch = append(batch, fVideoLast...)

	conv := NewJT1078T0GB28181()
	packs, err := conv.getPackets(batch)
	if err != nil {
		t.Fatalf("getPackets() error = %v", err)
	}
	if len(packs) != 2 {
		t.Fatalf("len(packs) = %d, want 2", len(packs))
	}
	if !bytes.Equal(packs[0].Body, audioBody) {
		t.Fatalf("audio packet body was overwritten: got first 8 bytes %x, want %x", packs[0].Body[:8], audioBody[:8])
	}
	wantVideo := append(append([]byte(nil), videoPart1...), videoPart2...)
	if !bytes.Equal(packs[1].Body, wantVideo) {
		t.Fatalf("video packet body len = %d, want %d", len(packs[1].Body), len(wantVideo))
	}
}

func TestJT1078ToGB28181_ConvertAndStreamIDs(t *testing.T) {
	conv := NewJT1078T0GB28181()
	info := &command.InviteInfo{
		SSRC: "0100000001",
		IP:   "127.0.0.1",
		Port: 20000,
	}
	info.JT1078Info.Sim = "1001"
	info.JT1078Info.Channel = 1
	info.JT1078Info.Port = 19900
	info.JT1078Info.StreamTypes = []jt1078.PTType{
		jt1078.PTH264,
		jt1078.PTH265,
		jt1078.PTG711A,
		jt1078.PTG711U,
		jt1078.PTAAC,
		jt1078.PTType(127), // 未知类型应被忽略，不写入 PSM
	}
	info.JT1078Info.RtpTypeConvert = func(pt jt1078.PTType) (byte, bool) {
		if pt == jt1078.PTH264 {
			return 96, true
		}
		return 0, false
	}
	conv.OnAck(info)
	defer conv.OnBye("test done")

	if !conv.hasAudio {
		t.Fatal("expected hasAudio=true")
	}
	sm := conv.createStreamMap()
	if len(sm) != 5 {
		t.Fatalf("len(createStreamMap) = %d, want 5 (unknown PT filtered)", len(sm))
	}

	// 验证 I 帧封装（PSH + SYS + PSM + PES + body），且自定义 PT=96 生效
	iFrame := buildJT1078Frame(jt1078.PTH264, jt1078.DataTypeI, jt1078.SubcontractTypeAtomic, 1000, bytes.Repeat([]byte{0x65}, 2000))
	rtps, err := conv.ConvertToGB28181(iFrame)
	if err != nil {
		t.Fatalf("ConvertToGB28181(I) error = %v", err)
	}
	if len(rtps) != 2 {
		t.Fatalf("expected 2 RTP chunks for 2000B frame, got %d", len(rtps))
	}
	var p0, p1 rtp.Packet
	if err := p0.Unmarshal(rtps[0][2:]); err != nil {
		t.Fatalf("unmarshal rtp[0] error = %v", err)
	}
	if err := p1.Unmarshal(rtps[1][2:]); err != nil {
		t.Fatalf("unmarshal rtp[1] error = %v", err)
	}
	if p0.PayloadType != 96 || p0.Marker || !p1.Marker || p1.SequenceNumber != p0.SequenceNumber+1 {
		t.Fatalf("unexpected RTP headers: p0=%+v p1=%+v", p0.Header, p1.Header)
	}
	// I 帧首包应包含 PSH(000001ba)、SYS(000001bb)、PSM(000001bc) 和 H264 PES(000001e0)
	for _, sig := range [][]byte{{0x00, 0x00, 0x01, 0xba}, {0x00, 0x00, 0x01, 0xbb}, {0x00, 0x00, 0x01, 0xbc}, {0x00, 0x00, 0x01, 0xe0}} {
		if !bytes.Contains(p0.Payload, sig) {
			t.Fatalf("I-frame payload missing signature %x", sig)
		}
	}

	// 验证 H265 / G711U / AAC 的 PES StreamID 与 PSM 一致
	codecChecks := []struct {
		pt         jt1078.PTType
		dataType   jt1078.DataType
		wantStream byte
	}{
		{jt1078.PTH265, jt1078.DataTypeP, streamIDVideoH265},
		{jt1078.PTG711A, jt1078.DataTypeA, streamIDAudio},
		{jt1078.PTG711U, jt1078.DataTypeA, streamIDAudioG711U},
		{jt1078.PTAAC, jt1078.DataTypeA, streamIDAudioAAC},
	}
	for _, cc := range codecChecks {
		frame := buildJT1078Frame(cc.pt, cc.dataType, jt1078.SubcontractTypeAtomic, 2000, []byte{0x01, 0x02, 0x03, 0x04})
		out, err := conv.ConvertToGB28181(frame)
		if err != nil || len(out) != 1 {
			t.Fatalf("pt=%v ConvertToGB28181 failed: len=%d err=%v", cc.pt, len(out), err)
		}
		var pkt rtp.Packet
		if err := pkt.Unmarshal(out[0][2:]); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		// PSH 固定 14 字节，紧接着 PES 起始码 00 00 01 <streamID>
		if pkt.Payload[17] != cc.wantStream {
			t.Fatalf("pt=%v PES streamID = 0x%02x, want 0x%02x", cc.pt, pkt.Payload[17], cc.wantStream)
		}
	}

	// 验证超大帧 (>65527) 的 PES 长度字段为 0 而非溢出
	bigBody := bytes.Repeat([]byte{0xaa}, 66000)
	bigRTPs := conv.jt1078ToGB28181(&jt1078.Packet{
		Flag:      jt1078.Flag{PT: jt1078.PTH264},
		DataType:  jt1078.DataTypeP,
		Timestamp: 3000,
		Body:      bigBody,
	})
	var bigFirst rtp.Packet
	if err := bigFirst.Unmarshal(bigRTPs[0][2:]); err != nil {
		t.Fatalf("unmarshal big frame error = %v", err)
	}
	if pesLen := binary.BigEndian.Uint16(bigFirst.Payload[18:20]); pesLen != 0 {
		t.Fatalf("oversized frame PES length = %d, want 0", pesLen)
	}
}

func TestRelayPS(t *testing.T) {
	r := NewRelayPS()
	r.OnAck(&command.InviteInfo{SSRC: "123456"})
	if r.ssrc32 != 123456 {
		t.Fatalf("ssrc32 = %d, want 123456", r.ssrc32)
	}
	out, err := r.ConvertToGB28181([]byte("raw-ps"))
	if err != nil || len(out) != 1 || string(out[0]) != "raw-ps" {
		t.Fatalf("ConvertToGB28181() = (%v, %v)", out, err)
	}
	r.OnBye("bye")
}
