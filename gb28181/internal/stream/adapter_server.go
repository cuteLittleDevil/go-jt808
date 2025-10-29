package stream

import (
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"io"
	"log/slog"
	"net"
	"sync"
)

type adapterServer struct {
	stopOnce    sync.Once
	stopChan    chan struct{}
	listen      *net.TCPListener
	rtpChan     chan [][]byte
	streamPort  int
	gb28181IP   string
	gb28181Port int
	toGB28181er command.ToGB28181er
}

func newAdapterServer(info *command.InviteInfo, toGB28181er command.ToGB28181er) *adapterServer {
	if info.Adapter.Type == command.JT1078ToPS || info.Adapter.Type == command.JT1078ToPSFilterPacket {
		info.Adapter.Port = info.JT1078Info.Port
	}
	return &adapterServer{
		stopOnce:    sync.Once{},
		stopChan:    make(chan struct{}),
		listen:      nil,
		rtpChan:     make(chan [][]byte, 100),
		streamPort:  info.Adapter.Port,
		gb28181IP:   info.IP,
		gb28181Port: info.Port,
		toGB28181er: toGB28181er,
	}
}

func (j *adapterServer) run() {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", j.streamPort))
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
		slog.Info("jt1078 connect success",
			slog.Any("addr", conn.RemoteAddr()))
		go j.readPacket(conn)
		gb28181Address := fmt.Sprintf("%s:%d", j.gb28181IP, j.gb28181Port)
		if gb28181Conn, err := net.Dial("tcp", gb28181Address); err == nil {
			go j.writeGB28181Packet(gb28181Conn)
		}
	}
}

func (j *adapterServer) stop(msg string) {
	j.stopOnce.Do(func() {
		close(j.stopChan)
		if j.listen != nil {
			_ = j.listen.Close()
		}
		if j.toGB28181er != nil {
			j.toGB28181er.OnBye(msg)
		}
	})
}

func (j *adapterServer) readPacket(conn *net.TCPConn) {
	// 根据真实设备报文有3w+ byte调整
	data := make([]byte, 10*10*1024)
	defer func() {
		clear(data)
		_ = conn.Close()
		j.stop("")
		close(j.rtpChan)
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
				slog.Warn("read packet fail",
					slog.String("address", conn.RemoteAddr().String()),
					slog.Any("err", err))
				return
			} else if n > 0 {
				// 一开始认为处理够快 数据污染就污染了 用模拟流有时间间隔可以 实际流不行 #12
				effectData := make([]byte, n)
				copy(effectData, data[:n])
				if rtps, err := j.toGB28181er.ConvertToGB28181(effectData); err != nil {
					slog.Error("convert to gb28181 packet fail",
						slog.String("address", conn.RemoteAddr().String()),
						slog.String("data", fmt.Sprintf("%x", data)),
						slog.Any("err", err))
					return
				} else {
					j.rtpChan <- rtps
				}
			}
		}
	}
}

func (j *adapterServer) writeGB28181Packet(conn net.Conn) {
	defer func() {
		_ = conn.Close()
		j.stop("")
	}()
	for {
		select {
		case <-j.stopChan:
			return
		case rtps := <-j.rtpChan:
			for _, rtp := range rtps {
				if len(rtp) > 0 {
					if _, err := conn.Write(rtp); err != nil {
						return
					}
				}
			}
		}
	}
}
