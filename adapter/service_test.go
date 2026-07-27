package adapter

import (
	"bytes"
	"encoding/hex"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

// 真实终端上行心跳（0x0002），与 service 单测同源。
var testUplinkHeartbeat = mustHex("7e0002000001234567890100008a7e")

// 平台通用应答 0x8001（Leader 可回写；Follower 默认不允许）。
var testDownlink8001 = mustHex("7e8001000500000000100100000002010201957e")

// 平台查询资源列表 0x9205（Follower AllowCommands 放行）。
var testDownlink9205 = mustHex("7e920500180000000010010003012412091636532412101636530000000000000000000000857e")

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// mockBackend 模拟一个后端 JT808 TCP 服务：记录连接/上行，可主动下行。
type mockBackend struct {
	t    *testing.T
	ln   net.Listener
	addr string

	mu          sync.Mutex
	conns       []net.Conn
	activeConns int32

	uplink    chan []byte
	connOpen  chan struct{} // 每次 Accept 成功发一记
	connClose chan struct{} // 每次连接读结束发一记
	stopCh    chan struct{}
	stopOnce  sync.Once
}

func startMockBackend(t *testing.T) *mockBackend {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen mock backend: %v", err)
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

// restartMockBackend 在同一地址重新监听（模拟 808 挂掉又恢复）。
func restartMockBackend(t *testing.T, addr string) *mockBackend {
	t.Helper()
	var ln net.Listener
	var err error
	deadline := time.Now().Add(2 * time.Second)
	for {
		ln, err = net.Listen("tcp", addr)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("re-listen %s: %v", addr, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	m := &mockBackend{
		t:         t,
		ln:        ln,
		addr:      addr,
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
			select {
			case <-m.stopCh:
				return
			default:
				return
			}
		}
		m.mu.Lock()
		m.conns = append(m.conns, c)
		m.mu.Unlock()
		atomic.AddInt32(&m.activeConns, 1)
		select {
		case m.connOpen <- struct{}{}:
		default:
		}
		go m.readLoop(c)
	}
}

func (m *mockBackend) readLoop(c net.Conn) {
	defer func() {
		_ = c.Close()
		atomic.AddInt32(&m.activeConns, -1)
		select {
		case m.connClose <- struct{}{}:
		default:
		}
	}()
	buf := make([]byte, 2048)
	for {
		n, err := c.Read(buf)
		if n > 0 {
			pkt := append([]byte(nil), buf[:n]...)
			select {
			case m.uplink <- pkt:
			case <-m.stopCh:
				return
			}
		}
		if err != nil {
			return
		}
	}
}

// writeDownlink 向当前仍存活的连接写入平台下行（多连接时写最新一条）。
func (m *mockBackend) writeDownlink(data []byte) {
	m.t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.conns) - 1; i >= 0; i-- {
		c := m.conns[i]
		_ = c.SetWriteDeadline(time.Now().Add(time.Second))
		if _, err := c.Write(data); err == nil {
			return
		}
	}
	m.t.Fatal("no live backend conn to write downlink")
}

func (m *mockBackend) ActiveConns() int {
	return int(atomic.LoadInt32(&m.activeConns))
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

func freeTCPAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freeTCPAddr: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func waitListen(t *testing.T, addr string) {
	t.Helper()
	waitFor(t, 2*time.Second, func() bool {
		c, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err != nil {
			return false
		}
		_ = c.Close()
		return true
	})
	// 探测连接会创建短生命周期 session，稍等后端侧连接回收
	time.Sleep(50 * time.Millisecond)
}

func waitBytes(t *testing.T, ch <-chan []byte, timeout time.Duration) []byte {
	t.Helper()
	select {
	case b := <-ch:
		return b
	case <-time.After(timeout):
		t.Fatal("timed out waiting for uplink bytes")
		return nil
	}
}

func waitSignal(t *testing.T, ch <-chan struct{}, timeout time.Duration, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func assertNoRead(t *testing.T, conn net.Conn, wait time.Duration) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(wait))
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if n > 0 {
		t.Fatalf("unexpected downlink %x", buf[:n])
	}
	if err == nil {
		t.Fatal("expected read timeout/error, got nil err")
	}
}

func readExact(t *testing.T, conn net.Conn, want []byte, timeout time.Duration) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatalf("ReadFull downlink: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("downlink got %x, want %x", got, want)
	}
}

func dialDevice(t *testing.T, addr string) net.Conn {
	t.Helper()
	c, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("dial adapter: %v", err)
	}
	return c
}

// TestLifeCycle 覆盖 adapter 生命周期与 Leader/Follower 回写策略（对齐 README / SDD AC-ADP-01）。
//
//  1. 两端 JT808 均在线时启动代理
//  2. 真实终端上线后，两端均收到上行；Leader 下行可达；Follower 仅 AllowCommands 可达
//  3. 真实终端下线后，两端模拟连接断开
//  4. 真实终端再次上线后恢复扇出
//  5. 后端意外挂掉再恢复后，经 TimeoutRetry 重连并恢复正常
func TestLifeCycle(t *testing.T) {
	leader := startMockBackend(t)
	follower := startMockBackend(t)

	adapterAddr := freeTCPAddr(t)
	retry := 50 * time.Millisecond
	ad := New(
		WithHostPorts(adapterAddr),
		WithTimeoutRetry(retry),
		WithTerminals(
			Terminal{
				Mode:       Leader,
				TargetAddr: leader.addr,
			},
			Terminal{
				Mode:       Follower,
				TargetAddr: follower.addr,
				AllowCommands: []consts.JT808CommandType{
					consts.P9205QueryResourceList,
				},
			},
		),
	)
	go ad.Run()
	t.Cleanup(ad.Stop)
	waitListen(t, adapterAddr)

	// ---------- 1 & 2：双后端在线，终端上线，上行扇出 + 下行策略 ----------
	dev := dialDevice(t, adapterAddr)

	// session 会向 Leader/Follower 各 Dial 一次
	waitSignal(t, leader.connOpen, 2*time.Second, "leader accept")
	waitSignal(t, follower.connOpen, 2*time.Second, "follower accept")

	if _, err := dev.Write(testUplinkHeartbeat); err != nil {
		t.Fatalf("device write uplink: %v", err)
	}
	if got := waitBytes(t, leader.uplink, 2*time.Second); !bytes.Equal(got, testUplinkHeartbeat) {
		t.Fatalf("leader uplink got %x, want %x", got, testUplinkHeartbeat)
	}
	if got := waitBytes(t, follower.uplink, 2*time.Second); !bytes.Equal(got, testUplinkHeartbeat) {
		t.Fatalf("follower uplink got %x, want %x", got, testUplinkHeartbeat)
	}

	// Leader 任意下行 → 真实终端
	leader.writeDownlink(testDownlink8001)
	readExact(t, dev, testDownlink8001, 2*time.Second)

	// Follower 非允许命令（0x8001）→ 不应到达真实终端
	follower.writeDownlink(testDownlink8001)
	assertNoRead(t, dev, 150*time.Millisecond)

	// Follower 允许的 0x9205 → 真实终端
	follower.writeDownlink(testDownlink9205)
	readExact(t, dev, testDownlink9205, 2*time.Second)

	// ---------- 3：终端下线 → 两端模拟连接断开 ----------
	_ = dev.Close()
	waitSignal(t, leader.connClose, 2*time.Second, "leader conn close after device offline")
	waitSignal(t, follower.connClose, 2*time.Second, "follower conn close after device offline")
	waitFor(t, 2*time.Second, func() bool {
		return leader.ActiveConns() == 0 && follower.ActiveConns() == 0
	})

	// ---------- 4：终端重新上线 → 恢复扇出 ----------
	dev = dialDevice(t, adapterAddr)
	t.Cleanup(func() { _ = dev.Close() })
	waitSignal(t, leader.connOpen, 2*time.Second, "leader accept on re-online")
	waitSignal(t, follower.connOpen, 2*time.Second, "follower accept on re-online")

	if _, err := dev.Write(testUplinkHeartbeat); err != nil {
		t.Fatalf("device rewrite uplink: %v", err)
	}
	if got := waitBytes(t, leader.uplink, 2*time.Second); !bytes.Equal(got, testUplinkHeartbeat) {
		t.Fatalf("leader uplink after re-online got %x", got)
	}
	if got := waitBytes(t, follower.uplink, 2*time.Second); !bytes.Equal(got, testUplinkHeartbeat) {
		t.Fatalf("follower uplink after re-online got %x", got)
	}

	// ---------- 5：Leader 挂掉 → 恢复 → 重连后两端再收上行 ----------
	leaderAddr := leader.addr
	leader.Close()
	// mock 关连接后，adapter 侧 readLoop/onDead 应立刻进入重连
	waitFor(t, 2*time.Second, func() bool {
		return leader.ActiveConns() == 0
	})

	// Follower 仍在，上行应仍能到 Follower
	if _, err := dev.Write(testUplinkHeartbeat); err != nil {
		t.Fatalf("device write while leader down: %v", err)
	}
	if got := waitBytes(t, follower.uplink, 2*time.Second); !bytes.Equal(got, testUplinkHeartbeat) {
		t.Fatalf("follower uplink while leader down got %x", got)
	}

	// 先起监听再等 Accept：重连协程按 TimeoutRetry 轮询 Dial
	leader = restartMockBackend(t, leaderAddr)
	waitSignal(t, leader.connOpen, 3*time.Second, "leader re-accept after restart")

	if _, err := dev.Write(testUplinkHeartbeat); err != nil {
		t.Fatalf("device write after leader recover: %v", err)
	}
	if got := waitBytes(t, leader.uplink, 3*time.Second); !bytes.Equal(got, testUplinkHeartbeat) {
		t.Fatalf("leader uplink after recover got %x, want %x", got, testUplinkHeartbeat)
	}
	if got := waitBytes(t, follower.uplink, 2*time.Second); !bytes.Equal(got, testUplinkHeartbeat) {
		t.Fatalf("follower uplink after leader recover got %x", got)
	}

	// Leader 恢复后下行仍可用
	leader.writeDownlink(testDownlink8001)
	readExact(t, dev, testDownlink8001, 2*time.Second)
}
