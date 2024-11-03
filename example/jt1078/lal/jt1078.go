package main

import (
	"fmt"
	"github.com/q191201771/lal/pkg/base"
	"github.com/q191201771/lal/pkg/logic"
	"log/slog"
	"net"
	"sync"
)

type jt1078 struct {
	lalServer logic.ILalServer
	addr      string
}

func newJt1078(addr string, filePath string) *jt1078 {
	if filePath == "" {
		filePath = "./conf/lalserver.conf.json"
	}
	if addr == "" {
		addr = "0.0.0.0:1078"
	}
	lalServer := logic.NewLalServer(func(option *logic.Option) {
		option.ConfFilename = filePath
	})
	return &jt1078{
		lalServer: lalServer,
		addr:      addr,
	}
}

func (j *jt1078) run() {
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

func (j *jt1078) handle(conn net.Conn) {
	historyData := make([]byte, 0, 10*1024)
	data := make([]byte, 1024)
	var (
		once sync.Once
		ch   chan<- *Packet
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
		packs := make([]*Packet, 0)
		for {
			tmp := NewPacket()
			if remainingData, err := tmp.Parse(historyData); err != nil {
				break
			} else {
				historyData = remainingData
				packs = append(packs, tmp)
				once.Do(func() {
					name := fmt.Sprintf("%s_%d", tmp.Sim, tmp.LogicChannel)
					fmt.Println("name is ", name)
					// http://49.234.235.7:8080/live/295696659617_1.flv
					fmt.Println("flv is", fmt.Sprintf("http://%s:8080/live/%s.flv", _ip, name))
					streamCh := j.createStream(name)
					ch = streamCh
				})
			}
		}
		for _, v := range packs {
			select {
			case ch <- v:
			default:
				slog.Warn("channel is full",
					slog.String("data", fmt.Sprintf("%x", v.Data)))
			}
		}
	}
}

func (j *jt1078) createStream(name string) chan<- *Packet {
	session, err := j.lalServer.AddCustomizePubSession(name)
	if err != nil {
		panic(err)
	}
	session.WithOption(func(option *base.AvPacketStreamOption) {
		option.VideoFormat = base.AvPacketStreamVideoFormatAnnexb
	})
	ch := make(chan *Packet, 100)
	go func(session logic.ICustomizePubSessionContext, ch <-chan *Packet) {
		defer func() {
			j.lalServer.DelCustomizePubSession(session)
		}()
		var (
			once    sync.Once
			startTs int64
		)
		record := make(map[DataType][]byte)
		sum := int64(1) // 测试是循环播放的 所以手动增加
		for v := range ch {
			isComplete := false
			switch v.SubcontractType {
			case SubcontractTypeAtomic:
				record[v.DataType] = v.Data
				isComplete = true
			case SubcontractTypeFirst:
				record[v.DataType] = nil
				record[v.DataType] = v.Data
			case SubcontractTypeLast:
				record[v.DataType] = append(record[v.DataType], v.Data...)
				isComplete = true
			case SubcontractTypeMiddle:
				record[v.DataType] = append(record[v.DataType], v.Data...)
			default:
				panic("unknown SubcontractType")
			}
			if isComplete {
				data := record[v.DataType]
				once.Do(func() {
					startTs = v.Timestamp
				})
				tmp := base.AvPacket{
					PayloadType: base.AvPacketPtAvc,
					Timestamp:   v.Timestamp - startTs,
					Pts:         v.Timestamp - startTs,
					Payload:     data,
				}
				tmp.Timestamp = sum
				tmp.Pts = sum
				sum += 50
				switch v.Flag.PT {
				case PTG711A:
					tmp.PayloadType = base.AvPacketPtG711A
				case PTG711U:
					tmp.PayloadType = base.AvPacketPtG711U
				case PTH264:
				case PTH265:
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
