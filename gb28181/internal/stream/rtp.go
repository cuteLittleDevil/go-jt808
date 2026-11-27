package stream

import (
	"encoding/binary"
	"github.com/pion/rtp"
)

func createRTPPacket(payload []byte, ops ...func(*rtp.Packet)) []byte {
	rtpPacket := &rtp.Packet{
		Header: rtp.Header{
			Version:          2,
			PayloadType:      96,
			SequenceNumber:   1,
			ExtensionProfile: 1,
			CSRC:             []uint32{},
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
