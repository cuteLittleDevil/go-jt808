package adapter

import (
	"bytes"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

func TestOptions_DefaultsAndOverrides(t *testing.T) {
	adDefault := New()
	if adDefault.opts.Addr != "0.0.0.0:18080" ||
		adDefault.opts.DialTimeout != 10*time.Second ||
		adDefault.opts.TimeoutRetry != 10*time.Second ||
		adDefault.opts.WriteTimeout != 10*time.Second ||
		adDefault.opts.IdleTimeout != 0 ||
		readBufSize != 2048 ||
		adDefault.opts.MaxFrameBufferSize != 2048*2 ||
		len(adDefault.opts.AllowCommand) != 10 {
		t.Fatalf("unexpected default opts: %+v", adDefault.opts)
	}

	adCustom := New(
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
		WithAllowCommand(),
		WithTerminals(
			Terminal{Mode: Follower, TargetAddr: "127.0.0.1:8080"},
			Terminal{Mode: Leader, TargetAddr: "127.0.0.1:8080"},
			Terminal{Mode: Follower, TargetAddr: ""},
		),
	)
	if adCustom.opts.DialTimeout != 3*time.Second ||
		adCustom.opts.TimeoutRetry != 5*time.Second ||
		adCustom.opts.WriteTimeout != 4*time.Second ||
		adCustom.opts.IdleTimeout != 30*time.Second ||
		adCustom.opts.MaxFrameBufferSize != 0 ||
		len(adCustom.opts.AllowCommand) != 0 {
		t.Fatalf("unexpected custom opts: %+v", adCustom.opts)
	}
	ts := adCustom.createTerminals()
	if len(ts) != 1 || ts[0].Mode != Leader {
		t.Fatalf("createTerminals() = %+v, want single deduplicated Leader", ts)
	}
}

func TestAdapter_CreateTerminalsDoesNotShareBackingArray(t *testing.T) {
	// 故意预留多余 cap，验证多连接并发 createTerminals 不会踩同一底层数组
	baseCmds := make([]consts.JT808CommandType, 1, 8)
	baseCmds[0] = consts.P9205QueryResourceList

	ad := New(
		WithAllowCommand(consts.P8001GeneralRespond),
		WithTerminals(Terminal{
			Mode:          Follower,
			TargetAddr:    "127.0.0.1:8080",
			AllowCommands: baseCmds,
		}),
	)

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
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

func TestAdapter_StopConcurrentWithRunNoRace(t *testing.T) {
	ad := New(WithHostPorts("127.0.0.1:0"))
	runDone := make(chan struct{})
	go func() {
		defer close(runDone)
		ad.Run()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ad.Stop()
		}()
	}
	wg.Wait()

	select {
	case <-runDone:
	case <-time.After(time.Second):
		t.Fatal("Run() did not exit after Stop()")
	}
}

func TestBackendClient_ReadLoopDownlinkCopyNoRace(t *testing.T) {
	// 连续下发两种不同报文，验证异步消费 onReply 时不会因 readBuf 覆写导致数据竞争或串包
	pkt1 := testDownlink8001
	pkt2 := testDownlink9205

	clientEnd, serverEnd := net.Pipe()
	defer func() {
		_ = clientEnd.Close()
		_ = serverEnd.Close()
	}()

	const total = 100
	replies := make(chan []byte, total)
	bc := &backendClient{
		terminal:    Terminal{Mode: Leader, TargetAddr: "pipe"},
		opts:        newOptions(nil),
		sessionStop: make(chan struct{}),
		onReply: func(data []byte) {
			replies <- data
		},
		conn:       clientEnd,
		writeQueue: make(chan []byte, 10),
		stopCh:     make(chan struct{}),
	}

	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		bc.readLoop()
	}()

	go func() {
		for i := 0; i < total; i++ {
			pkt := pkt1
			if i%2 == 1 {
				pkt = pkt2
			}
			if _, err := serverEnd.Write(pkt); err != nil {
				t.Errorf("write[%d] error = %v", i, err)
				return
			}
		}
		_ = serverEnd.Close()
	}()

	for i := 0; i < total; i++ {
		want := pkt1
		if i%2 == 1 {
			want = pkt2
		}
		select {
		case got := <-replies:
			if !bytes.Equal(got, want) {
				t.Fatalf("reply[%d] = %x, want %x", i, got, want)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for reply[%d]", i)
		}
	}

	select {
	case <-readDone:
	case <-time.After(time.Second):
		t.Fatal("readLoop did not exit")
	}
}

func TestBackendClient_FrameBufferOverflowAndUnpackError(t *testing.T) {
	t.Run("unclosed 0x7e reaching MaxFrameBufferSize disconnects backend", func(t *testing.T) {
		clientEnd, serverEnd := net.Pipe()
		defer func() {
			_ = clientEnd.Close()
			_ = serverEnd.Close()
		}()

		var deadCalled atomic.Int32
		readDone := make(chan struct{})
		bc := &backendClient{
			terminal:    Terminal{Mode: Leader, TargetAddr: "pipe-overflow"},
			opts:        newOptions([]Option{WithMaxFrameBufferSize(64)}),
			sessionStop: make(chan struct{}),
			onReply:     func([]byte) {},
			onDead: func() {
				deadCalled.Add(1)
			},
			conn:       clientEnd,
			writeQueue: make(chan []byte, 10),
			stopCh:     make(chan struct{}),
		}
		go func() {
			defer close(readDone)
			bc.readLoop()
		}()

		unclosed := append([]byte{0x7e}, bytes.Repeat([]byte{0x01}, 64)...)
		if _, err := serverEnd.Write(unclosed); err != nil {
			t.Fatalf("write unclosed frame: %v", err)
		}

		select {
		case <-readDone:
		case <-time.After(time.Second):
			t.Fatal("readLoop did not exit on frame buffer overflow")
		}
		if deadCalled.Load() != 1 {
			t.Fatalf("onDead called = %d, want 1", deadCalled.Load())
		}
	})

	t.Run("valid frame followed by malformed frame delivers valid then disconnects", func(t *testing.T) {
		clientEnd, serverEnd := net.Pipe()
		defer func() {
			_ = clientEnd.Close()
			_ = serverEnd.Close()
		}()

		replies := make(chan []byte, 4)
		var deadCalled atomic.Int32
		readDone := make(chan struct{})
		bc := &backendClient{
			terminal:    Terminal{Mode: Leader, TargetAddr: "pipe-bad-frame"},
			opts:        newOptions(nil),
			sessionStop: make(chan struct{}),
			onReply: func(data []byte) {
				replies <- data
			},
			onDead: func() {
				deadCalled.Add(1)
			},
			conn:       clientEnd,
			writeQueue: make(chan []byte, 10),
			stopCh:     make(chan struct{}),
		}
		go func() {
			defer close(readDone)
			bc.readLoop()
		}()

		badFrame := mustHex("7e000200000123456789017fff7e")
		stream := append(append([]byte(nil), testDownlink9205...), badFrame...)
		if _, err := serverEnd.Write(stream); err != nil {
			t.Fatalf("write mixed stream: %v", err)
		}

		select {
		case got := <-replies:
			if !bytes.Equal(got, testDownlink9205) {
				t.Fatalf("reply = %x, want %x", got, testDownlink9205)
			}
		case <-time.After(time.Second):
			t.Fatal("missing valid frame before malformed frame")
		}
		select {
		case <-readDone:
		case <-time.After(time.Second):
			t.Fatal("readLoop did not exit on malformed frame")
		}
		if deadCalled.Load() != 1 {
			t.Fatalf("onDead called = %d, want 1", deadCalled.Load())
		}
	})
}

func TestSession_IdleTimeoutWriteFailureAndTimerStop(t *testing.T) {
	t.Run("idle timeout closes session and backend", func(t *testing.T) {
		devClient, devServer := net.Pipe()
		defer func() {
			_ = devClient.Close()
			_ = devServer.Close()
		}()

		backendClientConn, backendServerConn := net.Pipe()
		defer backendServerConn.Close()

		closed := make(chan struct{})
		s := newSessionWithOptions(devServer, newOptions([]Option{
			WithIdleTimeout(40 * time.Millisecond),
		}), nil, func() {
			close(closed)
		})
		term := Terminal{Mode: Leader, TargetAddr: "pipe-backend"}
		bc := &backendClient{
			terminal:    term,
			sessionStop: s.stopCh,
			conn:        backendClientConn,
			writeQueue:  make(chan []byte, 10),
			stopCh:      make(chan struct{}),
		}
		s.putBackend(term, bc)
		go s.readLoop()
		go s.writeLoop()

		select {
		case <-closed:
		case <-time.After(time.Second):
			t.Fatal("session did not close on idle timeout")
		}
		select {
		case <-bc.stopCh:
		default:
			t.Fatal("backendClient was not stopped when session idle-timed out")
		}
	})

	t.Run("write failure stops session", func(t *testing.T) {
		devClient, devServer := net.Pipe()
		_ = devClient.Close()

		closed := make(chan struct{})
		s := newSessionWithOptions(devServer, newOptions([]Option{
			WithWriteTimeout(50 * time.Millisecond),
		}), nil, func() {
			close(closed)
		})
		go s.writeLoop()
		s.enqueueReply(testDownlink8001)

		select {
		case <-closed:
		case <-time.After(time.Second):
			t.Fatal("session did not stop after write failure")
		}
	})

	t.Run("reconnect timer exits immediately on stop", func(t *testing.T) {
		devClient, devServer := net.Pipe()
		defer func() {
			_ = devClient.Close()
			_ = devServer.Close()
		}()

		s := newSessionWithOptions(devServer, newOptions([]Option{
			WithTimeoutRetry(time.Hour),
			WithDialTimeout(50 * time.Millisecond),
		}), nil, nil)
		term := Terminal{Mode: Leader, TargetAddr: "127.0.0.1:99999"}
		s.putBackend(term, nil)

		done := make(chan struct{})
		s.reconnecting.Store(term.TargetAddr, struct{}{})
		go func() {
			defer close(done)
			s.reconnectBackend(term)
		}()

		time.Sleep(20 * time.Millisecond)
		s.stop()

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("reconnectBackend did not exit promptly after session stop")
		}
	})
}
