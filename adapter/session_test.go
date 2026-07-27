package adapter

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// TestStartReconnectSingleFlight 同一地址并发/反复 startReconnect，reconnecting 中只有一条，
// 且不会为每次调用各起一个长期重连协程。
func TestStartReconnectSingleFlight(t *testing.T) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	realConn, err := net.DialTCP("tcp", nil, listener.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer realConn.Close()

	serverConn, err := listener.AcceptTCP()
	if err != nil {
		t.Fatal(err)
	}
	defer serverConn.Close()

	// 不可达后端 + 较长重试间隔，保证测试窗口内仍处于 reconnecting
	s := newSession(serverConn, time.Hour, nil)
	defer s.stop()

	term := Terminal{Mode: Leader, TargetAddr: "127.0.0.1:1"}
	s.putBackend(term, nil)

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.startReconnect(term)
		}()
	}
	wg.Wait()

	// 再模拟「新数据到来」路径上的重复触发
	for i := 0; i < 64; i++ {
		s.startReconnect(term)
	}

	n := 0
	s.reconnecting.Range(func(key, _ any) bool {
		n++
		if key.(string) != term.TargetAddr {
			t.Fatalf("unexpected reconnecting key %v", key)
		}
		return true
	})
	if n != 1 {
		t.Fatalf("reconnecting entries = %d, want 1", n)
	}
}

// TestSessionEnqueueReplyReturnsSafelyAfterStop 关闭后 enqueueReply 应能返回，不永久阻塞。
func TestSessionEnqueueReplyReturnsSafelyAfterStop(t *testing.T) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	clientConn, err := net.DialTCP("tcp", nil, listener.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer clientConn.Close()

	serverConn, err := listener.AcceptTCP()
	if err != nil {
		t.Fatal(err)
	}
	s := newSession(serverConn, 0, nil)

	s.stop()

	done := make(chan struct{})
	go func() {
		defer close(done)
		s.enqueueReply([]byte("late reply"))
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("enqueueReply did not return after session stop")
	}
}

// TestSessionBackendsUseTargetAddrStringKeys 约束：backends 的 key 必须是 TargetAddr 字符串。
func TestSessionBackendsUseTargetAddrStringKeys(t *testing.T) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	badAddr := "127.0.0.1:1"
	realConn, err := net.DialTCP("tcp", nil, listener.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer realConn.Close()

	serverConn, err := listener.AcceptTCP()
	if err != nil {
		t.Fatal(err)
	}
	defer serverConn.Close()

	s := newSession(serverConn, time.Millisecond, []Terminal{
		{Mode: Leader, TargetAddr: badAddr},
	})
	defer s.stop()

	found := 0
	s.backends.Range(func(key, value any) bool {
		found++
		addr, ok := key.(string)
		if !ok {
			t.Fatalf("map key type = %T, want string", key)
		}
		if addr != badAddr {
			t.Fatalf("map key = %q, want %q", addr, badAddr)
		}
		b, ok := value.(*backend)
		if !ok || b == nil {
			t.Fatalf("map value type = %T, want *backend", value)
		}
		if b.terminal.TargetAddr != badAddr {
			t.Fatalf("backend.terminal.TargetAddr = %q", b.terminal.TargetAddr)
		}
		if b.client != nil {
			t.Fatal("expected nil client for unreachable backend")
		}
		return true
	})
	if found != 1 {
		t.Fatalf("map entries = %d, want 1", found)
	}
}

// TestSessionPutBackendReplacesSameAddr 同 TargetAddr 再次 put 应覆盖，map 中仅一条。
func TestSessionPutBackendReplacesSameAddr(t *testing.T) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	realConn, err := net.DialTCP("tcp", nil, listener.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer realConn.Close()

	serverConn, err := listener.AcceptTCP()
	if err != nil {
		t.Fatal(err)
	}
	defer serverConn.Close()

	s := newSession(serverConn, time.Millisecond, nil)
	defer s.stop()

	term := Terminal{Mode: Leader, TargetAddr: "backend-a"}
	s.putBackend(term, nil)
	s.putBackend(term, nil)

	count := 0
	s.backends.Range(func(key, _ any) bool {
		count++
		if key.(string) != "backend-a" {
			t.Fatalf("unexpected key %v", key)
		}
		return true
	})
	if count != 1 {
		t.Fatalf("entries after replace = %d, want 1", count)
	}
}

// TestBackendClientEnqueueKeepsPayloadAfterSourceMutates 验证独立 payload 约定：
// 污染原始读缓冲不会改变已入队并写出的内容。
func TestBackendClientEnqueueKeepsPayloadAfterSourceMutates(t *testing.T) {
	backendLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer backendLn.Close()

	var (
		got []byte
		wg  sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		c, err := backendLn.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
		got, _ = io.ReadAll(c)
	}()

	term := Terminal{Mode: Leader, TargetAddr: backendLn.Addr().String()}
	bc, err := newBackendClient(term, make(chan struct{}), func([]byte) {}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer bc.stop()

	src := []byte{0x7e, 0x01, 0x02, 0x03, 0x7e}
	payload := make([]byte, len(src))
	copy(payload, src)

	if ok := bc.enqueueUplink(payload); !ok {
		t.Fatal("enqueueUplink failed")
	}

	for i := range src {
		src[i] = 0xff
	}

	time.Sleep(50 * time.Millisecond)
	bc.stop()
	wg.Wait()

	want := []byte{0x7e, 0x01, 0x02, 0x03, 0x7e}
	if !bytes.Equal(got, want) {
		t.Fatalf("backend got %x, want %x", got, want)
	}
}
