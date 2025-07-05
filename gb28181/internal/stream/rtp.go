package stream

import (
	"encoding/binary"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"github.com/pion/rtp"
)

//// RTPPacket rtp包
///*
//https://www.gpssoft.cn/download/protocol/RFC-3550-%E4%B8%AD%E6%96%87%E7%89%88.pdf
// *  0                   1                   2                   3
// *  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// * |V=2|P|X|  CC   |M|     PT      |       sequence number         |
// * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// * |                           timestamp                           |
// * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// * |           synchronization source (SSRC) identifier            |
// * +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
// * |            contributing source (CSRC) identifiers             |
// * |                             ....                              |
// * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//*/
//type RTPPacket struct {
//	Version        byte   // 2 bit，RTP协议版本，当前标准版本为2
//	Padding        byte   // 1 bit，若置位，表示包尾存在填充字节
//	Extension      byte   // 1 bit，指示是否包含RTP头部扩展
//	CSRCCount      byte   // 4 bit，“贡献源”标识符（CSRC）的数目
//	Marker         byte   // 1 bit，对于视频表示结束 对于音频表示开始
//	PayloadType    byte   // 7 bit，指定负载中携带的媒体数据类型
//	SequenceNumber uint16 // 16 bit，用于标识发送者发出的RTP包的顺序
//	Timestamp      uint32 // 32 bit，表示媒体数据的采样时间
//	SSRC           uint32 // 32 bit，标识数据包的发送源
//	//CSRCList       []uint32 // 若CSRCCount非零，这里包含CSRCCount个32位的CSRC标识符
//	Payload []byte // 负载数据，内容取决于PayloadType
//}
//
//func NewRTPPacket(ssrc uint32, pt jt1078.PTType, pts uint32) *RTPPacket {
//	payloadType := byte(96)
//	// GB28181 2016 附录C C.2.2
//	switch pt {
//	case jt1078.PTH264:
//		payloadType = 98
//	case jt1078.PTH265:
//		payloadType = 100
//	case jt1078.PTG711U:
//		payloadType = 0
//	case jt1078.PTG711A:
//		payloadType = 8
//	}
//	return &RTPPacket{
//		Version:     2,
//		SSRC:        ssrc,
//		PayloadType: payloadType,
//		Timestamp:   pts,
//	}
//}
//
//func (r *RTPPacket) Encode() []byte {
//	data := make([]byte, 0, 2+12+len(r.Payload))
//
//	// 都使用TCP r.Payload不会大于1400 不会溢出
//	// 可以换 https://github.com/pion/rtp
//	data = binary.BigEndian.AppendUint16(data, uint16(12+len(r.Payload)))
//
//	// Version(2) + Padding(1) + Extension(1) + CSRCCount(4)
//	data = append(data, r.Version<<6&0b1100_0000|
//		r.Padding<<5&0b0010_0000|
//		r.Extension<<4&0b0001_0000|
//		r.CSRCCount&0b0000_1111)
//
//	// Marker(1) + PayloadType(7)
//	data = append(data, r.Marker<<7&0b1000_0000|
//		r.PayloadType&0b0111_1111)
//
//	data = binary.BigEndian.AppendUint16(data, r.SequenceNumber)
//	data = binary.BigEndian.AppendUint32(data, r.Timestamp)
//	data = binary.BigEndian.AppendUint32(data, r.SSRC)
//
//	data = append(data, r.Payload...)
//	return data
//}

func createRTPPacket(pt jt1078.PTType, payload []byte, ops ...func(*rtp.Packet)) []byte {
	payloadType := byte(96)
	// GB28181 2016 附录C C.2.2 h264推荐98 h265推荐100
	switch pt {
	case jt1078.PTH264:
		payloadType = 98
	case jt1078.PTH265:
		payloadType = 100
	case jt1078.PTG711U:
		payloadType = 0
	case jt1078.PTG711A:
		payloadType = 8
	}

	rtpPacket := &rtp.Packet{
		Header: rtp.Header{
			Padding:          false,
			Marker:           false,
			Extension:        false,
			ExtensionProfile: 1,
			Extensions:       nil,
			Version:          2,
			PayloadType:      payloadType,
			SequenceNumber:   1,
			Timestamp:        0,
			SSRC:             0,
			CSRC:             []uint32{},
			PaddingSize:      0,
		},
		Payload: payload,
	}
	for _, op := range ops {
		op(rtpPacket)
	}
	data := make([]byte, 0, 2+12+len(payload))
	// tcp 需要增加固定头部
	data = binary.BigEndian.AppendUint16(data, uint16(rtpPacket.MarshalSize()))
	body, _ := rtpPacket.Marshal()
	data = append(data, body...)
	return data
}
