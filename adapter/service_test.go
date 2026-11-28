package adapter

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

var (
	testUplinkHeartbeat = mustHex("7e0002000001234567890100008a7e")
	testDownlink8001    = mustHex("7e8001000500000000100100000002010201957e")
	testDownlink9205    = mustHex("7e920500180000000010010003012412091636532412101636530000000000000000000000857e")
	testBadFrame        = mustHex("7e000200000123456789017fff7e")
)

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

type mockBackend struct {
	t           *testing.T
	ln          net.Listener
	addr        string
	mu          sync.Mutex
	conns       []net.Conn
	activeConns atomic.Int32
	uplink      chan []byte
	connOpen    chan struct{}
	connClose   chan struct{}
	stopCh      chan struct{}
	stopOnce    sync.Once
}

func startMockBackend(t *testing.T, addr string) *mockBackend {
	t.Helper()
	var (
		ln  net.Listener
		err error
	)
	deadline := time.Now().Add(2 * time.Second)
	for {
		ln, err = net.Listen("tcp", addr)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("listen %s: %v", addr, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	m := &mockBackend{
		t:         t,
		ln:        ln,
		addr:      ln.Addr().String(),
		uplink:    make(chan []byte, 32),
		connOpen:  make(chan struct{}, 32),
		connClose: make(chan struct{}, 32),
		stopCh:    make(chan struct{}),
	}
	go m.acceptLoop()
	t.Cleanup(m.Close)
	return m
}

func (m *mockBackend) acceptLoop() {
	for {
		c, err := m.ln.Accept()
		if err != nil {
			return
		}
		m.mu.Lock()
		m.conns = append(m.conns, c)
		m.mu.Unlock()
		m.activeConns.Add(1)
		m.connOpen <- struct{}{}
		go m.readLoop(c)
	}
}

func (m *mockBackend) readLoop(c net.Conn) {
	defer func() {
		_ = c.Close()
		m.activeConns.Add(-1)
		m.connClose <- struct{}{}
	}()
	buf := make([]byte, 2048)
	for {
		n, err := c.Read(buf)
		if n > 0 {
			select {
			case m.uplink <- bytes.Clone(buf[:n]):
			case <-m.stopCh:
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func (m *mockBackend) writeDownlink(data []byte) {
	m.t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.conns) - 1; i >= 0; i-- {
		_ = m.conns[i].SetWriteDeadline(time.Now().Add(time.Second))
		if _, err := m.conns[i].Write(data); err == nil {
			return
		}
	}
	m.t.Fatal("no live backend conn to write downlink")
}

func (m *mockBackend) Close() {
	m.stopOnce.Do(func() {
		close(m.stopCh)
		_ = m.ln.Close()
		m.mu.Lock()
		for _, c := range m.conns {
			_ = c.Close()
		}
		m.mu.Unlock()
	})
}

type stubListener struct {
	steps chan func() (net.Conn, error)
}

func (l *stubListener) Accept() (net.Conn, error) {
	fn, ok := <-l.steps
	if !ok {
		return nil, net.ErrClosed
	}
	return fn()
}

func (l *stubListener) Close() error   { return nil }
func (l *stubListener) Addr() net.Addr { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 18080} }

func waitSignal(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func waitBytes(t *testing.T, ch <-chan []byte) []byte {
	t.Helper()
	select {
	case b := <-ch:
		return b
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for uplink bytes")
		return nil
	}
}

func TestOptionsAndCreateTerminals(t *testing.T) {
	adDefault := New()
	if adDefault.opts.Addr != "0.0.0.0:18080" ||
		adDefault.opts.DialTimeout != 10*time.Second ||
		adDefault.opts.TimeoutRetry != 10*time.Second ||
		adDefault.opts.WriteTimeout != 10*time.Second ||
		adDefault.opts.IdleTimeout != 0 ||
		adDefault.opts.MaxFrameBufferSize != 2048*2 ||
		len(adDefault.opts.AllowCommand) != 10 {
		t.Fatalf("unexpected default opts: %+v", adDefault.opts)
	}

	baseCmds := make([]consts.JT808CommandType, 1, 8)
	baseCmds[0] = consts.P9205QueryResourceList
	ad := New(
		WithHostPorts("127.0.0.1:28080"),
		WithDialTimeout(3*time.Second),
		WithDialTimeout(0),
		WithTimeoutRetry(5*time.Second),
		WithTimeoutRetry(-time.Second),
		WithWriteTimeout(4*time.Second),
		WithWriteTimeout(0),
		WithIdleTimeout(30*time.Second),
		WithIdleTimeout(0),
		WithMaxFrameBufferSize(0),
		WithAllowCommand(consts.P8001GeneralRespond),
		WithTerminals(
			Terminal{Mode: Follower, TargetAddr: "127.0.0.1:8080"},
			Terminal{Mode: Follower, TargetAddr: "127.0.0.1:8080", AllowCommands: baseCmds},
			Terminal{Mode: Follower, TargetAddr: ""},
		),
	)
	if ad.opts.DialTimeout != 3*time.Second || ad.opts.TimeoutRetry != 5*time.Second ||
		ad.opts.WriteTimeout != 4*time.Second || ad.opts.IdleTimeout != 30*time.Second ||
		ad.opts.MaxFrameBufferSize != 0 {
		t.Fatalf("unexpected custom opts: %+v", ad.opts)
	}

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ts := ad.createTerminals()
			if len(ts) != 1 || len(ts[0].AllowCommands) != 2 {
				t.Errorf("unexpected terminals: %+v", ts)
				return
			}
			ts[0].AllowCommands[0] = consts.T0002HeartBeat
		}()
	}
	wg.Wait()
	if baseCmds[0] != consts.P9205QueryResourceList {
		t.Fatalf("original AllowCommands mutated: %v", baseCmds)
	}
}

func TestAdapter_RunAndServe(t *testing.T) {
	// 1. Run 地址解析失败、监听失败与正常进入 serve 分支
	New(WithHostPorts(":::")).Run()
	New(WithHostPorts("203.0.113.1:18080")).Run()
	if probe, err := net.Listen("tcp", "127.0.0.1:0"); err == nil {
		_ = probe.Close()
		go New(WithHostPorts("127.0.0.1:0")).Run()
		time.Sleep(15 * time.Millisecond)
	}

	// 2. serve 临时错误退避 + 正常接收连接 + ErrClosed 退出
	srvConn, cliConn := net.Pipe()
	_ = cliConn.Close()
	ln := &stubListener{steps: make(chan func() (net.Conn, error), 3)}
	ln.steps <- func() (net.Conn, error) { return nil, errors.New("temporary emfile") }
	ln.steps <- func() (net.Conn, error) { return srvConn, nil }
	ln.steps <- func() (net.Conn, error) { return nil, net.ErrClosed }
	New().serve(ln)
}

func TestLifeCycle(t *testing.T) {
	leader := startMockBackend(t, "127.0.0.1:0")
	follower := startMockBackend(t, "127.0.0.1:0")

	in, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen adapter: %v", err)
	}
	t.Cleanup(func() { _ = in.Close() })

	ad := New(
		WithTimeoutRetry(30*time.Millisecond),
		WithTerminals(
			Terminal{Mode: Leader, TargetAddr: leader.addr},
			Terminal{
				Mode:          Follower,
				TargetAddr:    follower.addr,
				AllowCommands: []consts.JT808CommandType{consts.P9205QueryResourceList},
			},
		),
	)
	go ad.serve(in)

	dev, err := net.DialTimeout("tcp", in.Addr().String(), time.Second)
	if err != nil {
		t.Fatalf("dial adapter: %v", err)
	}
	defer func() { _ = dev.Close() }()

	waitSignal(t, leader.connOpen, "leader accept")
	waitSignal(t, follower.connOpen, "follower accept")

	// 上行扇出至双端
	if _, err := dev.Write(testUplinkHeartbeat); err != nil {
		t.Fatalf("write uplink: %v", err)
	}
	if got := waitBytes(t, leader.uplink); !bytes.Equal(got, testUplinkHeartbeat) {
		t.Fatalf("leader uplink got %x", got)
	}
	if got := waitBytes(t, follower.uplink); !bytes.Equal(got, testUplinkHeartbeat) {
		t.Fatalf("follower uplink got %x", got)
	}

	// Leader 任意下行可达；Follower 0x8001 拦截、0x9205 放行
	leader.writeDownlink(testDownlink8001)
	got := make([]byte, len(testDownlink8001))
	_ = dev.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := io.ReadFull(dev, got); err != nil || !bytes.Equal(got, testDownlink8001) {
		t.Fatalf("leader downlink got %x, err=%v", got, err)
	}

	follower.writeDownlink(testDownlink8001)
	_ = dev.SetReadDeadline(time.Now().Add(80 * time.Millisecond))
	if n, err := dev.Read(make([]byte, 64)); n > 0 || err == nil {
		t.Fatalf("unexpected follower 0x8001 downlink n=%d err=%v", n, err)
	}

	follower.writeDownlink(testDownlink9205)
	got9205 := make([]byte, len(testDownlink9205))
	_ = dev.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := io.ReadFull(dev, got9205); err != nil || !bytes.Equal(got9205, testDownlink9205) {
		t.Fatalf("follower 0x9205 got %x, err=%v", got9205, err)
	}

	// Leader 宕机 → 重启自动重连 → 再次宕机（覆盖重连后生成的 onDead 回调）→ 终端下线回收所有连接
	leaderAddr := leader.addr
	leader.Close()
	waitSignal(t, leader.connClose, "leader first close")

	if _, err := dev.Write(testUplinkHeartbeat); err != nil {
		t.Fatalf("write while leader down: %v", err)
	}
	_ = waitBytes(t, follower.uplink)

	leader = startMockBackend(t, leaderAddr)
	waitSignal(t, leader.connOpen, "leader rejoin")

	// Accept 返回时 reconnectBackend 可能尚未完成 putBackend，短轮询写入直至重连后的 leader 收到上行
	deadline := time.Now().Add(2 * time.Second)
	for received := false; !received; {
		if _, err := dev.Write(testUplinkHeartbeat); err != nil {
			t.Fatalf("write after leader rejoin: %v", err)
		}
		select {
		case <-leader.uplink:
			received = true
		case <-time.After(10 * time.Millisecond):
			if time.Now().After(deadline) {
				t.Fatalf("timed out waiting for leader uplink after rejoin")
			}
		}
	}
	_ = waitBytes(t, follower.uplink)

	// 再次关闭重连后的 leader，覆盖 reconnectBackend 内的 onDead 回调
	leader.Close()
	waitSignal(t, leader.connClose, "leader second close")
	time.Sleep(20 * time.Millisecond)

	// 真实终端断开后，follower 模拟连接同步断开
	_ = dev.Close()
	waitSignal(t, follower.connClose, "follower close after device offline")
}
