package adapter

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type errReadConn struct {
	net.Conn
	err error
}

func (e *errReadConn) Read([]byte) (int, error) { return 0, e.err }

func TestBackendClient_BranchesAndDownlinkSafety(t *testing.T) {
	// 1. 粘包连续下发零串包 + 粘包多帧拆包路径 + 坏帧断开并回调 onDead
	clientEnd, serverEnd := net.Pipe()
	defer func() {
		_ = clientEnd.Close()
		_ = serverEnd.Close()
	}()

	replies := make(chan []byte, 8)
	var deadCalled atomic.Int32
	readDone := make(chan struct{})
	bc := &backendClient{
		terminal:    Terminal{Mode: Leader, TargetAddr: "pipe"},
		opts:        newOptions(nil),
		sessionStop: make(chan struct{}),
		onReply:     func(data []byte) { replies <- data },
		onDead:      func() { deadCalled.Add(1) },
		conn:        clientEnd,
		writeQueue:  make(chan []byte, 1),
		stopCh:      make(chan struct{}),
	}
	go func() {
		defer close(readDone)
		bc.readLoop()
	}()

	// 单帧快路径 + 多帧粘包 + 坏帧触发 unpackErr 退出
	if _, err := serverEnd.Write(testDownlink8001); err != nil {
		t.Fatal(err)
	}
	combo := append(append(bytes.Clone(testDownlink8001), testDownlink9205...), testBadFrame...)
	if _, err := serverEnd.Write(combo); err != nil {
		t.Fatal(err)
	}
	waitSignal(t, readDone, "readLoop exit on bad frame")
	if deadCalled.Load() != 1 || len(replies) != 3 {
		t.Fatalf("deadCalled=%d replies=%d, want 1 and 3", deadCalled.Load(), len(replies))
	}

	// 2. 粘包缓冲区达到 MaxFrameBufferSize 断开连接
	ovClient, ovServer := net.Pipe()
	defer ovServer.Close()
	ovDone := make(chan struct{})
	bcOverflow := &backendClient{
		terminal:    Terminal{Mode: Leader, TargetAddr: "pipe-overflow"},
		opts:        newOptions([]Option{WithMaxFrameBufferSize(32)}),
		sessionStop: make(chan struct{}),
		onReply:     func([]byte) {},
		conn:        ovClient,
		writeQueue:  make(chan []byte, 1),
		stopCh:      make(chan struct{}),
	}
	go func() {
		defer close(ovDone)
		bcOverflow.readLoop()
	}()
	_, _ = ovServer.Write(append([]byte{0x7e}, bytes.Repeat([]byte{0x01}, 32)...))
	waitSignal(t, ovDone, "readLoop exit on buffer overflow")

	// 3. enqueueUplink 在队列满且 stopCh / sessionStop 关闭时返回 false；run 写失败触发退出
	wClient, wServer := net.Pipe()
	_ = wServer.Close()
	sessStop := make(chan struct{})
	bcWriteErr := &backendClient{
		terminal:    Terminal{Mode: Leader, TargetAddr: "pipe-write-err"},
		opts:        newOptions(nil),
		sessionStop: sessStop,
		conn:        wClient,
		writeQueue:  make(chan []byte, 1),
		stopCh:      make(chan struct{}),
	}
	if !bcWriteErr.enqueueUplink([]byte{0x7e, 0x01, 0x7e}) {
		t.Fatal("expected enqueueUplink true before stop")
	}
	bcWriteErr.run() // writeQueue 中有 1 包，写入已关闭的 wClient 报错退出并关闭 stopCh

	// 填满队列使 select 必走 stopCh / sessionStop 分支
	bcWriteErr.writeQueue <- []byte{1}
	if bcWriteErr.enqueueUplink([]byte{2}) {
		t.Fatal("expected false when stopCh closed")
	}

	bcSessClosed := &backendClient{
		sessionStop: sessStop,
		writeQueue:  make(chan []byte, 1),
		stopCh:      make(chan struct{}),
	}
	bcSessClosed.writeQueue <- []byte{1}
	close(sessStop)
	if bcSessClosed.enqueueUplink([]byte{2}) {
		t.Fatal("expected false when sessionStop closed")
	}

	// 4. newBackendClient(opts=nil) 独立 payload 不被源切片修改污染
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	var got []byte
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		got, _ = io.ReadAll(c)
	}()
	realBC, err := newBackendClient(Terminal{Mode: Leader, TargetAddr: ln.Addr().String()}, nil, make(chan struct{}), func([]byte) {}, nil)
	if err != nil {
		t.Fatal(err)
	}
	src := []byte{0x7e, 0x01, 0x02, 0x7e}
	payload := bytes.Clone(src)
	realBC.enqueueUplink(payload)
	src[1] = 0xff
	time.Sleep(30 * time.Millisecond)
	realBC.stop()
	wg.Wait()
	if !bytes.Equal(got, []byte{0x7e, 0x01, 0x02, 0x7e}) {
		t.Fatalf("backend got %x", got)
	}
}

func TestSession_EdgeBranchesAndReconnect(t *testing.T) {
	// 1. 单飞重连 + TargetAddr 字符串 key 约束 + Timer 在 stop 后立即退出
	devA, devB := net.Pipe()
	defer func() {
		_ = devA.Close()
		_ = devB.Close()
	}()

	term := Terminal{Mode: Leader, TargetAddr: "127.0.0.1:1"}
	s := newSession(devA, time.Hour, []Terminal{term})
	s.putBackend(term, nil)

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.startReconnect(term)
		}()
	}
	wg.Wait()

	entries := 0
	s.backends.Range(func(key, value any) bool {
		entries++
		if key.(string) != term.TargetAddr || value.(*backend).client != nil {
			t.Fatalf("unexpected backend entry %v -> %+v", key, value)
		}
		return true
	})
	reconnCount := 0
	s.reconnecting.Range(func(_, _ any) bool {
		reconnCount++
		return true
	})
	if entries != 1 || reconnCount != 1 {
		t.Fatalf("entries=%d reconnCount=%d, want 1 and 1", entries, reconnCount)
	}
	s.stop()

	// session 已 stop 后的各安全早退分支：run / startReconnect / reconnectBackend / enqueueReply(队列满)
	s.run()
	s.startReconnect(term)
	s.reconnectBackend(term)
	for len(s.replyQueue) < cap(s.replyQueue) {
		s.replyQueue <- []byte{1}
	}
	s.enqueueReply([]byte{2})

	// 2. s.run 首次 dialBackend 失败分支 + readLoop 扇出遇到已失效 client 分支 + 非预期读错误分支
	upClient, upServer := net.Pipe()
	defer upClient.Close()
	sFail := newSessionWithOptions(upServer, newOptions([]Option{
		WithDialTimeout(30 * time.Millisecond),
		WithTimeoutRetry(time.Hour),
	}), []Terminal{{Mode: Leader, TargetAddr: "127.0.0.1:99999"}}, nil)
	sFail.run() // 触发 dialBackend 拨号失败日志与重连占位

	// 塞入一个已关闭但尚未从 map 摘除的 backendClient，验证 readLoop 扇出失败置空分支
	deadTerm := Terminal{Mode: Follower, TargetAddr: "dead-backend"}
	deadBC := &backendClient{
		terminal:    deadTerm,
		sessionStop: sFail.stopCh,
		writeQueue:  make(chan []byte, 1),
		stopCh:      make(chan struct{}),
	}
	deadBC.writeQueue <- []byte{1}
	close(deadBC.stopCh)
	sFail.putBackend(deadTerm, deadBC)

	_, _ = upClient.Write(testUplinkHeartbeat)
	time.Sleep(30 * time.Millisecond)
	sFail.stop()

	// 非 EOF / 非超时的普通读错误分支
	baseA, baseB := net.Pipe()
	defer baseB.Close()
	sReadErr := newSession(&errReadConn{Conn: baseA, err: errors.New("boom")}, time.Second, nil)
	sReadErr.readLoop()

	// 3. IdleTimeout 空闲超时与 WriteLoop 写失败主动关闭会话
	idleCli, idleSrv := net.Pipe()
	defer idleCli.Close()
	bcPipeA, bcPipeB := net.Pipe()
	defer bcPipeB.Close()

	idleClosed := make(chan struct{})
	sIdle := newSessionWithOptions(idleSrv, newOptions([]Option{
		WithIdleTimeout(30 * time.Millisecond),
	}), nil, func() { close(idleClosed) })
	bcAlive := &backendClient{
		terminal:    term,
		sessionStop: sIdle.stopCh,
		conn:        bcPipeA,
		writeQueue:  make(chan []byte, 1),
		stopCh:      make(chan struct{}),
	}
	sIdle.putBackend(term, bcAlive)
	go sIdle.readLoop()
	go sIdle.writeLoop()
	waitSignal(t, idleClosed, "idle timeout close")

	writeCli, writeSrv := net.Pipe()
	_ = writeCli.Close()
	writeClosed := make(chan struct{})
	sWrite := newSessionWithOptions(writeSrv, newOptions([]Option{
		WithWriteTimeout(30 * time.Millisecond),
	}), nil, func() { close(writeClosed) })
	go sWrite.writeLoop()
	sWrite.enqueueReply(testDownlink8001)
	waitSignal(t, writeClosed, "write error close")
}
