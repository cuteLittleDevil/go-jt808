package stream

import (
	"encoding/binary"
	"hash/crc32"
)

const (
	streamIDVideo = byte(0xe0)
	streamIDAudio = byte(0xc0)
)

// ProgramStreamPackHeader Table 2-33 https://ocw.unican.es/pluginfile.php/2825/course/section/2777/iso13818-1.pdf.
type ProgramStreamPackHeader struct {
	PackStartCode                 uint32
	Fixed                         byte   // 2 bit 固定0x01
	SystemClockReferenceBase1     byte   // 3 bit
	MarkerBit1                    byte   // 1 bit
	SystemClockReferenceBase2     uint16 // 15 bit
	MarkerBit2                    byte   // 1 bit
	SystemClockReferenceBase3     uint16 // 15 bit
	MarkerBit3                    byte   // 1 bit
	SystemClockReferenceExtension uint16 // 9 bit
	MarkerBit4                    byte   // 1 bit
	ProgramMuxRate                uint32 // 22 bit
	MarkerBit5                    byte   // 1 bit
	MarkerBit6                    byte   // 1 bit
	Reserved                      byte   // 5 bit
	PackStuffingLength            byte   // 3 bit

}

func NewProgramStreamPackHeader(pts uint32) *ProgramStreamPackHeader {
	return &ProgramStreamPackHeader{
		PackStartCode:                 0x000001ba,
		Fixed:                         0x01,
		SystemClockReferenceBase1:     byte(pts >> 30 & 0b0111),
		MarkerBit1:                    1,
		SystemClockReferenceBase2:     uint16(pts >> 15 & 0b01111_1111_1111_1111),
		MarkerBit2:                    1,
		SystemClockReferenceBase3:     uint16(pts & 0b01111_1111_1111_1111),
		MarkerBit3:                    1,
		SystemClockReferenceExtension: 0,
		MarkerBit4:                    1,
		ProgramMuxRate:                0b10111_1101_1010,
		MarkerBit5:                    1,
		MarkerBit6:                    1,
		Reserved:                      31,
		PackStuffingLength:            0,
	}
}

func (p *ProgramStreamPackHeader) Encode() []byte {
	data := make([]byte, 0, 14)
	data = binary.BigEndian.AppendUint32(data, p.PackStartCode)

	// Fixed(2) + SystemClockReferenceBase1(3) + MarkerBit1(1) + SystemClockReferenceBase2(2)
	data = append(data, (p.Fixed<<6&0b1100_0000)|
		p.SystemClockReferenceBase1<<3&0b000011_1000|
		p.MarkerBit1<<2&0b0000_0100|
		byte(p.SystemClockReferenceBase2>>13)&0b0000_0011)

	// SystemClockReferenceBase2(剩余13位中的高8位)
	data = append(data, byte((p.SystemClockReferenceBase2>>5)&0b1111_1111))

	// SystemClockReferenceBase2(剩余5位) + MarkerBit2(1) + SystemClockReferenceBase3(高2位)
	data = append(data, byte(p.SystemClockReferenceBase2<<3&0b1111_1000)|
		p.MarkerBit2<<2|
		byte(p.SystemClockReferenceBase3>>13&0b0000_0011))

	// SystemClockReferenceBase3(剩余13位中的高8位)
	data = append(data, byte(p.SystemClockReferenceBase3>>5&0b1111_1111))

	// SystemClockReferenceBase3(剩余5位) + MarkerBit3(1) + SystemClockReferenceExtension(高2位)
	data = append(data, (byte(p.SystemClockReferenceBase3<<3)&0b1111_1000)|
		p.MarkerBit3<<2|
		(byte(p.SystemClockReferenceExtension>>5)&0b0000_0011))

	// SystemClockReferenceExtension(剩余7位) + MarkerBit4(1)
	data = append(data, byte(p.SystemClockReferenceExtension<<1&0b1111_1110)|p.MarkerBit4)

	// ProgramMuxRate(22位) + MarkerBit5(1) + MarkerBit6(1)
	data = append(data, byte(p.ProgramMuxRate>>14&0b1111_1111))
	data = append(data, byte(p.ProgramMuxRate>>6&0b1111_1111))
	data = append(data, byte(p.ProgramMuxRate<<2)&0b1111_1100|
		p.MarkerBit5<<1|
		p.MarkerBit6)
	//  Reserved(5) + PackStuffingLength(3)
	data = append(data, p.Reserved<<3&0b1111_1000|p.PackStuffingLength&0b0000_0111)

	return data
}

// SystemHeader Table 2-34.
type (
	SystemHeader struct {
		PacketStartCodePrefix     uint32
		MapStreamID               byte
		ProgramStreamMapLength    uint16
		MarkerBit1                byte   // 1 bit
		RateBound                 uint32 // 22 bit
		MarkerBit2                byte   // 1 bit
		AudioBound                byte   // 6 bit
		FixedFlag                 byte   // 1 bit
		CSPSFlag                  byte   // 1 bit
		SystemAudioLockFlag       byte   // 1 bit
		SystemVideoLockFlag       byte   // 1 bit
		MarkerBit3                byte   // 1 bit
		VideoBound                byte   // 5 bit
		PacketRateRestrictionFlag byte   // 1 bit
		ReservedBits              byte   // 7 bit
		Bounds                    []SystemBound
	}

	SystemBound struct {
		StreamID             byte
		MarkerBit            byte   // 固定11
		PTSDBufferBoundScale byte   // 1 bit 1-视频(缓存大小2^14字节) 0-音频(缓存大小2^13字节)
		PTSDBufferSizeBound  uint16 // 13 bit 缓存上限
	}
)

func NewSystemHeader(hasAudio bool) *SystemHeader {
	tmp := &SystemHeader{
		PacketStartCodePrefix:     0x000001,
		MapStreamID:               0xbb,
		ProgramStreamMapLength:    6,
		MarkerBit1:                1,
		RateBound:                 40960,
		MarkerBit2:                1,
		AudioBound:                0,
		FixedFlag:                 0,
		CSPSFlag:                  0,
		SystemAudioLockFlag:       0,
		SystemVideoLockFlag:       0,
		MarkerBit3:                1,
		VideoBound:                1,
		PacketRateRestrictionFlag: 0,
		ReservedBits:              0b0111_1111,
		Bounds: []SystemBound{
			{
				StreamID:             0xe0,
				MarkerBit:            0b11,
				PTSDBufferBoundScale: 1,
				PTSDBufferSizeBound:  2048,
			},
		},
	}
	if hasAudio {
		tmp.AudioBound = 1
		tmp.Bounds = append(tmp.Bounds, SystemBound{
			StreamID:             0xc0,
			MarkerBit:            0b11,
			PTSDBufferBoundScale: 0,
			PTSDBufferSizeBound:  1024,
		})
	}
	tmp.ProgramStreamMapLength += uint16(3 * len(tmp.Bounds))
	return tmp
}

func (s *SystemHeader) Encode() []byte {
	data := make([]byte, 0, 18)
	packetStartCode := s.PacketStartCodePrefix<<8 | uint32(s.MapStreamID)
	data = binary.BigEndian.AppendUint32(data, packetStartCode)
	data = binary.BigEndian.AppendUint16(data, s.ProgramStreamMapLength)

	// MarkerBit1(1) + RateBound(7)
	data = append(data, s.MarkerBit1<<7|byte(s.RateBound>>15))

	// RateBound(8)
	data = append(data, byte(s.RateBound>>7&0b1111_1111))

	// RateBound(7) + MarkerBit2(1)
	data = append(data, byte(s.RateBound<<1&0b1111_1110)|
		s.MarkerBit2&0b0000_0001)

	// AudioBound(6) + FixedFlag(1) + CSPSFlag(1)
	data = append(data, s.AudioBound<<2&0b1111_1100|
		s.FixedFlag<<1&0b0000_0010|
		s.CSPSFlag&0b0000_0001)

	// SystemAudioLockFlag(1) + SystemVideoLockFlag(1) + MarkerBit3(1) + VideoBound(5)
	data = append(data, s.SystemVideoLockFlag<<7&0b1000_0000|
		s.SystemAudioLockFlag<<6&0b0100_0000|
		s.MarkerBit3<<5&0b0010_0000|
		s.VideoBound&0b0001_1111)

	// PacketRateRestrictionFlag(1) + ReservedBits(7)
	data = append(data, s.PacketRateRestrictionFlag<<7&0b1000_0000|
		s.ReservedBits&0b0111_1111)

	for _, v := range s.Bounds {
		data = append(data, v.StreamID)
		// MarkerBit(2) + PTSDBufferBoundScale(1) + PTSDBufferSizeBound(5)
		data = append(data, v.MarkerBit<<6&0b1100_0000|
			v.PTSDBufferBoundScale<<5&0b0010_0000|
			byte(v.PTSDBufferSizeBound>>8)&0b0001_1111)
		// PTSDBufferSizeBound(8)
		data = append(data, byte(v.PTSDBufferSizeBound&0b1111_1111))
	}
	return data
}

// ProgramStreamMap Table 2-35.
type (
	ProgramStreamMap struct {
		PacketStartCodePrefix     uint32
		MapStreamID               byte
		ProgramStreamMapLength    uint16
		CurrentNextIndicator      byte // 1 bit
		Reserved                  byte // 2 bit 扩展目前 0b11
		ProgramStreamMapVersion   byte // 5 bit
		Reserved2                 byte // 7 bit 扩展 目前0b0111_1111
		MarkerBit                 byte
		ProgramStreamInfoLength   uint16
		ElementaryStreamMapLength uint16
		Streams                   []streamMap
		CRC32                     uint32
	}

	streamMap struct {
		// StreamType Table 2-29
		// 0x0F ISO/IEC 13818-7 Audio with ADTS transport syntax
		// 0x10	MPEG-4 视频流
		// 0x1B	H.264 视频流
		// 0x24 H.265 视频流, ISO/IEC 13818-1:2018 增加了这个
		// 0x80	SVAC 视频流
		// 0x90	G.711 音频流
		// 0x92	G.722.1 音频流
		// 0x93	G.723.1 音频流
		// 0x99	G.729 音频流
		// 0x9B	SVAC音频流.
		StreamType uint8
		// ElementaryStreamId 0x(C0~DF)指音频, 0x(E0~EF)为视频.
		ElementaryStreamId         uint8
		ElementaryStreamInfoLength uint16 // 16bit, 指出紧跟在该字段后的描述的字节长度
	}
)

func NewProgramStreamMap(streams ...streamMap) *ProgramStreamMap {
	tmp := &ProgramStreamMap{
		PacketStartCodePrefix:     0x000001,
		MapStreamID:               0xbc,
		ProgramStreamMapLength:    10,
		CurrentNextIndicator:      1,
		Reserved:                  0b0011,
		ProgramStreamMapVersion:   0,
		Reserved2:                 0b0111_1111,
		MarkerBit:                 1,
		ProgramStreamInfoLength:   0,
		ElementaryStreamMapLength: 0,
		//Streams: []streamMap{
		//	{
		//		StreamType:                 0x1b,
		//		ElementaryStreamId:         0xe0,
		//		ElementaryStreamInfoLength: 0,
		//	},
		//	{
		//		StreamType:                 0x90,
		//		ElementaryStreamId:         0xc0,
		//		ElementaryStreamInfoLength: 0,
		//	},
		//},
		Streams: streams,
	}
	tmp.ElementaryStreamMapLength = uint16(4 * len(tmp.Streams))
	tmp.ProgramStreamMapLength += tmp.ElementaryStreamMapLength
	return tmp
}

func (p *ProgramStreamMap) Encode() []byte {
	data := make([]byte, 0, 24)
	packetStartCode := p.PacketStartCodePrefix<<8 | uint32(p.MapStreamID)
	data = binary.BigEndian.AppendUint32(data, packetStartCode)
	data = binary.BigEndian.AppendUint16(data, p.ProgramStreamMapLength)

	// CurrentNextIndicator(1) + Reserved(2) + ProgramStreamMapVersion(5)
	data = append(data, p.CurrentNextIndicator<<7&0b1000_0000|
		p.Reserved<<5&0b0110_0000|
		p.ProgramStreamMapVersion&0b0001_1111)

	// Reserved2(7) + MarkerBit(1)
	data = append(data, p.Reserved2<<1&0b1111_1110|
		p.MarkerBit&0b0000_0001)
	data = binary.BigEndian.AppendUint16(data, p.ProgramStreamInfoLength)
	data = binary.BigEndian.AppendUint16(data, p.ElementaryStreamMapLength)

	for _, v := range p.Streams {
		data = append(data, v.StreamType)
		data = append(data, v.ElementaryStreamId)
		data = binary.BigEndian.AppendUint16(data, v.ElementaryStreamInfoLength)
	}

	p.CRC32 = crc32.Checksum(data, crc32.MakeTable(crc32.IEEE))
	data = binary.BigEndian.AppendUint32(data, p.CRC32)
	return data
}

// PESPacket Table 2-17.
type PESPacket struct {
	PacketStartCodePrefix uint32 // 24 bits 同跟随它的 stream_id 一起组成标识包起始端的包起始码
	StreamID              byte   // 8 bits stream_id 指示基本流的类型和编号
	// PesPacketLength 16 bits 指示 PES 包中跟随该字段最后字节的字节数.0->指示 PES 包长度既未指示也未限定并且仅在这样的PES包中才被允许,
	// 该 PES 包的有效载荷由来自传输流包中所包含的视频基本流的字节组成
	// 离下一个pes包的长度 即00 00 01 0e xx xx到00 00 01 0e之间的.
	PesPacketLength      uint16
	ConstTen             byte // 2 bits 常量10
	PesScramblingControl byte // 2 bit
	PesPriority          byte // 1 bit 指示在此 PES 包中该有效载荷的优先级
	// DataAlignmentIndicator 1 bit 1->指示 PES 包头之后紧随 2.6.10 中
	//data_stream_alignment_descriptor字段中指示的视频句法单元或音频同步字
	//只要该描述符字段存在.若置于值"1"并且该描述符不存在,则要求表 2-53,表 2-54 或
	//表 2-55 的 alignment_type"01"中所指示的那种校准.0->不能确定任何此类校准是否发生.
	DataAlignmentIndicator byte
	Copyright              byte // 1 bit 1->指示相关 PES 包有效载荷的素材依靠版权所保护.0->不能确定该素材是否依靠版权所保护
	OriginalOrCopy         byte // 1 bit 1->指示相关 PES 包有效载荷的内容是原始的.0->指示相关 PES 包有效载荷的内容是复制的
	// PtsDtsFlags 2 bits 10->PES 包头中 PTS 字段存在. 11->PES 包头中 PTS 字段和 DTS 字段均存在.
	//00->PES 包头中既无任何 PTS 字段也无任何 DTS 字段存在. 01->禁用.
	PtsDtsFlags            byte
	EscrFlag               byte // 1 bit 1->指示 PES 包头中 ESCR 基准字段和 ESCR 扩展字段均存在.0->指示无任何 ESCR 字段存在
	EsRateFlag             byte // 1 bit 1->指示 PES 包头中 ES_rate 字段存在.0->指示无任何 ES_rate 字段存在
	DsmTrickModeFlag       byte // 1 bit 1->指示 8 比特特技方式字段存在.0->指示此字段不存在
	AdditionalCopyInfoFlag byte // 1 bit 1->指示 additional_copy_info 存在.0->时指示此字段不存在
	PesCRCFlag             byte // 1 bit 1->指示 PES 包中 CRC 字段存在.0->指示此字段不存在
	PesExtensionFlag       byte // 1 bit 1->时指示 PES 包头中扩展字段存在.0->指示此字段不存在
	// PesHeaderDataLength 8 bits 指示在此 PES包头中包含的由任选字段和任意填充字节所占据的字节总数.
	//任选字段的存在由前导 PES_header_data_length 字段的字节来指定.
	PesHeaderDataLength byte

	// Optional Field
	PTS uint32
	DTS uint32
}

func NewPESPacket(streamID byte, bodyLen uint16, pts uint32) *PESPacket {
	pes := &PESPacket{
		PacketStartCodePrefix:  0x000001,
		StreamID:               streamID,
		PesPacketLength:        3 + 5 + bodyLen,
		ConstTen:               0b10,
		PesScramblingControl:   0,
		PesPriority:            1,
		DataAlignmentIndicator: 1,
		Copyright:              0,
		OriginalOrCopy:         0,
		PtsDtsFlags:            0b10,
		EscrFlag:               0,
		EsRateFlag:             0,
		DsmTrickModeFlag:       0,
		AdditionalCopyInfoFlag: 0,
		PesCRCFlag:             0,
		PesExtensionFlag:       0,
		PesHeaderDataLength:    5,
		PTS:                    pts,
		DTS:                    0,
	}
	if streamID == streamIDAudio {
		pes.DataAlignmentIndicator = 0
	}
	return pes
}

func (p *PESPacket) Encode() []byte {
	data := make([]byte, 0, 19)
	packetStartCode := p.PacketStartCodePrefix<<8 | uint32(p.StreamID)
	data = binary.BigEndian.AppendUint32(data, packetStartCode)
	data = binary.BigEndian.AppendUint16(data, p.PesPacketLength)

	// ConstTen(2) + PesScramblingControl(2) + PesPriority(1) +
	// DataAlignmentIndicator(1) + Copyright(1) + OriginalOrCopy(1)
	data = append(data, p.ConstTen<<6&0b1100_0000|
		p.PesScramblingControl<<4&0b0011_0000|
		p.PesPriority<<3&0b0000_1000|
		p.DataAlignmentIndicator<<2&0b0000_0100|
		p.Copyright<<1&0b0000_0010|
		p.OriginalOrCopy&0b0000_0001)

	// PtsDtsFlags(2) + ESCRFlag(1) + EsRateFlag(1) + DsmTrickModeFlag(1) +
	// AdditionalCopyInfoFlag(1) + PesCRCFlag(1) + PesExtensionFlag(1)
	data = append(data, p.PtsDtsFlags<<6&0b1100_0000|
		p.EscrFlag<<5&0b0010_0000|
		p.EsRateFlag<<4&0b0001_0000|
		p.DsmTrickModeFlag<<3&0b0000_1000|
		p.AdditionalCopyInfoFlag<<2&0b0000_0100|
		p.PesCRCFlag<<1&0b0000_0010|
		p.PesExtensionFlag&0b0000_0001)

	data = append(data, p.PesHeaderDataLength)
	//  仅当PTS存在
	// 固定4位 + PTS(32-30) + MarkerBit1(1)
	data = append(data, 0b0010_0000| // PTS 0010 PTS和DTS 0011
		byte(p.PTS>>29)&0b0000_1110|
		byte(1),
	)
	// PTS(29-15) + MarkerBit2(1)
	data = binary.BigEndian.AppendUint16(data, uint16(p.PTS>>14&0b1111_1111_1111_1110|1))

	// PTS(14-0) + MarkerBit3(1)
	data = binary.BigEndian.AppendUint16(data, uint16(p.PTS<<1&0b1111_1111_1111_1110|1))

	//// 固定4位 + DTS(32-30) + MarkerBit4(1)
	//data = append(data, 0b0001_0000|
	//	byte(p.DTS>>29)&0b0000_1110|
	//	byte(1),
	//)
	//// DTS(29-15) + MarkerBit5(1)
	//data = binary.BigEndian.AppendUint16(data, uint16(p.DTS>>14&0b1111_1111_1111_1110|1))
	//
	//// DTS(14-0) + MarkerBit6(1)
	//data = binary.BigEndian.AppendUint16(data, uint16(p.DTS<<1&0b1111_1111_1111_1110|1))
	return data
}
