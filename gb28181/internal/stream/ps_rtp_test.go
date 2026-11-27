package stream

import (
	"encoding/binary"
	"hash/crc32"
	"math"
	"testing"

	"github.com/pion/rtp"
)

func TestProgramStreamPackHeader_Encode(t *testing.T) {
	psh := NewProgramStreamPackHeader(0x12345678)
	// 验证 9-bit SystemClockReferenceExtension 高 2 位 (>>7) 与低 7 位编码正确
	psh.SystemClockReferenceExtension = 0b1_1010_0101 // 高2位=0b11, 低7位=0b010_0101

	encoded := psh.Encode()
	if len(encoded) != 14 {
		t.Fatalf("len(encoded) = %d, want 14", len(encoded))
	}
	if got := binary.BigEndian.Uint32(encoded[0:4]); got != 0x000001ba {
		t.Fatalf("start code = 0x%08x, want 0x000001ba", got)
	}
	// byte 8 低 2 位应为 extension 高 2 位 (0b11)，markerBit3(bit2)=1
	if got := encoded[8] & 0b0000_0111; got != 0b0000_0111 {
		t.Fatalf("byte[8] low 3 bits = 0b%03b, want 0b111", got)
	}
	// byte 9 应为 extension 低 7 位左移 1 位 | markerBit4(1)
	wantByte9 := byte((0b010_0101 << 1) | 1)
	if encoded[9] != wantByte9 {
		t.Fatalf("byte[9] = 0b%08b, want 0b%08b", encoded[9], wantByte9)
	}
}

func TestSystemHeader_Encode(t *testing.T) {
	sh := NewSystemHeader(true)
	sh.SystemAudioLockFlag = 1
	sh.SystemVideoLockFlag = 0

	encoded := sh.Encode()
	// 4(start) + 2(len) + 6(header) + 2*3(bounds) = 18
	if len(encoded) != 18 {
		t.Fatalf("len(encoded) = %d, want 18", len(encoded))
	}
	if got := binary.BigEndian.Uint32(encoded[0:4]); got != 0x000001bb {
		t.Fatalf("start code = 0x%08x, want 0x000001bb", got)
	}
	if got := binary.BigEndian.Uint16(encoded[4:6]); got != 12 {
		t.Fatalf("header length = %d, want 12", got)
	}
	// byte 10: SystemAudioLockFlag(bit7)=1, SystemVideoLockFlag(bit6)=0, MarkerBit3(bit5)=1, VideoBound(bit4..0)=1
	wantByte10 := byte(0b1010_0001)
	if encoded[10] != wantByte10 {
		t.Fatalf("byte[10] = 0b%08b, want 0b%08b (audioLock=bit7, videoLock=bit6)", encoded[10], wantByte10)
	}

	shNoAudio := NewSystemHeader(false)
	if len(shNoAudio.Encode()) != 15 {
		t.Fatalf("len(noAudio) = %d, want 15", len(shNoAudio.Encode()))
	}
}

func TestProgramStreamMap_Encode(t *testing.T) {
	psm := NewProgramStreamMap(
		streamMap{StreamType: 0x1b, ElementaryStreamId: streamIDVideo},
		streamMap{StreamType: 0x90, ElementaryStreamId: streamIDAudio},
	)
	encoded := psm.Encode()
	// 4(start) + 2(len) + 10(fixed+crc) + 8(2 streams) = 24
	if len(encoded) != 24 {
		t.Fatalf("len(encoded) = %d, want 24", len(encoded))
	}
	if got := binary.BigEndian.Uint32(encoded[0:4]); got != 0x000001bc {
		t.Fatalf("start code = 0x%08x, want 0x000001bc", got)
	}
	wantCRC := crc32.Checksum(encoded[:len(encoded)-4], crc32.MakeTable(crc32.IEEE))
	if gotCRC := binary.BigEndian.Uint32(encoded[len(encoded)-4:]); gotCRC != wantCRC {
		t.Fatalf("CRC32 = 0x%08x, want 0x%08x", gotCRC, wantCRC)
	}
}

func TestPESPacket_Encode(t *testing.T) {
	tests := []struct {
		name      string
		streamID  byte
		bodyLen   uint16
		wantLen   uint16
		wantAlign byte
	}{
		{name: "普通视频帧", streamID: streamIDVideo, bodyLen: 100, wantLen: 108, wantAlign: 1},
		{name: "H265视频帧", streamID: streamIDVideoH265, bodyLen: 200, wantLen: 208, wantAlign: 1},
		{name: "最大边界65527", streamID: streamIDVideo, bodyLen: math.MaxUint16 - 8, wantLen: math.MaxUint16, wantAlign: 1},
		{name: "溢出边界65528置0", streamID: streamIDVideo, bodyLen: math.MaxUint16 - 7, wantLen: 0, wantAlign: 1},
		{name: "超大帧传0保持0", streamID: streamIDVideo, bodyLen: 0, wantLen: 0, wantAlign: 1},
		{name: "音频G711A", streamID: streamIDAudio, bodyLen: 160, wantLen: 168, wantAlign: 0},
		{name: "音频G711U", streamID: streamIDAudioG711U, bodyLen: 160, wantLen: 168, wantAlign: 0},
		{name: "音频AAC", streamID: streamIDAudioAAC, bodyLen: 256, wantLen: 264, wantAlign: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pes := NewPESPacket(tt.streamID, tt.bodyLen, 90000)
			if pes.PesPacketLength != tt.wantLen {
				t.Fatalf("PesPacketLength = %d, want %d", pes.PesPacketLength, tt.wantLen)
			}
			if pes.DataAlignmentIndicator != tt.wantAlign {
				t.Fatalf("DataAlignmentIndicator = %d, want %d", pes.DataAlignmentIndicator, tt.wantAlign)
			}
			encoded := pes.Encode()
			if len(encoded) != 14 {
				t.Fatalf("len(encoded) = %d, want 14", len(encoded))
			}
			if got := binary.BigEndian.Uint16(encoded[4:6]); got != tt.wantLen {
				t.Fatalf("encoded PesPacketLength = %d, want %d", got, tt.wantLen)
			}
			if encoded[3] != tt.streamID {
				t.Fatalf("encoded StreamID = 0x%02x, want 0x%02x", encoded[3], tt.streamID)
			}
		})
	}
}

func TestCreateRTPPacket(t *testing.T) {
	payload := []byte("hello-ps-payload")
	frame := createRTPPacket(payload, func(p *rtp.Packet) {
		p.PayloadType = 96
		p.SequenceNumber = 42
		p.Timestamp = 90000
		p.SSRC = 0x11223344
		p.Marker = true
	})

	rtpLen := int(binary.BigEndian.Uint16(frame[:2]))
	if rtpLen != len(frame)-2 {
		t.Fatalf("TCP RTP length prefix = %d, want %d", rtpLen, len(frame)-2)
	}
	var parsed rtp.Packet
	if err := parsed.Unmarshal(frame[2:]); err != nil {
		t.Fatalf("rtp.Unmarshal() error = %v", err)
	}
	if parsed.PayloadType != 96 || parsed.SequenceNumber != 42 || parsed.Timestamp != 90000 || parsed.SSRC != 0x11223344 || !parsed.Marker {
		t.Fatalf("unexpected parsed RTP header: %+v", parsed.Header)
	}
	if string(parsed.Payload) != string(payload) {
		t.Fatalf("payload = %q, want %q", parsed.Payload, payload)
	}
}
