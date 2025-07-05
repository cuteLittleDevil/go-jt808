package stream

import (
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"github.com/pion/rtp"
	"io"
	"log/slog"
	"math/rand/v2"
	"net"
	"strconv"
	"sync"
	"time"
)

type jt1078Server struct {
	stopOnce    sync.Once
	stopChan    chan struct{}
	listen      *net.TCPListener
	packetChan  chan []*jt1078.Packet
	ssrc32      uint32
	seq         uint16
	packHandle  *packageParse
	jt1078Port  int
	gb28181IP   string
	gb28181Port int
	streamTypes []jt1078.PTType
	hasAudio    bool
}

func newJt1078Server(info *command.InviteInfo) *jt1078Server {
	ssrc32, err := strconv.ParseUint(info.SSRC, 10, 32)
	if err != nil {
		ssrc32 = rand.Uint64()
	}
	hasAudio := false
	for _, v := range info.JT1078Info.StreamTypes {
		if v == jt1078.PTG711A || v == jt1078.PTG711U || v == jt1078.PTAAC {
			hasAudio = true
		}
	}
	return &jt1078Server{
		stopOnce:    sync.Once{},
		stopChan:    make(chan struct{}),
		listen:      nil,
		packetChan:  make(chan []*jt1078.Packet, 100),
		ssrc32:      uint32(ssrc32),
		seq:         0,
		packHandle:  newPackageParse(),
		jt1078Port:  info.JT1078Info.Port,
		gb28181IP:   info.IP,
		gb28181Port: info.Port,
		streamTypes: info.JT1078Info.StreamTypes,
		hasAudio:    hasAudio,
	}
}

func (j *jt1078Server) run() {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", j.jt1078Port))
	if err != nil {
		slog.Error("resolve tcp addr error",
			slog.Any("err", err))
		return
	}

	in, err := net.ListenTCP("tcp", addr)
	if err != nil {
		slog.Error("tcp listen fail",
			slog.Any("addr", addr),
			slog.Any("err", err))
		return
	}
	j.listen = in

	conn, err := in.AcceptTCP()
	if err == nil {
		go j.readJt1078Packet(conn)
		gb28181Address := fmt.Sprintf("%s:%d", j.gb28181IP, j.gb28181Port)
		gb28181Conn, err := net.Dial("tcp", gb28181Address)
		if err == nil {
			go j.writeGB28181Packet(gb28181Conn)
			slog.Info("connect success",
				slog.Int("jt1078收流", j.jt1078Port),
				slog.String("gb28181推流", gb28181Address))
		}
	}
}

func (j *jt1078Server) stop() {
	j.stopOnce.Do(func() {
		close(j.stopChan)
		if j.listen != nil {
			_ = j.listen.Close()
		}
		time.Sleep(time.Second)
		close(j.packetChan)
		j.packHandle.clear()
	})
}

func (j *jt1078Server) readJt1078Packet(conn *net.TCPConn) {
	data := make([]byte, 10*1024)
	defer func() {
		clear(data)
		_ = conn.Close()
		j.stop()
	}()
	for {
		select {
		case <-j.stopChan:
			return
		default:
			if n, err := conn.Read(data); err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
					return
				}
				slog.Warn("read jt1078 packet fail",
					slog.String("address", conn.RemoteAddr().String()),
					slog.Any("err", err))
				return
			} else if n > 0 {
				packs, err := j.getPackets(data[:n])
				if err != nil {
					slog.Error("parse jt1078 packet fail",
						slog.String("address", conn.RemoteAddr().String()),
						slog.String("data", fmt.Sprintf("%x", data[:n])),
						slog.Any("err", err))
					return
				}
				if len(packs) > 0 {
					select {
					case <-j.stopChan:
						return
					default:
						j.packetChan <- packs
					}
				}
			}
		}
	}
}

func (j *jt1078Server) getPackets(data []byte) ([]*jt1078.Packet, error) {
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

func (j *jt1078Server) writeGB28181Packet(conn net.Conn) {
	defer func() {
		_ = conn.Close()
		j.stop()
	}()
	for {
		select {
		case <-j.stopChan:
			return
		case packs := <-j.packetChan:
			// 转gb28181的ps包发送
			for _, pack := range packs {
				rtps := j.jt1078ToGB28181(pack)
				for _, data := range rtps {
					if len(data) > 0 {
						if _, err := conn.Write(data); err != nil {
							return
						}
					}
				}
			}
		}
	}
}

func (j *jt1078Server) jt1078ToGB28181(pack *jt1078.Packet) [][]byte {
	streamID := streamIDAudio
	if pack.Flag.PT == jt1078.PTH264 || pack.Flag.PT == jt1078.PTH265 {
		streamID = streamIDVideo
	}
	var (
		pts  = uint32(0)
		data = make([]byte, 0, 1460)
	)

	// 如果jt1078包的时间不准确 就使用本地时间
	if cur := time.Now().UnixMilli(); uint64(cur)-pack.Timestamp > 10*1000 {
		pts = uint32(cur)
	} else {
		pts = uint32(pack.Timestamp)
	}

	// 第一个包 I帧 psh + sys + psm + pes + h.264
	// 第一个包 P帧或者B帧 psh + pes + h.264
	// 音频 psh + pes + g711
	// https://blog.csdn.net/fanyun_01/article/details/120537670
	// https://ocw.unican.es/pluginfile.php/2825/course/section/2777/iso13818-1.pdf
	psh := NewProgramStreamPackHeader(pts)
	data = append(data, psh.Encode()...)
	if pack.DataType == jt1078.DataTypeI {
		sys := NewSystemHeader(j.hasAudio)
		data = append(data, sys.Encode()...)
		// psm可以只发一次
		psm := NewProgramStreamMap(j.createStreamMap()...)
		data = append(data, psm.Encode()...)
	}
	pes := NewPESPacket(streamID, uint16(len(pack.Body)), pts)
	data = append(data, pes.Encode()...)

	var (
		offset = 0
		end    = len(pack.Body)
		result = make([][]byte, 0, 1)
	)
	for end > 0 {
		chunkSize := 1350
		//chunkSize = len(pack.Body)
		if offset+chunkSize >= len(pack.Body) {
			chunkSize = len(pack.Body) - offset
		}
		data = append(data, pack.Body[offset:offset+chunkSize]...)

		offset += chunkSize
		end -= chunkSize
		result = append(result, createRTPPacket(pack.Flag.PT, data, func(packet *rtp.Packet) {
			packet.SSRC = j.ssrc32
			packet.Timestamp = pts
			packet.SequenceNumber = j.seq
			packet.Marker = end == 0
			j.seq++
		}))
		data = make([]byte, 0, 1460)
	}
	return result
}

func (j *jt1078Server) createStreamMap() []streamMap {
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
