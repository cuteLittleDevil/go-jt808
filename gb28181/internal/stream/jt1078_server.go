package stream

import (
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"io"
	"log/slog"
	"math/rand/v2"
	"net"
	"strconv"
	"sync"
	"time"
)

type jt1078Server struct {
	stopOnce   sync.Once
	stopChan   chan struct{}
	listen     *net.TCPListener
	packetChan chan []*jt1078.Packet
	ssrc32     uint32
	seq        uint16
	packHandle *packageParse
}

func newJt1078Server(ssrc string) *jt1078Server {
	ssrc32, err := strconv.ParseUint(ssrc, 10, 32)
	if err != nil {
		ssrc32 = rand.Uint64()
	}
	return &jt1078Server{
		stopOnce:   sync.Once{},
		stopChan:   make(chan struct{}),
		listen:     nil,
		packetChan: make(chan []*jt1078.Packet, 100),
		ssrc32:     uint32(ssrc32),
		packHandle: newPackageParse(),
	}
}

func (j *jt1078Server) run(jt808Port int, ip string, gb28181Port int) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", jt808Port))
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
		gb28181Conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, gb28181Port))
		if err == nil {
			go j.writeGB28181Packet(gb28181Conn)
			slog.Info("connect gb28181 success",
				slog.Int("jt1078", jt808Port),
				slog.String("ip", ip),
				slog.Int("gb28181", gb28181Port))
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
				data := j.jt1078ToGB28181(pack)
				if len(data) > 0 {
					if _, err := conn.Write(data); err != nil {
						return
					}
				}
			}
		}
	}
}

func (j *jt1078Server) jt1078ToGB28181(pack *jt1078.Packet) []byte {
	return nil
}
