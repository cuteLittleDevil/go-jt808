package adapter

import (
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

// maxBodyLen JT808 消息体最大长度（协议 10bit），读缓冲按此上限复用。
const maxBodyLen = 1023

// backend 是 session.backends 的 value。
//
// key 固定为 terminal.TargetAddr（string），初始化与重连必须一致。
// 历史上曾用 *Terminal 作 key、重连用 string，sync.Map 按类型判等会导致孤儿条目。
// Dial 失败时 client 为 nil，仍需要完整 Terminal 才能重连，故 value 不能只有 *backendClient。
type backend struct {
	terminal Terminal
	// client 与后端 808 的连接；未连上时为 nil，由 reconnectBackend 重建。
	client *backendClient
}

// backendClient 表示连向某一个后端 JT808 服务的 TCP 客户端。
type backendClient struct {
	terminal    Terminal
	sessionStop <-chan struct{}
	// onReply 后端下行且允许回写时，交给 session 写回真实终端。
	onReply func(data []byte)
	// onDead 连接因读/写失败关闭时回调一次（session 停止时不再触发重连）。
	onDead func()

	conn       net.Conn
	writeQueue chan []byte // 待发往后端的上行数据（由 session 扇出入队）
	stopCh     chan struct{}
	stopOnce   sync.Once
}

func newBackendClient(
	terminal Terminal,
	sessionStop <-chan struct{},
	onReply func(data []byte),
	onDead func(),
) (*backendClient, error) {
	conn, err := net.Dial("tcp", terminal.TargetAddr)
	if err != nil {
		return nil, err
	}
	c := &backendClient{
		terminal:    terminal,
		sessionStop: sessionStop,
		onReply:     onReply,
		onDead:      onDead,
		conn:        conn,
		writeQueue:  make(chan []byte, 10),
		stopCh:      make(chan struct{}),
		stopOnce:    sync.Once{},
	}
	go c.run()
	return c, nil
}

func (c *backendClient) run() {
	go c.readLoop()
	defer c.stop()
	for {
		select {
		case <-c.sessionStop:
			return
		case <-c.stopCh:
			return
		case data, ok := <-c.writeQueue:
			if !ok {
				return
			}
			if _, err := c.conn.Write(data); err != nil {
				slog.Warn("write",
					slog.String("data", fmt.Sprintf("%x", data)),
					slog.Any("err", err))
				// 写失败视为连接死亡，退出后 stop → onDead 触发重连
				return
			}
		}
	}
}

// enqueueUplink 将待发往后端的上行数据入队。
//
// 此处不拷贝 data。调用方必须保证 data 在异步 Write 完成前不被覆盖
// （session.readLoop 在扇出前已 copy 出独立 payload）。
// 返回 false 表示本客户端或 session 已停止，调用方应摘除并重连。
func (c *backendClient) enqueueUplink(data []byte) bool {
	select {
	case <-c.stopCh:
		return false
	case <-c.sessionStop:
		return false
	case c.writeQueue <- data:
		return true
	}
}

func (c *backendClient) readLoop() {
	readBuf := make([]byte, maxBodyLen)
	pack := newPackageParse()
	defer func() {
		c.stop()
		clear(readBuf)
		pack.clear()
	}()

	for {
		select {
		case <-c.stopCh:
			return
		default:
			n, err := c.conn.Read(readBuf)
			if err != nil {
				return
			}
			if n <= 0 {
				continue
			}
			msgs, _ := pack.unpack(readBuf[:n])
			for _, msg := range msgs {
				command := consts.JT808CommandType(msg.JTMessage.Header.ID)
				if c.terminal.allowReply(command) {
					c.onReply(msg.originalData)
				}
			}
		}
	}
}

func (c *backendClient) stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
		_ = c.conn.Close()
		// session 整体退出时不再回调，避免对已关闭 session 重连
		select {
		case <-c.sessionStop:
			return
		default:
		}
		if c.onDead != nil {
			c.onDead()
		}
	})
}
