package stream

import (
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"io"
	"log/slog"
	"net"
	"sync"
)

type jt1078Server struct {
	stopOnce        sync.Once
	stopChan        chan struct{}
	listen          *net.TCPListener
	packetChan      chan []*jt1078.Packet
	packHandle      *packageParse
	jt1078Port      int
	gb28181IP       string
	gb28181Port     int
	jt1078ToGB28181 command.JT1078ToGB28181er
}

func newJt1078Server(info *command.InviteInfo, jt1078ToGB28181 command.JT1078ToGB28181er) *jt1078Server {
	return &jt1078Server{
		stopOnce:        sync.Once{},
		stopChan:        make(chan struct{}),
		listen:          nil,
		packetChan:      make(chan []*jt1078.Packet, 100),
		packHandle:      newPackageParse(),
		jt1078Port:      info.JT1078Info.Port,
		gb28181IP:       info.IP,
		gb28181Port:     info.Port,
		jt1078ToGB28181: jt1078ToGB28181,
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
		}
	}
}

func (j *jt1078Server) stop(msg string) {
	j.stopOnce.Do(func() {
		close(j.stopChan)
		if j.listen != nil {
			_ = j.listen.Close()
		}
		j.packHandle.clear()
		if j.jt1078ToGB28181 != nil {
			j.jt1078ToGB28181.OnBye(msg)
		}
	})
}

func (j *jt1078Server) readJt1078Packet(conn *net.TCPConn) {
	data := make([]byte, 10*1024)
	defer func() {
		clear(data)
		_ = conn.Close()
		j.stop("")
		close(j.packetChan)
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
				if packs, err := j.getPackets(data[:n]); err != nil {
					return
				} else if len(packs) > 0 {
					j.packetChan <- packs
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
		j.stop("")
	}()
	for {
		select {
		case <-j.stopChan:
			return
		case packs := <-j.packetChan:
			// 转gb28181的ps包发送
			for _, pack := range packs {
				rtps := j.jt1078ToGB28181.ConvertToGB28181(pack)
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
