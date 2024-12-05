package pkg

import (
	"errors"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	_ "github.com/go-resty/resty/v2"
	"io"
	"log/slog"
	"m7s.live/v5"
	"m7s.live/v5/pkg"
	"m7s.live/v5/pkg/codec"
	"net"
	"sync"
	"time"
)

type connection struct {
	conn net.Conn
	*slog.Logger
	stopChan     chan struct{}
	stopOnce     sync.Once
	publisher    *m7s.Publisher
	onJoinEvent  func(c *connection, pack *jt1078.Packet) error
	onLeaveEvent func()
	ptsFunc      func(pack *jt1078.Packet) time.Duration
}

func newConnection(c net.Conn, log *slog.Logger, ptsFunc func(pack *jt1078.Packet) time.Duration) *connection {
	return &connection{
		Logger:   log,
		conn:     c,
		stopChan: make(chan struct{}),
		ptsFunc:  ptsFunc,
	}
}

func (c *connection) run(audioChan <-chan []byte) error {
	var (
		data      = make([]byte, 10*1024)
		packParse = newPackageParse()
		once      sync.Once
		onJoinErr error
	)
	defer func() {
		packParse.clear()
		clear(data)
		c.stop()
	}()
	if audioChan != nil {
		go c.write(audioChan)
	}
	for {
		if n, err := c.conn.Read(data); err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return nil
			}
			return err
		} else if n > 0 {
			for pack, err := range packParse.parse(data[:n]) {
				if err == nil {
					once.Do(func() {
						onJoinErr = c.onJoinEvent(c, pack)
					})
					if onJoinErr == nil {
						c.handle(pack)
					}
				} else if errors.Is(err, jt1078.ErrBodyLength2Short) || errors.Is(err, jt1078.ErrHeaderLength2Short) {
					// 数据长度不够的 忽略
				} else {
					return err
				}
			}
			if onJoinErr != nil {
				return onJoinErr
			}
		}
	}
}

func (c *connection) write(audioChan <-chan []byte) {
	for {
		select {
		case <-c.stopChan:
			return
		case data := <-audioChan:
			if _, err := c.conn.Write(data); err != nil {
				return
			}
		}
	}
}

func (c *connection) stop() {
	c.stopOnce.Do(func() {
		close(c.stopChan)
		c.onLeaveEvent()
	})
}

func (c *connection) handle(packet *jt1078.Packet) {
	pts := c.ptsFunc(packet)
	data := packet.Body
	var (
		result    pkg.IAVFrame
		writeFunc func(pkg.IAVFrame) error
	)
	switch pt := packet.Flag.PT; pt {
	case jt1078.PTAAC, jt1078.PTG711A, jt1078.PTG711U:
		result = c.parseAudioPacket(pt, pts, data)
		writeFunc = c.publisher.WriteAudio
	case jt1078.PTH264, jt1078.PTH265:
		result = c.parseVideoPacket(pt, pts, data)
		writeFunc = c.publisher.WriteVideo
	default:
		c.Warn("unknown pt",
			slog.String("pt", pt.String()))
		return
	}
	if result != nil && writeFunc != nil {
		if err := writeFunc(result); err != nil {
			c.Warn("publish fail",
				slog.Any("packet", packet.String()),
				slog.String("err", err.Error()))
			return
		}
	}
}

func (c *connection) parseAudioPacket(pt jt1078.PTType, pts time.Duration, data []byte) pkg.IAVFrame {
	var result pkg.IAVFrame
	switch pt {
	case jt1078.PTAAC:
		var adts = &pkg.ADTS{
			DTS: pts,
		}
		adts.Memory.AppendOne(data)
		result = adts
	case jt1078.PTG711A:
		rawAudio := &pkg.RawAudio{
			Timestamp: pts,
			FourCC:    codec.FourCC_ALAW,
		}
		rawAudio.Memory.AppendOne(data)
		result = rawAudio
	case jt1078.PTG711U:
		rawAudio := &pkg.RawAudio{
			Timestamp: pts,
			FourCC:    codec.FourCC_ULAW,
		}
		rawAudio.Memory.AppendOne(data)
		result = rawAudio
	}
	return result
}

func (c *connection) parseVideoPacket(pt jt1078.PTType, pts time.Duration, data []byte) pkg.IAVFrame {
	result := &pkg.AnnexB{
		PTS: pts,
		DTS: pts, // 没有b帧的情况是一样的
	}
	if pt == jt1078.PTH265 {
		result.Hevc = true
	}
	result.Memory.AppendOne(data)
	return result
}
