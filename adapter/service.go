package adapter

import (
	"errors"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type Adapter struct {
	opts     *Options
	ln       atomic.Pointer[net.TCPListener]
	sessions sync.Map
	stopCh   chan struct{}
	stopOnce sync.Once
}

func New(opts ...Option) *Adapter {
	options := newOptions(opts)
	return &Adapter{
		opts:   options,
		stopCh: make(chan struct{}),
	}
}

// Run 启动适配器并阻塞接受真实终端连接，直到 Stop 或监听失败。
func (a *Adapter) Run() {
	addr, err := net.ResolveTCPAddr("tcp", a.opts.Addr)
	if err != nil {
		slog.Error("resolve tcp addr error",
			slog.String("addr", a.opts.Addr),
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
	a.ln.Store(in)
	select {
	case <-a.stopCh:
		if ln := a.ln.Swap(nil); ln != nil {
			_ = ln.Close()
		}
		return
	default:
	}

	for {
		c, err := in.AcceptTCP()
		if err != nil {
			select {
			case <-a.stopCh:
				return
			default:
			}
			// 监听被关闭时 Accept 也会失败
			if errors.Is(err, net.ErrClosed) {
				return
			}
			slog.Warn("accept fail",
				slog.Any("err", err))
			time.Sleep(100 * time.Millisecond) // 如文件描述符耗尽 稍等再重试 避免空转
			continue
		}

		var s *session
		s = newSessionWithOptions(c, a.opts, a.createTerminals(), func() {
			a.sessions.Delete(s)
		})
		a.sessions.Store(s, struct{}{})
		select {
		case <-a.stopCh:
			s.stop()
			return
		default:
		}
		go s.run()
	}
}

// Stop 关闭监听并主动关闭所有已建立的终端会话及其后端模拟连接。
func (a *Adapter) Stop() {
	a.stopOnce.Do(func() {
		close(a.stopCh)
		if ln := a.ln.Swap(nil); ln != nil {
			_ = ln.Close()
		}
		a.sessions.Range(func(key, _ any) bool {
			if s, ok := key.(*session); ok {
				s.stop()
			}
			return true
		})
	})
}

func (a *Adapter) createTerminals() []Terminal {
	ts := make([]Terminal, 0, len(a.opts.Terminals))
	seen := make(map[string]int, len(a.opts.Terminals))
	for _, t := range a.opts.Terminals {
		if t.TargetAddr == "" {
			continue
		}
		cmds := make([]consts.JT808CommandType, 0, len(t.AllowCommands)+len(a.opts.AllowCommand))
		cmds = append(cmds, t.AllowCommands...)
		cmds = append(cmds, a.opts.AllowCommand...)
		t.AllowCommands = cmds
		if idx, ok := seen[t.TargetAddr]; ok {
			ts[idx] = t
			continue
		}
		seen[t.TargetAddr] = len(ts)
		ts = append(ts, t)
	}
	return ts
}
