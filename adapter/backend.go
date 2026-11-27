package adapter

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

// readBufSize 单次读取缓冲区大小（JT808 消息体最大 1023B，算上消息头、校验码、首尾 0x7e 及 0x7d 转义膨胀后上限约 2048B）。
const readBufSize = 2048

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
	opts        *Options
	sessionStop <-chan struct{}
	// onReply 后端下行且允许回写时，交给 session 写回真实终端。
	onReply func(data []byte)
	// onDead 连接因读/写/拆包失败关闭时回调一次（session 停止时不再触发重连）。
	onDead func()

	conn       net.Conn
	writeQueue chan []byte // 待发往后端的上行数据（由 session 扇出入队）
	stopCh     chan struct{}
	stopOnce   sync.Once
}

func newBackendClient(
	terminal Terminal,
	opts *Options,
	sessionStop <-chan struct{},
	onReply func(data []byte),
	onDead func(),
) (*backendClient, error) {
	if opts == nil {
		opts = newOptions(nil)
	}
	conn, err := net.DialTimeout("tcp", terminal.TargetAddr, opts.DialTimeout)
	if err != nil {
		return nil, err
	}
	c := &backendClient{
		terminal:    terminal,
		opts:        opts,
		sessionStop: sessionStop,
		onReply:     onReply,
		onDead:      onDead,
		conn:        conn,
		writeQueue:  make(chan []byte, 10),
		stopCh:      make(chan struct{}),
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
			if c.opts != nil && c.opts.WriteTimeout > 0 {
				_ = c.conn.SetWriteDeadline(time.Now().Add(c.opts.WriteTimeout))
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
	readBuf := make([]byte, readBufSize)
	pack := newPackageParse()
	defer func() {
		c.stop()
		clear(readBuf)
		pack.clear()
	}()

	maxBufSize := defaultMaxFrameBufferSize
	if c.opts != nil {
		maxBufSize = c.opts.MaxFrameBufferSize
	}

	for {
		select {
		case <-c.stopCh:
			return
		default:
			n, err := c.conn.Read(readBuf)
			if n > 0 {
				msgs, unpackErr := pack.unpack(readBuf[:n])
				for _, msg := range msgs {
					command := consts.JT808CommandType(msg.JTMessage.Header.ID)
					if c.terminal.allowReply(command) {
						// 必须拷贝后再异步入队：msg.originalData 直接引用 readBuf 或粘包缓冲，
						// 下一轮 Read 会覆盖尚未写回真实终端的数据。
						c.onReply(bytes.Clone(msg.originalData))
					}
				}
				if unpackErr != nil {
					slog.Warn("unpack downlink data",
						slog.String("addr", c.terminal.TargetAddr),
						slog.String("effective data", fmt.Sprintf("%x", readBuf[:n])),
						slog.Any("err", unpackErr))
					return
				}
				// 粘包缓冲区未组成完整帧的数据过多（如 7e... 迟迟未闭合或非 0x7e 开头）说明数据异常，断开连接防止内存无限增长
				if pending := pack.frameReader.Pending(); maxBufSize > 0 && pending >= maxBufSize {
					slog.Warn("frame buffer overflow",
						slog.String("addr", c.terminal.TargetAddr),
						slog.Int("pending", pending),
						slog.Int("max", maxBufSize))
					return
				}
			}
			if err != nil {
				return
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
