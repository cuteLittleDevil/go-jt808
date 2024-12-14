package adapter

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"
)

type group struct {
	conn         *net.TCPConn
	timeoutRetry time.Duration
	clients      sync.Map
	writeMsgChan chan []byte
	stopChan     chan struct{}
	stopOnce     sync.Once
}

func newGroup(conn *net.TCPConn, timeoutRetry time.Duration, terminals []Terminal) *group {
	g := &group{
		conn:         conn,
		timeoutRetry: timeoutRetry,
		clients:      sync.Map{},
		writeMsgChan: make(chan []byte, 10),
		stopChan:     make(chan struct{}),
		stopOnce:     sync.Once{},
	}
	for _, terminal := range terminals {
		if c, err := newClient(terminal, g.stopChan, g.sendData); err == nil {
			g.clients.Store(&terminal, c)
		} else {
			slog.Error("init client error",
				slog.String("addr", terminal.TargetAddr),
				slog.Any("err", err))
			g.clients.Store(&terminal, nil)
		}
	}
	return g
}

func (g *group) run() {
	go g.reader()
	go g.write()
}

func (g *group) reader() {
	var (
		// 消息体长度最大为 10bit 也就是 1023 的字节
		curData = make([]byte, 1023)
	)
	defer func() {
		g.stop()
	}()

	for {
		if n, err := g.conn.Read(curData); err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				slog.Debug("client close",
					slog.Any("err", err))
				return
			}
			slog.Error("read data",
				slog.Any("err", err))
			return
		} else if n > 0 {
			// 把数据发送下去 如果某个终端挂了 尝试新建终端
			var wg sync.WaitGroup
			g.clients.Range(func(key, value any) bool {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if v, ok := value.(*client); ok {
						if ok := v.sendData(curData[:n]); !ok {
							slog.Warn("send data",
								slog.Any("addr", v.terminal.TargetAddr),
								slog.String("data", fmt.Sprintf("%x", curData[:n])),
								slog.Any("err", err))
							g.clients.Delete(key) // 服务断开的
							go g.addClient(v.terminal)
						}
					} else if t, ok := key.(*Terminal); ok { // 初始化失败的
						g.clients.Delete(t)
						go g.addClient(*t)
						return
					}
				}()
				return true
			})
			wg.Wait()
		}
	}
}

func (g *group) sendData(data []byte) {
	select {
	case <-g.stopChan:
		return
	default:
		g.writeMsgChan <- data
	}
}

func (g *group) write() {
	for {
		select {
		case <-g.stopChan:
			return
		case data, ok := <-g.writeMsgChan:
			if ok {
				if _, err := g.conn.Write(data); err != nil {
					slog.Warn("write",
						slog.String("data", fmt.Sprintf("%x", data)),
						slog.Any("err", err))
				}
			}
		}
	}
}

func (g *group) stop() {
	g.stopOnce.Do(func() {
		close(g.stopChan)
		_ = g.conn.Close()
		g.clients.Clear()
		close(g.writeMsgChan)
	})
}

func (g *group) addClient(terminal Terminal) {
	for {
		select {
		case <-g.stopChan:
			return
		default:
			// 隔一段时间试一试 重新和808服务建立连接
			if c, err := newClient(terminal, g.stopChan, g.sendData); err == nil {
				g.clients.Store(terminal.TargetAddr, c)
				slog.Info("rejoin",
					slog.String("addr", terminal.TargetAddr))
				return
			} else {
				slog.Warn("new client error",
					slog.String("addr", terminal.TargetAddr),
					slog.Any("err", err))
				time.Sleep(g.timeoutRetry)
			}
		}
	}
}
