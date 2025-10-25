package stream

import (
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"github.com/pion/rtp"
	"log/slog"
	"math/rand/v2"
	"strconv"
	"sync"
	"time"
)

type JT1078ToGB28181 struct {
	// HasFilterPacket 收到jt1078错误包的时候 主动过滤 打印异常报文.
	HasFilterPacket bool
	streamTypes     []jt1078.PTType
	hasAudio        bool
	ssrc32          uint32
	seq             uint16
	sim             string
	convertFunc     func(ptType jt1078.PTType) (byte, bool)
	packHandle      *packageParse
}

func NewJT1078T0GB28181(opts ...func(gb28181 *JT1078ToGB28181)) *JT1078ToGB28181 {
	tmp := &JT1078ToGB28181{}
	for _, opt := range opts {
		opt(tmp)
	}
	tmp.packHandle = newPackageParse(tmp.HasFilterPacket)
	return tmp
}

func (j *JT1078ToGB28181) OnAck(info *command.InviteInfo) {
	ssrc32, err := strconv.ParseUint(info.SSRC, 10, 32)
	if err != nil {
		ssrc32 = uint64(rand.Uint32())
	}
	hasAudio := false
	for _, v := range info.JT1078Info.StreamTypes {
		if v == jt1078.PTG711A || v == jt1078.PTG711U || v == jt1078.PTAAC {
			hasAudio = true
		}
	}
	j.ssrc32 = uint32(ssrc32)
	j.hasAudio = hasAudio
	j.streamTypes = info.JT1078Info.StreamTypes
	j.seq = 0
	j.sim = info.JT1078Info.Sim
	j.convertFunc = info.JT1078Info.RtpTypeConvert

	slog.Info("connect success",
		slog.String("sim", j.sim),
		slog.Int("channel", info.JT1078Info.Channel),
		slog.Any("ssrc32", j.ssrc32),
		slog.Int("jt1078收流", info.JT1078Info.Port),
		slog.String("gb28181推流", fmt.Sprintf("%s:%d", info.IP, info.Port)))
}

func (j *JT1078ToGB28181) OnBye(msg string) {
	j.packHandle.clear()
	slog.Info("jt1078 bye",
		slog.String("sim", j.sim),
		slog.Any("ssrc32", j.ssrc32),
		slog.String("msg", msg))
}

func (j *JT1078ToGB28181) ConvertToGB28181(jt1078Data []byte) ([][]byte, error) {
	if packs, err := j.getPackets(jt1078Data); err != nil {
		return nil, err
	} else if len(packs) > 0 {
		result := make([][]byte, 0, 10*len(packs))
		for _, pack := range packs {
			result = append(result, j.jt1078ToGB28181(pack)...)
		}
		return result, nil
	}
	return nil, nil
}

func (j *JT1078ToGB28181) jt1078ToGB28181(pack *jt1078.Packet) [][]byte {
	streamID := streamIDAudio
	if pack.Flag.PT == jt1078.PTH264 || pack.Flag.PT == jt1078.PTH265 {
		streamID = streamIDVideo
	}
	pts := uint32(pack.Timestamp)
	// 如果jt1078包的时间不准确 就使用本地时间
	//if cur := time.Now().UnixMilli(); uint64(cur)-pack.Timestamp > 10*1000 {
	//	pts = uint32(cur)
	//}

	var (
		offset = 0
		end    = len(pack.Body)
		result = make([][]byte, 0, 1)
		once   sync.Once
	)

	for end > 0 {
		chunkSize := 1350
		//chunkSize = len(pack.Body)
		if offset+chunkSize >= len(pack.Body) {
			chunkSize = len(pack.Body) - offset
		}
		data := make([]byte, 0, 1460)
		once.Do(func() {
			// 第一个包 I帧 psh + sys + psm + pes + h.264
			// 第一个包 P帧或者B帧 psh + pes + h.264
			// 音频 psh + pes + g711
			// gb28181格式 https://blog.csdn.net/fanyun_01/article/details/120537670
			// ps规范 https://ocw.unican.es/pluginfile.php/2825/course/section/2777/iso13818-1.pdf
			psh := NewProgramStreamPackHeader(pts)
			data = append(data, psh.Encode()...)
			if pack.DataType == jt1078.DataTypeI {
				sys := NewSystemHeader(j.hasAudio)
				data = append(data, sys.Encode()...)
				psm := NewProgramStreamMap(j.createStreamMap()...)
				data = append(data, psm.Encode()...)
			}
			pes := NewPESPacket(streamID, uint16(len(pack.Body)), pts)
			data = append(data, pes.Encode()...)
		})
		data = append(data, pack.Body[offset:offset+chunkSize]...)

		offset += chunkSize
		end -= chunkSize
		result = append(result, createRTPPacket(data, func(packet *rtp.Packet) {
			packet.PayloadType = j.getRtpType(pack.Flag.PT)
			packet.SSRC = j.ssrc32
			packet.Timestamp = uint32(time.Now().UnixMilli())
			packet.SequenceNumber = j.seq
			packet.Marker = end == 0
			j.seq++
		}))
	}
	return result
}

func (j *JT1078ToGB28181) getRtpType(pt jt1078.PTType) byte {
	if j.convertFunc != nil {
		// zlm需要pt是96
		if v, ok := j.convertFunc(pt); ok {
			return v
		}
	}
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
	return payloadType
}

func (j *JT1078ToGB28181) createStreamMap() []streamMap {
	if len(j.streamTypes) == 0 {
		return []streamMap{
			{
				StreamType:                 0x1b, // h264
				ElementaryStreamId:         0xe0,
				ElementaryStreamInfoLength: 0,
			},
			{
				StreamType:                 0x90, // g711
				ElementaryStreamId:         0xc0,
				ElementaryStreamInfoLength: 0,
			},
		}
	}
	list := make([]streamMap, 0, 2)
	for _, ptType := range j.streamTypes {
		var tmp streamMap
		switch ptType {
		case jt1078.PTH264:
			tmp = streamMap{
				StreamType:         0x1b, // h264
				ElementaryStreamId: 0xe0,
			}
		case jt1078.PTH265:
			tmp = streamMap{
				StreamType:         0x24, // h265
				ElementaryStreamId: 0xe1,
			}
		case jt1078.PTG711A:
			tmp = streamMap{
				StreamType:         0x90, // g711a
				ElementaryStreamId: 0xc0,
			}
		case jt1078.PTG711U:
			tmp = streamMap{
				StreamType:         0x90, // g711u
				ElementaryStreamId: 0xc1,
			}
		case jt1078.PTAAC:
			tmp = streamMap{
				StreamType:         0x0F, // aac
				ElementaryStreamId: 0xc2,
			}
		}
		list = append(list, tmp)
	}
	return list
}

func (j *JT1078ToGB28181) getPackets(data []byte) ([]*jt1078.Packet, error) {
	packs := make([]*jt1078.Packet, 0, 1)
	for pack, err := range j.packHandle.parse(data) {
		if err == nil {
			if pack != nil {
				packs = append(packs, pack)
			}
		} else if errors.Is(err, jt1078.ErrBodyLength2Short) || errors.Is(err, jt1078.ErrHeaderLength2Short) {
			// 数据长度不够的 忽略
		} else {
			return nil, err
		}
	}
	return packs, nil
}
