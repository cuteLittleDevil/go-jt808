package adapter

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"net"
	"sync"
)

type client struct {
	replyFunc     func(data []byte)
	terminal      Terminal
	groupStopChan <-chan struct{}

	conn     net.Conn
	readChan chan []byte
	stopChan chan struct{}
	stopOnce sync.Once
}

func newClient(terminal Terminal, groupStopChan <-chan struct{}, replyFunc func(data []byte)) (*client, error) {
	conn, err := net.Dial("tcp", terminal.TargetAddr)
	if err != nil {
		return nil, err
	}
	tmp := &client{
		terminal:      terminal,
		groupStopChan: groupStopChan,
		replyFunc:     replyFunc,
		conn:          conn,
		readChan:      make(chan []byte, 10),
		stopChan:      make(chan struct{}),
		stopOnce:      sync.Once{},
	}
	go tmp.run()
	return tmp, nil
}

func (c *client) run() {
	go c.reader()
	defer func() {
		c.stop()
	}()
	for {
		select {
		case <-c.groupStopChan:
			return
		case <-c.stopChan:
			return
		case data, ok := <-c.readChan:
			if ok {
				if _, err := c.conn.Write(data); err != nil {
					slog.Warn("write",
						slog.String("data", fmt.Sprintf("%x", data)),
						slog.Any("err", err))
				}
			}
		}
	}
}

func (c *client) sendData(data []byte) bool {
	select {
	case <-c.stopChan:
		return false
	case <-c.groupStopChan:
		return false
	default:
		c.readChan <- data
		return true
	}
}

func (c *client) reader() {
	var (
		// 消息体长度最大为 10bit 也就是 1023 的字节
		curData = make([]byte, 1023)
		pack    = newPackageParse()
	)
	defer func() {
		c.stop()
		clear(curData)
		pack.clear()
	}()

	for {
		select {
		case <-c.stopChan:
			return
		default:
			if n, err := c.conn.Read(curData); err != nil {
				return
			} else if n > 0 {
				msgs, _ := pack.unpack(curData[:n])
				for _, msg := range msgs {
					command := consts.JT808CommandType(msg.JTMessage.Header.ID)
					if c.terminal.allowReply(command) {
						c.replyFunc(msg.originalData)
					}
				}
			}
		}
	}
}

func (c *client) stop() {
	c.stopOnce.Do(func() {
		close(c.stopChan)
		_ = c.conn.Close()
		close(c.readChan)
	})
}
