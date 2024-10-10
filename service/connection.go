package service

import (
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"io"
	"log/slog"
	"net"
	"sync"
)

type connection struct {
	conn        *net.TCPConn
	handles     map[consts.JT808CommandType]Handler
	stopOnce    sync.Once
	stopChan    chan struct{}
	msgBuffChan chan *Message
}

func newConnection(conn *net.TCPConn, handles map[consts.JT808CommandType]Handler) *connection {
	return &connection{
		conn:        conn,
		handles:     handles,
		stopOnce:    sync.Once{},
		stopChan:    make(chan struct{}),
		msgBuffChan: make(chan *Message, 10),
	}
}

func (c *connection) Start() {
	go c.reader()
	go c.write()
}

func (c *connection) reader() {
	// 消息体长度最大为 10bit 也就是 1023 的字节
	curData := make([]byte, 1023)
	pack := newPackageParse()

	defer func() {
		c.stop()
		curData = nil
		pack.clear()
	}()

	for {
		select {
		case <-c.stopChan:
			return
		default:
			if n, err := c.conn.Read(curData); err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
					slog.Debug("connection close",
						slog.Any("err", err))
					return
				}
				slog.Error("read data",
					slog.Any("err", err))
				return
			} else if n > 0 {
				effectiveData := curData[:n]
				msgs, err := pack.parse(effectiveData)
				if err != nil {
					slog.Error("parse data",
						slog.String("effective data", fmt.Sprintf("%x", effectiveData)),
						slog.Any("err", err))
					return
				}
				for _, msg := range msgs {
					key := consts.JT808CommandType(msg.JTMessage.Header.ID)
					if handler, ok := c.handles[key]; ok {
						msg.Handler = handler
						msg.OnReadExecutionEvent(msg)
					} else {
						slog.Warn("key not found",
							slog.Int("id", int(msg.JTMessage.Header.ID)),
							slog.String("remark", key.String()))
						continue
					}
					select {
					case c.msgBuffChan <- msg:
					default:
						slog.Warn("msg buff full",
							slog.String("original data", fmt.Sprintf("%x", msg.OriginalData)))
					}
				}

			}
		}
	}
}

func (c *connection) write() {
	// 到了math.MaxUint16后+1重新变成0
	num := uint16(0)
	for {
		select {
		case <-c.stopChan:
			return
		case msg, ok := <-c.msgBuffChan:
			if ok {
				header := msg.JTMessage.Header
				header.ReplyID = uint16(msg.ReplyProtocol())
				header.PlatformSerialNumber = num
				num++
				if has := msg.HasReply(); !has {
					continue
				}
				body, err := msg.ReplyBody(msg.JTMessage)
				if err != nil {
					slog.Warn("reply body fail",
						slog.String("original data", fmt.Sprintf("%x", msg.OriginalData)),
						slog.Any("err", err))
					continue
				}
				data := header.Encode(body)
				_, err = c.conn.Write(data)
				if err != nil {
					slog.Warn("write fail",
						slog.String("data", fmt.Sprintf("%x", data)),
						slog.Any("err", err))
				}
				msg.ReplyData = data
				msg.WriteErr = err
				msg.OnWriteExecutionEvent(*msg)
			} else if msg != nil && len(msg.OriginalData) > 0 {
				slog.Warn("msgBuffChan is close",
					slog.String("original data", fmt.Sprintf("%x", msg.OriginalData)))
			}
		}
	}
}

func (c *connection) stop() {
	c.stopOnce.Do(func() {
		close(c.msgBuffChan)
		close(c.stopChan)
	})
}
