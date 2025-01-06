package main

import (
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"github.com/q191201771/lal/pkg/base"
	"github.com/q191201771/lal/pkg/logic"
	"log/slog"
	"net"
	"sync"
)

type Lal1078 struct {
	lalServer logic.ILalServer
	addr      string
	ip        string
}

func newLal1078(addr string, ip string, filePath string) *Lal1078 {
	if filePath == "" {
		filePath = "./conf/lalserver.conf.json"
	}
	if addr == "" {
		addr = "0.0.0.0:1078"
	}
	lalServer := logic.NewLalServer(func(option *logic.Option) {
		option.ConfFilename = filePath
	})
	return &Lal1078{
		lalServer: lalServer,
		addr:      addr,
		ip:        ip,
	}
}

func (j *Lal1078) run() {
	go func() {
		if err := j.lalServer.RunLoop(); err != nil {
			panic(err)
		}
	}()

	listener, err := net.Listen("tcp", j.addr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go j.handle(conn)
	}
}

func (j *Lal1078) handle(conn net.Conn) {
	historyData := make([]byte, 0, 10*1024)
	data := make([]byte, 1024)
	var (
		once sync.Once
		ch   chan<- *jt1078.Packet
	)
	defer func() {
		_ = conn.Close()
		clear(historyData)
		clear(data)
	}()
	for {
		if n, err := conn.Read(data); err != nil {
			if ch != nil {
				close(ch)
			}
			return
		} else {
			historyData = append(historyData, data[:n]...)
		}
		packs := make([]*jt1078.Packet, 0)
		for {
			tmp := jt1078.NewPacket()
			if remainData, err := tmp.Decode(historyData); err != nil {
				if errors.Is(err, protocol.ErrUnqualifiedData) {
					return
				}
				break
			} else {
				historyData = remainData
				packs = append(packs, tmp)
				once.Do(func() {
					name := fmt.Sprintf("%s_%d", tmp.Sim, tmp.LogicChannel)
					fmt.Println("name is ", name)
					// http://49.234.235.7:8080/live/295696659617_1.flv
					fmt.Println("flv is", fmt.Sprintf("http://%s:8080/live/%s.flv", j.ip, name))
					streamCh := j.createStream(name)
					ch = streamCh
				})
				if len(remainData) == 0 {
					break
				}
			}
		}
		for _, v := range packs {
			select {
			case ch <- v:
			default:
				slog.Warn("channel is full",
					slog.String("body", fmt.Sprintf("%x", v.Body)))
			}
		}
	}
}

func (j *Lal1078) createStream(name string) chan<- *jt1078.Packet {
	session, err := j.lalServer.AddCustomizePubSession(name)
	if err != nil {
		panic(err)
	}
	session.WithOption(func(option *base.AvPacketStreamOption) {
		option.VideoFormat = base.AvPacketStreamVideoFormatAnnexb
	})
	ch := make(chan *jt1078.Packet, 100)
	go func(session logic.ICustomizePubSessionContext, ch <-chan *jt1078.Packet) {
		defer func() {
			j.lalServer.DelCustomizePubSession(session)
		}()
		var (
			once    sync.Once
			startTs int64
		)
		record := make(map[jt1078.DataType][]byte)
		for v := range ch {
			isComplete := false
			switch v.SubcontractType {
			case jt1078.SubcontractTypeAtomic:
				record[v.DataType] = v.Body
				isComplete = true
			case jt1078.SubcontractTypeFirst:
				record[v.DataType] = nil
				record[v.DataType] = v.Body
			case jt1078.SubcontractTypeLast:
				record[v.DataType] = append(record[v.DataType], v.Body...)
				isComplete = true
			case jt1078.SubcontractTypeMiddle:
				record[v.DataType] = append(record[v.DataType], v.Body...)
			default:
				panic("unknown SubcontractType")
			}
			if isComplete {
				data := record[v.DataType]
				once.Do(func() {
					startTs = int64(v.Timestamp)
				})
				tmp := base.AvPacket{
					PayloadType: base.AvPacketPtAvc,
					Timestamp:   int64(v.Timestamp) - startTs,
					Pts:         int64(v.Timestamp) - startTs,
					Payload:     data,
				}
				switch v.Flag.PT {
				case jt1078.PTG711A:
					tmp.PayloadType = base.AvPacketPtG711A
				case jt1078.PTG711U:
					tmp.PayloadType = base.AvPacketPtG711U
				case jt1078.PTH264:
				case jt1078.PTH265:
					tmp.PayloadType = base.AvPacketPtHevc
				default:
					slog.Warn("未知类型",
						slog.Any("pt", v.Flag.PT))
				}
				if err := session.FeedAvPacket(tmp); err != nil {
					slog.Warn("session.FeedAvPacket",
						slog.Any("err", err))
				}
			}
		}
	}(session, ch)
	return ch
}
