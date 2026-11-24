package service

import (
	"bytes"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

func newPipeConnection(t *testing.T, opts ...Option) (*GoJT808, *connection, net.Conn, *recordingTerminalEvent) {
	t.Helper()
	events := newRecordingTerminalEvent()
	allOpts := append([]Option{
		WithCustomTerminalEventer(func() TerminalEventer { return events }),
	}, opts...)
	g := New(allOpts...)

	serverConn, clientConn := net.Pipe()
	t.Cleanup(func() {
		_ = clientConn.Close()
		_ = serverConn.Close()
	})

	conn := newConnection(connectionParams{
		conn:                   serverConn,
		handles:                g.createCommandHandle(),
		activeRespondHandles:   g.createActiveRespondHandle(),
		terminalEvent:          g.opts.CustomTerminalEventerFunc(),
		filter:                 g.opts.FilterSubcontract,
		onTerminalTimeoutEvent: g.opts.OnTerminalTimeoutEvent,
		timeout: TerminalTimeout{
			ConnectionStartTime: time.Now(),
			Address:             "pipe",
			IdleTimeout:         g.opts.IdleTimeout,
		},
		onJoinEvent:  g.sessionManager.join,
		onLeaveEvent: g.sessionManager.leave,
	})
	conn.run()

	// 发送心跳完成入网并消费 0x8001 应答
	if _, err := clientConn.Write(heartbeatPacket); err != nil {
		t.Fatalf("write heartbeat error = %v", err)
	}
	_ = readPacket(t, clientConn, time.Second)
	if key := waitChan(t, events.joined, time.Second); key != "12345678901" {
		t.Fatalf("join key = %s, want 12345678901", key)
	}
	return g, conn, clientConn, events
}

func TestConnection_activeMessageSuccessAndPlatformData(t *testing.T) {
	g, _, clientConn, _ := newPipeConnection(t)

	const key = "12345678901"
	var sentWire []byte
	go func() {
		platformPkt := readPacket(t, clientConn, time.Second)
		sentWire = platformPkt
		platformMsg := decodeFirstMessage(t, platformPkt)
		resp := encodeLocationQueryRespond(t, platformMsg.Header.SerialNumber, 2)
		if _, err := clientConn.Write(resp); err != nil {
			t.Errorf("write 0x0201 error = %v", err)
		}
	}()

	reply := g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, time.Second))
	if reply.ExtensionFields.Err != nil {
		t.Fatalf("SendActiveMessage err = %v", reply.ExtensionFields.Err)
	}
	if reply.Command != consts.T0201QueryLocation {
		t.Fatalf("reply command = %s, want %s", reply.Command, consts.T0201QueryLocation)
	}
	// PlatformData 必须是完整的 0x7e...0x7e 编码帧，可被直接 Decode
	decoded := decodeFirstMessage(t, reply.ExtensionFields.PlatformData)
	if consts.JT808CommandType(decoded.Header.ID) != consts.P8201QueryLocation {
		t.Fatalf("decoded PlatformData ID = 0x%04X, want 0x8201", decoded.Header.ID)
	}
	if !bytes.Equal(reply.ExtensionFields.PlatformData, sentWire) {
		t.Fatalf("PlatformData = %x, want wire %x", reply.ExtensionFields.PlatformData, sentWire)
	}
}

func TestConnection_fastCloseWithoutInflight(t *testing.T) {
	_, conn, clientConn, events := newPipeConnection(t)

	start := time.Now()
	_ = clientConn.Close()
	if key := waitChan(t, events.left, time.Second); key != "12345678901" {
		t.Fatalf("leave key = %s, want 12345678901", key)
	}
	select {
	case <-conn.writeDoneChan:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("write goroutine did not exit promptly after disconnect")
	}
	if elapsed := time.Since(start); elapsed >= 500*time.Millisecond {
		t.Fatalf("close took %v, expected prompt exit without 1s sleep", elapsed)
	}
}

func TestConnection_disconnectWaitsForInflightTimeoutPlanB(t *testing.T) {
	g, conn, clientConn, _ := newPipeConnection(t)

	const key = "12345678901"
	const timeout = 150 * time.Millisecond

	// 终端收到平台下发后立即断开，不回复应答
	go func() {
		_ = readPacket(t, clientConn, time.Second)
		_ = clientConn.Close()
	}()

	start := time.Now()
	reply := g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, timeout))
	elapsed := time.Since(start)

	if !errors.Is(reply.ExtensionFields.Err, ErrWriteDataOverTime) {
		t.Fatalf("err = %v, want %v", reply.ExtensionFields.Err, ErrWriteDataOverTime)
	}
	if elapsed < timeout-30*time.Millisecond {
		t.Fatalf("elapsed = %v, expected to wait for timeout %v under plan B", elapsed, timeout)
	}

	select {
	case <-conn.writeDoneChan:
	case <-time.After(time.Second):
		t.Fatal("write goroutine did not exit after in-flight timeout")
	}
}

func TestConnection_replyArrivedRightBeforeEOF(t *testing.T) {
	g, conn, clientConn, _ := newPipeConnection(t)

	const key = "12345678901"
	go func() {
		platformPkt := readPacket(t, clientConn, time.Second)
		platformMsg := decodeFirstMessage(t, platformPkt)
		resp := encodeLocationQueryRespond(t, platformMsg.Header.SerialNumber, 3)
		_, _ = clientConn.Write(resp)
		_ = clientConn.Close()
	}()

	reply := g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, time.Second))
	if reply.ExtensionFields.Err != nil {
		t.Fatalf("expected reply received before EOF to succeed, got err = %v", reply.ExtensionFields.Err)
	}
	if reply.Command != consts.T0201QueryLocation {
		t.Fatalf("reply command = %s, want %s", reply.Command, consts.T0201QueryLocation)
	}

	select {
	case <-conn.writeDoneChan:
	case <-time.After(time.Second):
		t.Fatal("write goroutine did not exit")
	}
}

func TestConnection_drainPendingActiveMessagesOnDisconnect(t *testing.T) {
	_, conn, clientConn, _ := newPipeConnection(t)
	_ = clientConn.Close()
	// 等待连接完成退出
	select {
	case <-conn.writeDoneChan:
	case <-time.After(time.Second):
		t.Fatal("write goroutine did not exit")
	}

	// 单独测试 drainPendingOnStop：即使 activeMsgChan 中残留未发出的主动消息，也能立即返回 ErrWriteDataFail 而不挂死
	serverConn, peerConn := net.Pipe()
	_ = serverConn.Close()
	_ = peerConn.Close()

	g := New()
	c := newConnection(connectionParams{
		conn:                 serverConn,
		handles:              g.createCommandHandle(),
		activeRespondHandles: g.createActiveRespondHandle(),
		terminalEvent:        &defaultTerminalEvent{},
		filter:               true,
		onJoinEvent:          g.sessionManager.join,
		onLeaveEvent:         g.sessionManager.leave,
	})
	go c.write()

	baseHeader := mustDecodeJTMessage(t, heartbeatPacket).Header
	const pendingCount = 5
	replyChans := make([]chan *Message, pendingCount)
	for i := 0; i < pendingCount; i++ {
		replyChans[i] = make(chan *Message, 1)
		active := NewActiveMessage("12345678901", consts.P8201QueryLocation, nil, time.Second)
		active.header = baseHeader
		active.replyChan = replyChans[i]
		c.activeMsgChan <- active
	}
	close(c.stopWriteChan)

	for i, ch := range replyChans {
		select {
		case msg := <-ch:
			if !errors.Is(msg.ExtensionFields.Err, ErrWriteDataFail) {
				t.Fatalf("pending[%d] err = %v, want %v", i, msg.ExtensionFields.Err, ErrWriteDataFail)
			}
		case <-time.After(time.Second):
			t.Fatalf("pending[%d] blocked forever after disconnect", i)
		}
	}
	select {
	case <-c.writeDoneChan:
	case <-time.After(time.Second):
		t.Fatal("write goroutine did not exit")
	}
}

func TestConnection_manyConcurrentActiveMessagesNoDeadlock(t *testing.T) {
	// 使用默认非阻塞 TerminalEventer，避免 recordingTerminalEvent(cap=32) 阻塞写协程
	g, _, clientConn, events := newPipeConnection(t)
	stopDrain := make(chan struct{})
	defer close(stopDrain)
	go func() {
		for {
			select {
			case <-stopDrain:
				return
			case <-events.readCmd:
			case <-events.writeCmd:
			}
		}
	}()

	const key = "12345678901"
	const total = 24 // 大于通道缓冲 10，验证 write 不会向自身通道发送而自锁

	// net.Pipe 为同步无缓冲通道，测试终端需将读、写拆分为独立协程以模拟全双工 TCP
	respCh := make(chan []byte, total)
	doneReader := make(chan struct{})
	doneWriter := make(chan struct{})
	go func() {
		defer close(doneWriter)
		for resp := range respCh {
			if _, err := clientConn.Write(resp); err != nil {
				t.Errorf("client write error = %v", err)
				return
			}
		}
	}()
	go func() {
		defer close(doneReader)
		defer close(respCh)
		fr := jt808.NewFrameReader()
		buf := make([]byte, 4096)
		var respSeq uint16 = 10
		replied := 0
		for replied < total {
			_ = clientConn.SetReadDeadline(time.Now().Add(3 * time.Second))
			n, err := clientConn.Read(buf)
			if err != nil {
				t.Errorf("client read error = %v (replied %d/%d)", err, replied, total)
				return
			}
			for _, frame := range fr.ReadFrames(buf[:n]) {
				platformMsg := decodeFirstMessage(t, frame)
				respSeq++
				resp := encodeGeneralRespond(t, platformMsg.Header.SerialNumber, consts.JT808CommandType(platformMsg.Header.ID), respSeq)
				respCh <- resp
				replied++
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reply := g.SendActiveMessage(NewActiveMessage(key, consts.P8300TextInfoDistribution, []byte{0x01, 0x31}, 3*time.Second))
			if reply.ExtensionFields.Err != nil {
				t.Errorf("SendActiveMessage err = %v", reply.ExtensionFields.Err)
			}
		}()
	}
	wg.Wait()
	<-doneReader
	<-doneWriter
}

func TestConnection_activeWritePacketsSubcontract(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer func() {
		_ = serverConn.Close()
		_ = clientConn.Close()
	}()

	g := New()
	c := newConnection(connectionParams{
		conn:                 serverConn,
		handles:              g.createCommandHandle(),
		activeRespondHandles: g.createActiveRespondHandle(),
		terminalEvent:        &defaultTerminalEvent{},
		filter:               true,
	})

	baseHeader := mustDecodeJTMessage(t, heartbeatPacket).Header
	origReplyID := baseHeader.ReplyID
	origBodyLen := baseHeader.Property.BodyDayaLen

	body := bytes.Repeat([]byte{0xAB}, 1500) // > 1000，会分成 2 包
	active := NewActiveMessage("12345678901", consts.P8300TextInfoDistribution, body, time.Second)
	active.header = baseHeader

	readDone := make(chan []byte, 1)
	go func() {
		var all []byte
		buf := make([]byte, 4096)
		for len(all) < 1500 {
			_ = clientConn.SetReadDeadline(time.Now().Add(time.Second))
			n, err := clientConn.Read(buf)
			if err != nil {
				break
			}
			all = append(all, buf[:n]...)
		}
		readDone <- all
	}()

	jtMsg, platformData, err := c.activeWritePackets(active)
	if err != nil {
		t.Fatalf("activeWritePackets err = %v", err)
	}
	wire := <-readDone

	if !bytes.Equal(platformData, wire) {
		t.Fatalf("platformData len=%d != wire len=%d", len(platformData), len(wire))
	}
	if jtMsg.Header.SerialNumber != 0 {
		t.Fatalf("first packet serial = %d, want 0", jtMsg.Header.SerialNumber)
	}
	if !bytes.Equal(jtMsg.Body, body) {
		t.Fatalf("jtMsg.Body len = %d, want %d", len(jtMsg.Body), len(body))
	}
	if got := uint16(c.platformSerialNumber.Load()); got != 2 {
		t.Fatalf("next platformSerialNumber = %d, want 2", got)
	}
	// 不得就地修改会话保存的 baseHeader
	if baseHeader.ReplyID != origReplyID || baseHeader.Property.BodyDayaLen != origBodyLen {
		t.Fatalf("baseHeader was mutated: ReplyID=%d BodyDayaLen=%d", baseHeader.ReplyID, baseHeader.Property.BodyDayaLen)
	}
}

func TestGoJT808_RunInvalidAddrReturnsImmediately(t *testing.T) {
	g := New(WithHostPorts("invalid:host:port:99999"))
	done := make(chan struct{})
	go func() {
		g.Run()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run() with invalid addr did not return immediately")
	}
}

func TestConnection_rapidTimeoutVsDisconnectNoPanicNoRace(t *testing.T) {
	// 反复在极短超时与断线并发交错，验证不会触发 closed-channel panic 或数据竞争
	for i := 0; i < 30; i++ {
		g, conn, clientConn, _ := newPipeConnection(t)
		const key = "12345678901"

		go func() {
			_ = readPacket(t, clientConn, time.Second)
			_ = clientConn.Close()
		}()

		reply := g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, 5*time.Millisecond))
		if reply.ExtensionFields.Err == nil {
			t.Fatalf("iter %d: expected timeout or write fail error", i)
		}
		select {
		case <-conn.writeDoneChan:
		case <-time.After(time.Second):
			t.Fatalf("iter %d: write goroutine did not exit", i)
		}
	}
}

func TestConnection_highFreqZeroCopyUplinkNoRace(t *testing.T) {
	// 验证高频 0x0200 位置汇报与 0x0002 心跳在零拷贝路径下并发收发无数据竞争，
	// 且下行回复复用首包独立 Header 不会串号
	g, _, clientConn, events := newPipeConnection(t)
	stopDrain := make(chan struct{})
	defer close(stopDrain)
	go func() {
		for {
			select {
			case <-stopDrain:
				return
			case <-events.readCmd:
			case <-events.writeCmd:
			}
		}
	}()

	const burst = 100
	locBody := make([]byte, 28)
	doneRead := make(chan struct{})
	go func() {
		defer close(doneRead)
		fr := jt808.NewFrameReader()
		buf := make([]byte, 4096)
		got := 0
		for got < burst {
			_ = clientConn.SetReadDeadline(time.Now().Add(3 * time.Second))
			n, err := clientConn.Read(buf)
			if err != nil {
				t.Errorf("client read error = %v (%d/%d)", err, got, burst)
				return
			}
			for _, frame := range fr.ReadFrames(buf[:n]) {
				reply := decodeFirstMessage(t, frame)
				if reply.Header.TerminalPhoneNo != "12345678901" {
					t.Errorf("corrupted phone = %s, want 12345678901", reply.Header.TerminalPhoneNo)
					return
				}
				got++
			}
		}
	}()

	for i := 0; i < burst; i++ {
		pkt := encodeTerminalPacket(t, consts.T0200LocationReport, locBody, uint16(i+1))
		if _, err := clientConn.Write(pkt); err != nil {
			t.Fatalf("write 0x0200[%d] error = %v", i, err)
		}
	}
	<-doneRead
	_ = g
}
