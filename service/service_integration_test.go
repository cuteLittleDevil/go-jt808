package service

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

// recordingHandler 仅绑定单个连接。
// service 每个连接都会重新 createCommandHandle，事件在该连接协程内串行触发，无需加锁。
type recordingHandler struct {
	model.BaseHandle
	protocol consts.JT808CommandType
	reads    int
	writes   int
}

func (r *recordingHandler) Protocol() consts.JT808CommandType { return r.protocol }

func (r *recordingHandler) OnReadExecutionEvent(_ *Message) { r.reads++ }

func (r *recordingHandler) OnWriteExecutionEvent(_ Message) { r.writes++ }

// recordingTerminalEvent 用 channel 把事件交给测试协程。
// 单连接场景下可复用同一实例；跨协程同步靠 channel，不靠 mutex。
type recordingTerminalEvent struct {
	joined       chan string
	left         chan string
	notSupported chan consts.JT808CommandType
	readCmd      chan consts.JT808CommandType
	writeCmd     chan consts.JT808CommandType
}

func newRecordingTerminalEvent() *recordingTerminalEvent {
	return &recordingTerminalEvent{
		joined:       make(chan string, 8),
		left:         make(chan string, 8),
		notSupported: make(chan consts.JT808CommandType, 8),
		readCmd:      make(chan consts.JT808CommandType, 32),
		writeCmd:     make(chan consts.JT808CommandType, 32),
	}
}

func (r *recordingTerminalEvent) OnJoinEvent(_ *Message, key string, _ error) {
	r.joined <- key
}

func (r *recordingTerminalEvent) OnLeaveEvent(key string) {
	r.left <- key
}

func (r *recordingTerminalEvent) OnNotSupportedEvent(msg *Message) {
	r.notSupported <- msg.Command
}

func (r *recordingTerminalEvent) OnReadExecutionEvent(msg *Message) {
	r.readCmd <- msg.Command
}

func (r *recordingTerminalEvent) OnWriteExecutionEvent(msg Message) {
	r.writeCmd <- msg.Command
}

func waitChan[T any](t *testing.T, ch <-chan T, timeout time.Duration) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(timeout):
		t.Fatal("timed out waiting for event")
		var zero T
		return zero
	}
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

func startTestServer(t *testing.T, opts ...Option) (*GoJT808, string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	all := append([]Option{WithHostPorts(addr)}, opts...)
	g := New(all...)
	go g.Run()

	waitFor(t, 2*time.Second, func() bool {
		c, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err != nil {
			return false
		}
		_ = c.Close()
		return true
	})
	// 探测连接会触发 leave（key 可能为空），稍等并排空，避免干扰后续断言
	time.Sleep(50 * time.Millisecond)
	return g, addr
}

func dialTerminal(t *testing.T, addr string) net.Conn {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func readPacket(t *testing.T, conn net.Conn, timeout time.Duration) []byte {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	return append([]byte(nil), buf[:n]...)
}

func encodeTerminalPacket(t *testing.T, command consts.JT808CommandType, body []byte, serial uint16) []byte {
	t.Helper()
	base := mustDecodeJTMessage(t, heartbeatPacket)
	header := base.Header
	// 复用心跳包里的 BCD 手机号；测试统一使用 12345678901
	header.ReplyID = uint16(command)
	header.PlatformSerialNumber = serial
	return header.Encode(body)
}

func encodeGeneralRespond(t *testing.T, platformSeq uint16, platformCmd consts.JT808CommandType, serial uint16) []byte {
	t.Helper()
	body := (&model.T0x0001{
		SerialNumber: platformSeq,
		ID:           uint16(platformCmd),
		Result:       0,
	}).Encode()
	return encodeTerminalPacket(t, consts.T0001GeneralRespond, body, serial)
}

func encodeLocationQueryRespond(t *testing.T, platformSeq uint16, serial uint16) []byte {
	t.Helper()
	loc := model.T0x0201{
		RespondSerialNumber: platformSeq,
		T0x0200LocationItem: model.T0x0200LocationItem{
			Latitude:  31000000,
			Longitude: 121000000,
			Altitude:  10,
			Speed:     60,
			Direction: 90,
			DateTime:  "250101120000",
		},
	}
	return encodeTerminalPacket(t, consts.T0201QueryLocation, loc.Encode(), serial)
}

func decodeFirstMessage(t *testing.T, data []byte) *jt808.JTMessage {
	t.Helper()
	jtMsg := jt808.NewJTMessage()
	if err := jtMsg.Decode(data); err != nil {
		t.Fatalf("Decode platform packet error = %v, data=%x", err, data)
	}
	return jtMsg
}

func drainStringChan(ch <-chan string) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

func TestService_heartbeatJoinAndDefaultReply(t *testing.T) {
	events := newRecordingTerminalEvent()
	g, addr := startTestServer(t,
		WithCustomTerminalEventer(func() TerminalEventer { return events }),
		WithCustomHandleFunc(func() map[consts.JT808CommandType]Handler {
			// 每个连接独立一份 Handler
			return map[consts.JT808CommandType]Handler{
				consts.T0002HeartBeat: &recordingHandler{protocol: consts.T0002HeartBeat},
			}
		}),
	)
	_ = g
	drainStringChan(events.left)
	drainStringChan(events.joined)

	conn := dialTerminal(t, addr)
	if _, err := conn.Write(heartbeatPacket); err != nil {
		t.Fatalf("Write heartbeat error = %v", err)
	}

	reply := readPacket(t, conn, 2*time.Second)
	jtMsg := decodeFirstMessage(t, reply)
	if consts.JT808CommandType(jtMsg.Header.ID) != consts.P8001GeneralRespond {
		t.Fatalf("reply command = 0x%04X, want 0x8001", jtMsg.Header.ID)
	}

	if key := waitChan(t, events.joined, 2*time.Second); key != "12345678901" {
		t.Fatalf("join key = %s, want 12345678901", key)
	}
	if cmd := waitChan(t, events.readCmd, 2*time.Second); cmd != consts.T0002HeartBeat {
		t.Fatalf("read command = %s, want %s", cmd, consts.T0002HeartBeat)
	}
}

func TestService_sendActiveMessageSuccessAndTimeout(t *testing.T) {
	events := newRecordingTerminalEvent()
	g, addr := startTestServer(t,
		WithCustomTerminalEventer(func() TerminalEventer { return events }),
	)
	drainStringChan(events.left)
	drainStringChan(events.joined)

	conn := dialTerminal(t, addr)
	if _, err := conn.Write(heartbeatPacket); err != nil {
		t.Fatalf("Write heartbeat error = %v", err)
	}
	_ = readPacket(t, conn, 2*time.Second) // 消费 0x8001
	_ = waitChan(t, events.joined, 2*time.Second)

	key := "12345678901"

	go func() {
		platformPkt := readPacket(t, conn, 2*time.Second)
		platformMsg := decodeFirstMessage(t, platformPkt)
		if consts.JT808CommandType(platformMsg.Header.ID) != consts.P8201QueryLocation {
			t.Errorf("platform command = 0x%04X, want 0x8201", platformMsg.Header.ID)
			return
		}
		seq := platformMsg.Header.SerialNumber
		resp := encodeLocationQueryRespond(t, seq, 2)
		if _, err := conn.Write(resp); err != nil {
			t.Errorf("Write 0x0201 error = %v", err)
		}
	}()

	reply := g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, time.Second))
	if reply.ExtensionFields.Err != nil {
		t.Fatalf("SendActiveMessage success path err = %v", reply.ExtensionFields.Err)
	}
	if reply.Command != consts.T0201QueryLocation {
		t.Fatalf("reply command = %s, want %s", reply.Command, consts.T0201QueryLocation)
	}
	decodedPlatform := decodeFirstMessage(t, reply.ExtensionFields.PlatformData)
	if consts.JT808CommandType(decodedPlatform.Header.ID) != consts.P8201QueryLocation {
		t.Fatalf("PlatformData command = 0x%04X, want 0x8201", decodedPlatform.Header.ID)
	}

	// 分包主动下发(>1000B)：终端按最后一包流水号应答，应成功匹配
	subBody := bytes.Repeat([]byte{0x31}, 1500)
	go func() {
		fr := jt808.NewFrameReader()
		var frames [][]byte
		for len(frames) < 2 {
			pkt := readPacket(t, conn, 2*time.Second)
			frames = append(frames, fr.ReadFrames(pkt)...)
		}
		lastMsg := decodeFirstMessage(t, frames[len(frames)-1])
		resp := encodeGeneralRespond(t, lastMsg.Header.SerialNumber, consts.P8300TextInfoDistribution, 6)
		_, _ = conn.Write(resp)
	}()
	subReply := g.SendActiveMessage(NewActiveMessage(key, consts.P8300TextInfoDistribution, subBody, time.Second))
	if subReply.ExtensionFields.Err != nil {
		t.Fatalf("subcontract SendActiveMessage err = %v", subReply.ExtensionFields.Err)
	}

	go func() {
		_ = readPacket(t, conn, 2*time.Second) // 消费平台下发报文
	}()
	timeoutReply := g.SendActiveMessage(NewActiveMessage(key, consts.P8300TextInfoDistribution, []byte{0x01, 0x31, 0x32, 0x33}, 200*time.Millisecond))
	if !errors.Is(timeoutReply.ExtensionFields.Err, ErrWriteDataOverTime) {
		t.Fatalf("timeout err = %v, want %v", timeoutReply.ExtensionFields.Err, ErrWriteDataOverTime)
	}
}

func TestService_notSupportedCommand(t *testing.T) {
	events := newRecordingTerminalEvent()
	g, addr := startTestServer(t,
		WithCustomTerminalEventer(func() TerminalEventer { return events }),
	)
	_ = g
	drainStringChan(events.left)
	drainStringChan(events.joined)

	conn := dialTerminal(t, addr)
	if _, err := conn.Write(heartbeatPacket); err != nil {
		t.Fatalf("Write heartbeat error = %v", err)
	}
	_ = readPacket(t, conn, 2*time.Second)
	_ = waitChan(t, events.joined, 2*time.Second)

	unknown := encodeTerminalPacket(t, consts.JT808CommandType(0x0F00), []byte{0x01}, 3)
	if _, err := conn.Write(unknown); err != nil {
		t.Fatalf("Write unknown error = %v", err)
	}

	if got := waitChan(t, events.notSupported, 2*time.Second); got != consts.JT808CommandType(0x0F00) {
		t.Fatalf("notSupported = %s, want 0x0F00", got)
	}
}

func TestService_defaultHandlersIncludeNewlyRegisteredCommands(t *testing.T) {
	events := newRecordingTerminalEvent()
	g, addr := startTestServer(t,
		WithCustomTerminalEventer(func() TerminalEventer { return events }),
	)
	drainStringChan(events.left)
	drainStringChan(events.joined)

	conn := dialTerminal(t, addr)
	if _, err := conn.Write(heartbeatPacket); err != nil {
		t.Fatalf("Write heartbeat error = %v", err)
	}
	_ = readPacket(t, conn, 2*time.Second)
	_ = waitChan(t, events.joined, 2*time.Second)

	key := "12345678901"
	body := (&model.P0x9105{ChannelNo: 1, PackageLossRate: 2}).Encode()

	go func() {
		platformPkt := readPacket(t, conn, 2*time.Second)
		platformMsg := decodeFirstMessage(t, platformPkt)
		if consts.JT808CommandType(platformMsg.Header.ID) != consts.P9105AudioVideoControlStatusNotice {
			t.Errorf("platform command = 0x%04X, want 0x9105", platformMsg.Header.ID)
			return
		}
		seq := platformMsg.Header.SerialNumber
		resp := encodeGeneralRespond(t, seq, consts.P9105AudioVideoControlStatusNotice, 5)
		if _, err := conn.Write(resp); err != nil {
			t.Errorf("Write 0x0001 error = %v", err)
		}
	}()

	reply := g.SendActiveMessage(NewActiveMessage(key, consts.P9105AudioVideoControlStatusNotice, body, time.Second))
	if reply.ExtensionFields.Err != nil {
		t.Fatalf("P9105 SendActiveMessage err = %v", reply.ExtensionFields.Err)
	}
	if reply.Command != consts.T0001GeneralRespond {
		t.Fatalf("reply command = %s, want %s", reply.Command, consts.T0001GeneralRespond)
	}

	select {
	case cmd := <-events.notSupported:
		t.Fatalf("unexpected notSupported event: %s", cmd)
	default:
	}
}

func TestPackageParse_completeSubcontract(t *testing.T) {
	base := mustDecodeJTMessage(t, heartbeatPacket)
	header := base.Header
	header.ReplyID = uint16(consts.T0801MultimediaDataUpload)
	header.PlatformSerialNumber = 1

	body := make([]byte, 1500)
	for i := range body {
		body[i] = byte(i)
	}
	packets := header.EncodePackets(body)
	if len(packets) < 2 {
		t.Fatalf("expected subcontract packets, got %d", len(packets))
	}

	p := newPackageParse()
	var complete *Message
	for i, pkt := range packets {
		msgs, err := p.parse(pkt)
		if err != nil {
			t.Fatalf("parse packet[%d] error = %v", i, err)
		}
		for _, msg := range msgs {
			if msg.ExtensionFields.SubcontractComplete {
				complete = msg
			}
		}
	}
	if complete == nil {
		t.Fatal("expected completed subcontract message")
	}
	if len(complete.Body) != len(body) {
		t.Fatalf("complete body len = %d, want %d", len(complete.Body), len(body))
	}
	if complete.Body[0] != body[0] || complete.Body[len(body)-1] != body[len(body)-1] {
		t.Fatal("complete body content mismatch")
	}
}

func TestMakeSerialMatchHandler(t *testing.T) {
	g := New()
	match := g.makeSerialMatchHandler(func(jtMsg *jt808.JTMessage) (uint16, error) {
		return binary.BigEndian.Uint16(jtMsg.Body[:2]), nil
	})

	active := &ActiveMessage{Command: consts.P8201QueryLocation}
	active.ExtensionFields.PlatformSeq = 7

	body := make([]byte, 2)
	binary.BigEndian.PutUint16(body, 7)
	terminal := &Message{
		JTMessage: &jt808.JTMessage{Body: body, Header: &jt808.Header{}},
		Command:   consts.T0201QueryLocation,
	}
	if !match(active, terminal) {
		t.Fatal("expected serial match")
	}

	binary.BigEndian.PutUint16(body, 8)
	if match(active, terminal) {
		t.Fatal("expected serial mismatch")
	}
}

func newPipeTestConnection(t *testing.T) (*GoJT808, *connection, net.Conn, *recordingTerminalEvent) {
	t.Helper()
	events := newRecordingTerminalEvent()
	g := New(WithCustomTerminalEventer(func() TerminalEventer { return events }))
	serverConn, clientConn := net.Pipe()
	t.Cleanup(func() {
		_ = clientConn.Close()
		_ = serverConn.Close()
	})
	conn := newConnection(connectionParams{
		conn:                 serverConn,
		handles:              g.createCommandHandle(),
		activeRespondHandles: g.createActiveRespondHandle(),
		terminalEvent:        events,
		filter:               true,
		maxFrameBufferSize:   g.opts.MaxFrameBufferSize,
		onJoinEvent:          g.sessionManager.join,
		onLeaveEvent:         g.sessionManager.leave,
	})
	conn.run()

	if _, err := clientConn.Write(heartbeatPacket); err != nil {
		t.Fatalf("write heartbeat error = %v", err)
	}
	_ = readPacket(t, clientConn, time.Second)
	_ = waitChan(t, events.joined, time.Second)
	_ = waitChan(t, events.readCmd, time.Second)
	_ = waitChan(t, events.writeCmd, time.Second)
	return g, conn, clientConn, events
}

func TestConnection_activeTimerLifecycle(t *testing.T) {
	const key = "12345678901"

	t.Run("来源1_终端按时应答Stop取消定时器且不二次触发超时", func(t *testing.T) {
		g, _, clientConn, events := newPipeTestConnection(t)
		go func() {
			pkt := readPacket(t, clientConn, time.Second)
			msg := decodeFirstMessage(t, pkt)
			resp := encodeLocationQueryRespond(t, msg.Header.SerialNumber, 2)
			_, _ = clientConn.Write(resp)
		}()

		reply := g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, 80*time.Millisecond))
		if reply.ExtensionFields.Err != nil {
			t.Fatalf("SendActiveMessage err = %v", reply.ExtensionFields.Err)
		}
		// 消费成功完成时的两次写事件（0x0201 与 0x8201）
		_ = waitChan(t, events.writeCmd, time.Second)
		_ = waitChan(t, events.writeCmd, time.Second)

		// 等待超过原定 80ms 超时窗口，确认定时器已被 Stop，不会再触发超时写事件
		time.Sleep(120 * time.Millisecond)
		select {
		case cmd := <-events.writeCmd:
			t.Fatalf("unexpected late timeout write event: %s", cmd)
		default:
		}
	})

	t.Run("来源2_发包后断线等待在途超时_未发请求立即失败_断线前应答正常完成", func(t *testing.T) {
		// 2a: 发包后终端立即断线 -> 等待在途超时返回 ErrWriteDataOverTime
		g, _, clientConn, _ := newPipeTestConnection(t)
		const timeout = 120 * time.Millisecond
		go func() {
			_ = readPacket(t, clientConn, time.Second)
			_ = clientConn.Close()
		}()
		start := time.Now()
		reply := g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, timeout))
		if !errors.Is(reply.ExtensionFields.Err, ErrWriteDataOverTime) {
			t.Fatalf("err = %v, want %v", reply.ExtensionFields.Err, ErrWriteDataOverTime)
		}
		if elapsed := time.Since(start); elapsed < timeout-30*time.Millisecond {
			t.Fatalf("elapsed = %v, expected wait ~%v for in-flight timeout", elapsed, timeout)
		}

		// 2b: 终端回包后紧接着 EOF -> 应答被 drainBeforeExit 优先匹配，不误判超时
		g2, _, clientConn2, _ := newPipeTestConnection(t)
		go func() {
			pkt := readPacket(t, clientConn2, time.Second)
			msg := decodeFirstMessage(t, pkt)
			_, _ = clientConn2.Write(encodeLocationQueryRespond(t, msg.Header.SerialNumber, 3))
			_ = clientConn2.Close()
		}()
		reply2 := g2.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, time.Second))
		if reply2.ExtensionFields.Err != nil {
			t.Fatalf("reply before EOF failed: %v", reply2.ExtensionFields.Err)
		}

		// 2c: 无在途请求时普通连接断线应立即关闭（不再固定 sleep 1s）
		_, conn3, clientConn3, events3 := newPipeTestConnection(t)
		closeStart := time.Now()
		_ = clientConn3.Close()
		_ = waitChan(t, events3.left, time.Second)
		select {
		case <-conn3.finallyCompleteChan:
		case <-time.After(300 * time.Millisecond):
			t.Fatal("finallyCompleteChan did not close promptly")
		}
		if elapsed := time.Since(closeStart); elapsed >= 500*time.Millisecond {
			t.Fatalf("disconnect took %v, expected prompt close without 1s sleep", elapsed)
		}
	})

	t.Run("来源3_流水号回绕冲突停止旧请求且旧定时器不误杀新请求", func(t *testing.T) {
		g, conn, clientConn, _ := newPipeTestConnection(t)

		// 第 1 条请求发在流水号 1（设短超时 60ms），终端只收不回
		firstRes := make(chan *Message, 1)
		firstSent := make(chan struct{})
		go func() {
			_ = readPacket(t, clientConn, time.Second)
			close(firstSent)
		}()
		go func() {
			firstRes <- g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, 60*time.Millisecond))
		}()
		<-firstSent

		// 模拟跑满 65536 次后流水号回绕回 1，并发出第 2 条同流水号请求（超时 500ms）
		conn.platformSerialNumber.Store(1)
		go func() {
			pkt := readPacket(t, clientConn, time.Second)
			msg := decodeFirstMessage(t, pkt)
			// 故意等超过第 1 条的 60ms 超时时间后再回包，验证旧定时器不会误杀第 2 条请求
			time.Sleep(100 * time.Millisecond)
			_, _ = clientConn.Write(encodeLocationQueryRespond(t, msg.Header.SerialNumber, 9))
		}()

		secondReply := g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, 500*time.Millisecond))
		if secondReply.ExtensionFields.Err != nil {
			t.Fatalf("new request after seq wraparound failed: %v", secondReply.ExtensionFields.Err)
		}

		select {
		case oldReply := <-firstRes:
			if !errors.Is(oldReply.ExtensionFields.Err, ErrWriteDataOverTime) {
				t.Fatalf("collided old request err = %v, want %v", oldReply.ExtensionFields.Err, ErrWriteDataOverTime)
			}
		case <-time.After(time.Second):
			t.Fatal("collided old request leaked without returning")
		}
	})

	t.Run("来源4_高频并发超时与断线退出done兜底无Panic无死锁", func(t *testing.T) {
		for i := 0; i < 15; i++ {
			g, _, clientConn, _ := newPipeTestConnection(t)
			go func() {
				buf := make([]byte, 2048)
				for {
					if _, err := clientConn.Read(buf); err != nil {
						return
					}
					_ = clientConn.Close()
					return
				}
			}()

			var wg sync.WaitGroup
			for j := 0; j < 16; j++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					res := g.SendActiveMessage(NewActiveMessage(key, consts.P8300TextInfoDistribution, []byte{0x01, 0x31}, 5*time.Millisecond))
					if res.ExtensionFields.Err == nil {
						t.Errorf("expected error on disconnected/timed-out active message")
					}
				}()
			}
			wg.Wait()
		}
	})
}

func TestConnection_zeroCopyHighFreqUplinkAndActiveNoRace(t *testing.T) {
	g, _, clientConn, events := newPipeTestConnection(t)
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

	const burst = 80
	doneRead := make(chan struct{})
	go func() {
		defer close(doneRead)
		fr := jt808.NewFrameReader()
		buf := make([]byte, 4096)
		locReplies := 0
		for locReplies < burst {
			_ = clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, err := clientConn.Read(buf)
			if err != nil {
				t.Errorf("client read err = %v", err)
				return
			}
			for _, frame := range fr.ReadFrames(buf[:n]) {
				msg := decodeFirstMessage(t, frame)
				if msg.Header.TerminalPhoneNo != "12345678901" {
					t.Errorf("phone = %s, want 12345678901", msg.Header.TerminalPhoneNo)
					return
				}
				if consts.JT808CommandType(msg.Header.ID) == consts.P8201QueryLocation {
					_, _ = clientConn.Write(encodeLocationQueryRespond(t, msg.Header.SerialNumber, 99))
				} else {
					locReplies++
				}
			}
		}
	}()

	activeDone := make(chan error, 1)
	go func() {
		res := g.SendActiveMessage(NewActiveMessage("12345678901", consts.P8201QueryLocation, nil, 2*time.Second))
		activeDone <- res.ExtensionFields.Err
	}()

	locBody := make([]byte, 28)
	for i := 0; i < burst; i++ {
		pkt := encodeTerminalPacket(t, consts.T0200LocationReport, locBody, uint16(i+1))
		if _, err := clientConn.Write(pkt); err != nil {
			t.Fatalf("write 0x0200[%d] err = %v", i, err)
		}
	}
	if err := <-activeDone; err != nil {
		t.Fatalf("concurrent SendActiveMessage err = %v", err)
	}
	<-doneRead
}

func TestConnection_maxFrameBufferSize(t *testing.T) {
	t.Run("默认1023x2阈值_7e未闭合数据达到上限断开连接", func(t *testing.T) {
		g := New()
		if g.opts.MaxFrameBufferSize != 1023*2 {
			t.Fatalf("default MaxFrameBufferSize = %d, want %d", g.opts.MaxFrameBufferSize, 1023*2)
		}

		_, conn, clientConn, events := newPipeTestConnection(t)
		// 先分两次写入 0x7e + 1999 字节（共 2000 < 2046），不应断开
		part1 := append([]byte{0x7e}, bytes.Repeat([]byte{0x01}, 999)...)
		part2 := bytes.Repeat([]byte{0x02}, 1000)
		if _, err := clientConn.Write(part1); err != nil {
			t.Fatalf("write part1 err = %v", err)
		}
		if _, err := clientConn.Write(part2); err != nil {
			t.Fatalf("write part2 err = %v", err)
		}
		select {
		case <-conn.finallyCompleteChan:
			t.Fatal("connection closed prematurely before reaching 1023*2")
		case <-time.After(40 * time.Millisecond):
		}

		// 再补齐 46 字节达到 1023*2（2000 + 46 = 2046），达到上限后应立即断开连接
		part3 := bytes.Repeat([]byte{0x03}, 1023*2-len(part1)-len(part2))
		if _, err := clientConn.Write(part3); err != nil {
			t.Fatalf("write part3 err = %v", err)
		}

		_ = waitChan(t, events.left, time.Second)
		select {
		case <-conn.finallyCompleteChan:
		case <-time.After(time.Second):
			t.Fatal("connection did not close after frame buffer reached 1023*2")
		}

		// 客户端侧读取应收到 EOF
		_ = clientConn.SetReadDeadline(time.Now().Add(time.Second))
		buf := make([]byte, 16)
		if _, err := clientConn.Read(buf); !errors.Is(err, io.EOF) {
			t.Fatalf("client read after overflow err = %v, want io.EOF", err)
		}
	})

	t.Run("自定义WithMaxFrameBufferSize与禁用上限", func(t *testing.T) {
		// 1. 自定义较小阈值（64B）：非 0x7e 开头的垃圾数据达到阈值即断开
		events := newRecordingTerminalEvent()
		g := New(
			WithMaxFrameBufferSize(64),
			WithCustomTerminalEventer(func() TerminalEventer { return events }),
		)
		if g.opts.MaxFrameBufferSize != 64 {
			t.Fatalf("custom MaxFrameBufferSize = %d, want 64", g.opts.MaxFrameBufferSize)
		}
		serverConn, clientConn := net.Pipe()
		t.Cleanup(func() {
			_ = clientConn.Close()
			_ = serverConn.Close()
		})
		conn := newConnection(connectionParams{
			conn:                 serverConn,
			handles:              g.createCommandHandle(),
			activeRespondHandles: g.createActiveRespondHandle(),
			terminalEvent:        events,
			filter:               true,
			maxFrameBufferSize:   g.opts.MaxFrameBufferSize,
			onJoinEvent:          g.sessionManager.join,
			onLeaveEvent:         g.sessionManager.leave,
		})
		conn.run()

		garbage := bytes.Repeat([]byte("A"), 64)
		if _, err := clientConn.Write(garbage); err != nil {
			t.Fatalf("write garbage err = %v", err)
		}
		select {
		case <-conn.finallyCompleteChan:
		case <-time.After(time.Second):
			t.Fatal("connection did not close on custom 64B frame buffer overflow")
		}

		// 2. WithMaxFrameBufferSize(0) 禁用限制：超过 1023*2 仍保持连接不断开
		gUnlimited := New(WithMaxFrameBufferSize(0))
		if gUnlimited.opts.MaxFrameBufferSize != 0 {
			t.Fatalf("unlimited MaxFrameBufferSize = %d, want 0", gUnlimited.opts.MaxFrameBufferSize)
		}
		serverConn2, clientConn2 := net.Pipe()
		t.Cleanup(func() {
			_ = clientConn2.Close()
			_ = serverConn2.Close()
		})
		conn2 := newConnection(connectionParams{
			conn:                 serverConn2,
			handles:              gUnlimited.createCommandHandle(),
			activeRespondHandles: gUnlimited.createActiveRespondHandle(),
			terminalEvent:        newRecordingTerminalEvent(),
			filter:               true,
			maxFrameBufferSize:   gUnlimited.opts.MaxFrameBufferSize,
			onJoinEvent:          gUnlimited.sessionManager.join,
			onLeaveEvent:         gUnlimited.sessionManager.leave,
		})
		conn2.run()

		unclosed := append([]byte{0x7e}, bytes.Repeat([]byte{0x01}, 1023*2+100)...)
		if _, err := clientConn2.Write(unclosed); err != nil {
			t.Fatalf("write unclosed err = %v", err)
		}
		select {
		case <-conn2.finallyCompleteChan:
			t.Fatal("connection closed unexpectedly when MaxFrameBufferSize=0")
		case <-time.After(50 * time.Millisecond):
		}
	})
}

func TestConnection_writeTimeoutClosesStuckConnection(t *testing.T) {
	const key = "12345678901"

	t.Run("上行应答写超时自动断开连接并清理会话", func(t *testing.T) {
		events := newRecordingTerminalEvent()
		g := New(WithCustomTerminalEventer(func() TerminalEventer { return events }))
		serverConn, clientConn := net.Pipe()
		t.Cleanup(func() {
			_ = clientConn.Close()
			_ = serverConn.Close()
		})
		conn := newConnection(connectionParams{
			conn:                 serverConn,
			handles:              g.createCommandHandle(),
			activeRespondHandles: g.createActiveRespondHandle(),
			terminalEvent:        events,
			filter:               true,
			maxFrameBufferSize:   g.opts.MaxFrameBufferSize,
			writeTimeout:         50 * time.Millisecond,
			onJoinEvent:          g.sessionManager.join,
			onLeaveEvent:         g.sessionManager.leave,
		})
		conn.run()

		// 首次心跳正常读取应答完成上线
		if _, err := clientConn.Write(heartbeatPacket); err != nil {
			t.Fatalf("write heartbeat err = %v", err)
		}
		_ = readPacket(t, clientConn, time.Second)
		_ = waitChan(t, events.joined, time.Second)

		// 第二次心跳后客户端不再读取，迫使服务端 writeData 触发写超时并关闭连接
		if _, err := clientConn.Write(heartbeatPacket); err != nil {
			t.Fatalf("write second heartbeat err = %v", err)
		}

		_ = waitChan(t, events.left, time.Second)
		select {
		case <-conn.finallyCompleteChan:
		case <-time.After(time.Second):
			t.Fatal("connection did not close after write timeout")
		}

		reply := g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, 100*time.Millisecond))
		if !errors.Is(reply.ExtensionFields.Err, ErrNotExistKey) {
			t.Fatalf("after write timeout SendActiveMessage err = %v, want %v", reply.ExtensionFields.Err, ErrNotExistKey)
		}
	})

	t.Run("主动下发写超时返回ErrWriteDataFail并断开连接", func(t *testing.T) {
		events := newRecordingTerminalEvent()
		g := New(WithCustomTerminalEventer(func() TerminalEventer { return events }))
		serverConn, clientConn := net.Pipe()
		t.Cleanup(func() {
			_ = clientConn.Close()
			_ = serverConn.Close()
		})
		conn := newConnection(connectionParams{
			conn:                 serverConn,
			handles:              g.createCommandHandle(),
			activeRespondHandles: g.createActiveRespondHandle(),
			terminalEvent:        events,
			filter:               true,
			maxFrameBufferSize:   g.opts.MaxFrameBufferSize,
			writeTimeout:         50 * time.Millisecond,
			onJoinEvent:          g.sessionManager.join,
			onLeaveEvent:         g.sessionManager.leave,
		})
		conn.run()

		if _, err := clientConn.Write(heartbeatPacket); err != nil {
			t.Fatalf("write heartbeat err = %v", err)
		}
		_ = readPacket(t, clientConn, time.Second)
		_ = waitChan(t, events.joined, time.Second)

		// 客户端不读取主动下发报文，触发 50ms 写超时
		reply := g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, time.Second))
		if !errors.Is(reply.ExtensionFields.Err, ErrWriteDataFail) || !errors.Is(reply.ExtensionFields.Err, os.ErrDeadlineExceeded) {
			t.Fatalf("active write timeout err = %v, want ErrWriteDataFail + os.ErrDeadlineExceeded", reply.ExtensionFields.Err)
		}

		_ = waitChan(t, events.left, time.Second)
		select {
		case <-conn.finallyCompleteChan:
		case <-time.After(time.Second):
			t.Fatal("connection did not close after active write timeout")
		}
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
func (l *stubListener) Addr() net.Addr { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080} }

func TestService_serveReturnsOnErrClosedAndAppliesMaxFrameBufferSize(t *testing.T) {
	events := newRecordingTerminalEvent()
	g := New(
		WithMaxFrameBufferSize(32),
		WithCustomTerminalEventer(func() TerminalEventer { return events }),
	)

	serverConn, clientConn := net.Pipe()
	t.Cleanup(func() {
		_ = clientConn.Close()
		_ = serverConn.Close()
	})

	ln := &stubListener{steps: make(chan func() (net.Conn, error), 3)}
	// 1. 临时错误：退避后继续循环
	ln.steps <- func() (net.Conn, error) {
		return nil, errors.New("temporary accept error")
	}
	// 2. 正常接入一条连接（用于验证 serve 将 MaxFrameBufferSize 传入连接）
	ln.steps <- func() (net.Conn, error) {
		return serverConn, nil
	}
	// 3. 监听器关闭：serve 应立即退出而不死循环
	ln.steps <- func() (net.Conn, error) {
		return nil, net.ErrClosed
	}

	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		g.serve(ln)
	}()

	// 验证接入的连接正常工作且继承了 MaxFrameBufferSize=32
	if _, err := clientConn.Write(heartbeatPacket); err != nil {
		t.Fatalf("write heartbeat err = %v", err)
	}
	_ = readPacket(t, clientConn, time.Second)
	_ = waitChan(t, events.joined, time.Second)

	// 发送 32 字节未闭合帧，触发溢出断连
	if _, err := clientConn.Write(append([]byte{0x7e}, bytes.Repeat([]byte{0x01}, 31)...)); err != nil {
		t.Fatalf("write overflow err = %v", err)
	}
	_ = waitChan(t, events.left, time.Second)

	select {
	case <-serveDone:
	case <-time.After(time.Second):
		t.Fatal("g.serve did not return upon net.ErrClosed")
	}
}
