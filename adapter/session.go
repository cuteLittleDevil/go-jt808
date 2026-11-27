package adapter

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"
)

// session 管理「一个真实终端 TCP」到「多个后端 808」的扇出与回写。
//
//	真实终端 ──Read──► readLoop 扇出 ──► 多个 backendClient
//	真实终端 ◄─Write── writeLoop ◄── 后端允许回复的下行
type session struct {
	conn      net.Conn
	opts      *Options
	terminals []Terminal
	// backends：key=TargetAddr(string)，value=*backend。写入必须走 putBackend。
	backends sync.Map
	// reconnecting：key=TargetAddr，标记该地址是否已有重连协程在跑（含重试等待期间）。
	// 用 LoadOrStore 保证同一地址全局最多一个 reconnectBackend，不依赖「从 map 删掉占位」这种隐式约定。
	reconnecting sync.Map
	replyQueue   chan []byte // 写回真实终端的下行队列
	stopCh       chan struct{}
	stopOnce     sync.Once
}

func newSession(conn net.Conn, opts *Options, terminals []Terminal) *session {
	if opts == nil {
		opts = newOptions(nil)
	}
	s := &session{
		conn:       conn,
		opts:       opts,
		terminals:  terminals,
		replyQueue: make(chan []byte, 10),
		stopCh:     make(chan struct{}),
	}
	// 先写入 nil 占位避免阻塞 Accept，首次拨号在 run 协程内执行。
	for _, terminal := range terminals {
		s.putBackend(terminal, nil)
	}
	return s
}

// dialBackend 拨号并登记；失败则占位并单飞重连。
func (s *session) dialBackend(terminal Terminal) {
	onDead := func() {
		// 连接中途死亡：清 client 并重连（session 已 stop 时 stop() 内不会回调）
		s.putBackend(terminal, nil)
		s.startReconnect(terminal)
	}
	c, err := newBackendClient(terminal, s.opts, s.stopCh, s.enqueueReply, onDead)
	if err != nil {
		slog.Error("init backend error",
			slog.String("addr", terminal.TargetAddr),
			slog.Any("err", err))
		s.putBackend(terminal, nil)
		s.startReconnect(terminal)
		return
	}
	s.putBackend(terminal, c)
}

// putBackend 统一写入 backends，保证 key 始终为 TargetAddr。
func (s *session) putBackend(terminal Terminal, c *backendClient) {
	s.backends.Store(terminal.TargetAddr, &backend{
		terminal: terminal,
		client:   c,
	})
}

func (s *session) run() {
	for _, terminal := range s.terminals {
		select {
		case <-s.stopCh:
			return
		default:
			s.dialBackend(terminal)
		}
	}
	go s.readLoop()
	go s.writeLoop()
}

func (s *session) readLoop() {
	// readBuf 仅供 Read 复用，不得直接交给 backendClient。
	readBuf := make([]byte, readBufSize)
	defer s.stop()

	for {
		if s.opts.IdleTimeout > 0 {
			_ = s.conn.SetReadDeadline(time.Now().Add(s.opts.IdleTimeout))
		}
		n, err := s.conn.Read(readBuf)
		if n > 0 {
			// 必须先拷贝再扇出。
			//
			// enqueueUplink 只把 slice 头放入 writeQueue，真正 Write 在 backendClient 异步执行。
			// wg.Wait 只保证「入队完成」，不保证「写完」。
			// 若直接传 readBuf[:n]：下一轮 Read 会覆盖尚未写出的数据（data race）。
			// 同一轮扇出共享一份 payload 即可（只读）；每轮 Read 必须是新的 payload。
			payload := make([]byte, n)
			copy(payload, readBuf[:n])

			var wg sync.WaitGroup
			s.backends.Range(func(_, value any) bool {
				b := value.(*backend)
				wg.Add(1)
				go func(b *backend) {
					defer wg.Done()
					if b.client == nil {
						// 仍无可用连接：只触发「至多一个」重连协程，占位条目保留。
						// 新数据反复到达时 startReconnect 会直接返回，不会堆协程。
						s.startReconnect(b.terminal)
						return
					}
					if ok := b.client.enqueueUplink(payload); !ok {
						slog.Warn("forward uplink",
							slog.String("addr", b.terminal.TargetAddr),
							slog.String("data", fmt.Sprintf("%x", payload)))
						// 清掉失效 client，保留 Terminal 配置占位，再单飞重连
						s.putBackend(b.terminal, nil)
						s.startReconnect(b.terminal)
					}
				}(b)
				return true
			})
			// 等待本轮全部入队（非等待全部 Write）
			wg.Wait()
		}
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				slog.Info("terminal idle timeout",
					slog.Duration("idleTimeout", s.opts.IdleTimeout),
					slog.Any("err", err))
				return
			}
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
				slog.Debug("terminal close",
					slog.Any("err", err))
				return
			}
			slog.Error("read data",
				slog.Any("err", err))
			return
		}
	}
}

// enqueueReply 将后端允许回写的数据交给 writeLoop 写回真实终端。
func (s *session) enqueueReply(data []byte) {
	select {
	case <-s.stopCh:
		return
	case s.replyQueue <- data:
	}
}

func (s *session) writeLoop() {
	for {
		select {
		case <-s.stopCh:
			return
		case data := <-s.replyQueue:
			if s.opts.WriteTimeout > 0 {
				_ = s.conn.SetWriteDeadline(time.Now().Add(s.opts.WriteTimeout))
			}
			if _, err := s.conn.Write(data); err != nil {
				slog.Warn("write",
					slog.String("data", fmt.Sprintf("%x", data)),
					slog.Any("err", err))
				s.stop()
				return
			}
		}
	}
}

func (s *session) stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		_ = s.conn.Close()
		s.backends.Range(func(_, value any) bool {
			if b, ok := value.(*backend); ok && b != nil && b.client != nil {
				b.client.stop()
			}
			return true
		})
		s.backends.Clear()
		s.reconnecting.Clear()
	})
}

// startReconnect 确保同一 TargetAddr 最多只有一个 reconnectBackend 在执行。
// 无论调用多少次（每来一包、多路扇出并发），多余调用直接返回。
func (s *session) startReconnect(terminal Terminal) {
	select {
	case <-s.stopCh:
		return
	default:
	}
	if _, loaded := s.reconnecting.LoadOrStore(terminal.TargetAddr, struct{}{}); loaded {
		return
	}
	go s.reconnectBackend(terminal)
}

// reconnectBackend 后台重试 Dial，成功后 putBackend 写回可用 client。
// 整段重试（含 Timer 等待）期间 reconnecting 都占着位；退出时再删掉，允许日后再次重连。
func (s *session) reconnectBackend(terminal Terminal) {
	defer s.reconnecting.Delete(terminal.TargetAddr)

	timer := time.NewTimer(s.opts.TimeoutRetry)
	timer.Stop()
	defer timer.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		default:
		}
		onDead := func() {
			s.putBackend(terminal, nil)
			s.startReconnect(terminal)
		}
		c, err := newBackendClient(terminal, s.opts, s.stopCh, s.enqueueReply, onDead)
		if err == nil {
			s.putBackend(terminal, c)
			slog.Info("rejoin",
				slog.String("addr", terminal.TargetAddr))
			return
		}
		slog.Warn("reconnect backend error",
			slog.String("addr", terminal.TargetAddr),
			slog.Any("err", err))
		timer.Reset(s.opts.TimeoutRetry)
		select {
		case <-s.stopCh:
			return
		case <-timer.C:
		}
	}
}
