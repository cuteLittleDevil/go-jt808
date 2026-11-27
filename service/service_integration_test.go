package service

import (
	"bytes"
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

type recordingHandler struct {
	model.BaseHandle
	protocol  consts.JT808CommandType
	replyBody []byte
	replyErr  error
	reads     int
	writes    int
}

func (r *recordingHandler) Protocol() consts.JT808CommandType { return r.protocol }
func (r *recordingHandler) OnReadExecutionEvent(_ *Message)   { r.reads++ }
func (r *recordingHandler) OnWriteExecutionEvent(_ Message)   { r.writes++ }
func (r *recordingHandler) ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error) {
	if r.replyErr != nil || r.replyBody != nil {
		return r.replyBody, r.replyErr
	}
	return r.BaseHandle.ReplyBody(jtMsg)
}

type recordingTerminalEvent struct {
	joined       chan string
	left         chan string
	notSupported chan consts.JT808CommandType
	readCmd      chan consts.JT808CommandType
	writeCmd     chan consts.JT808CommandType
}

func newRecordingTerminalEvent() *recordingTerminalEvent {
	return &recordingTerminalEvent{
		joined:       make(chan string, 16),
		left:         make(chan string, 16),
		notSupported: make(chan consts.JT808CommandType, 16),
		readCmd:      make(chan consts.JT808CommandType, 64),
		writeCmd:     make(chan consts.JT808CommandType, 64),
	}
}

func (r *recordingTerminalEvent) OnJoinEvent(_ *Message, key string, _ error) { r.joined <- key }
func (r *recordingTerminalEvent) OnLeaveEvent(key string)                     { r.left <- key }
func (r *recordingTerminalEvent) OnNotSupportedEvent(msg *Message)            { r.notSupported <- msg.Command }
func (r *recordingTerminalEvent) OnReadExecutionEvent(msg *Message)           { r.readCmd <- msg.Command }
func (r *recordingTerminalEvent) OnWriteExecutionEvent(msg Message)           { r.writeCmd <- msg.Command }

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
	header.ReplyID = uint16(command)
	header.PlatformSerialNumber = serial
	return header.Encode(body)
}

func encodeGeneralRespond(t *testing.T, platformSeq uint16, platformCmd consts.JT808CommandType, serial uint16) []byte {
	t.Helper()
	body := (&model.T0x0001{SerialNumber: platformSeq, ID: uint16(platformCmd), Result: 0}).Encode()
	return encodeTerminalPacket(t, consts.T0001GeneralRespond, body, serial)
}

func encodeLocationQueryRespond(t *testing.T, platformSeq uint16, serial uint16) []byte {
	t.Helper()
	loc := model.T0x0201{
		RespondSerialNumber: platformSeq,
		T0x0200LocationItem: model.T0x0200LocationItem{DateTime: "250101120000"},
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

func startRawPipeConn(t *testing.T, custom func(*connectionParams), opts ...Option) (*GoJT808, *connection, net.Conn, *recordingTerminalEvent) {
	t.Helper()
	events := newRecordingTerminalEvent()
	allOpts := append([]Option{WithCustomTerminalEventer(func() TerminalEventer { return events })}, opts...)
	g := New(allOpts...)
	serverConn, clientConn := net.Pipe()
	t.Cleanup(func() {
		_ = clientConn.Close()
		_ = serverConn.Close()
	})
	params := connectionParams{
		conn:                   serverConn,
		handles:                g.createCommandHandle(),
		activeRespondHandles:   g.createActiveRespondHandle(),
		terminalEvent:          events,
		filter:                 g.opts.FilterSubcontract,
		maxFrameBufferSize:     g.opts.MaxFrameBufferSize,
		onTerminalTimeoutEvent: g.opts.OnTerminalTimeoutEvent,
		timeout: TerminalTimeout{
			ConnectionStartTime: time.Now(),
			Address:             "pipe",
			IdleTimeout:         g.opts.IdleTimeout,
		},
		onJoinEvent:  g.sessionManager.join,
		onLeaveEvent: g.sessionManager.leave,
	}
	if custom != nil {
		custom(&params)
	}
	conn := newConnection(params)
	conn.run()
	return g, conn, clientConn, events
}

func newPipeTestConnection(t *testing.T, opts ...Option) (*GoJT808, *connection, net.Conn, *recordingTerminalEvent) {
	t.Helper()
	g, conn, clientConn, events := startRawPipeConn(t, nil, opts...)
	if _, err := clientConn.Write(heartbeatPacket); err != nil {
		t.Fatalf("write heartbeat error = %v", err)
	}
	_ = readPacket(t, clientConn, time.Second)
	_ = waitChan(t, events.joined, time.Second)
	_ = waitChan(t, events.readCmd, time.Second)
	_ = waitChan(t, events.writeCmd, time.Second)
	return g, conn, clientConn, events
}

func TestConnection_uplinkReplyAndReissueBranches(t *testing.T) {
	errReplyHandler := &recordingHandler{protocol: consts.T0100Register, replyErr: errors.New("encode err")}
	bigReplyHandler := &recordingHandler{protocol: consts.T0102RegisterAuth, replyBody: bytes.Repeat([]byte{0x55}, 1200)}

	_, conn, clientConn, events := newPipeTestConnection(t, WithCustomHandleFunc(func() map[consts.JT808CommandType]Handler {
		return map[consts.JT808CommandType]Handler{
			consts.T0100Register:     errReplyHandler,
			consts.T0102RegisterAuth: bigReplyHandler,
		}
	}))

	// 1. 未注册的上行指令触发 OnNotSupportedEvent
	if _, err := clientConn.Write(encodeTerminalPacket(t, consts.JT808CommandType(0x0F00), []byte{0x01}, 2)); err != nil {
		t.Fatalf("write unknown err = %v", err)
	}
	if got := waitChan(t, events.notSupported, time.Second); got != consts.JT808CommandType(0x0F00) {
		t.Fatalf("notSupported = %s, want 0x0F00", got)
	}

	// 2. HasReply()==false 的上行报文（无在途主动请求时的 T0001）直接跳过回复
	if _, err := clientConn.Write(encodeGeneralRespond(t, 1, consts.P8300TextInfoDistribution, 3)); err != nil {
		t.Fatalf("write T0001 err = %v", err)
	}
	_ = waitChan(t, events.readCmd, time.Second)

	// 3. ReplyBody 返回错误时安全跳过写包
	if _, err := clientConn.Write(encodeTerminalPacket(t, consts.T0100Register, []byte{0x01}, 4)); err != nil {
		t.Fatalf("write T0100 err = %v", err)
	}
	_ = waitChan(t, events.readCmd, time.Second)

	// 4. ReplyBody 超过 1000B 触发多分包默认回复
	doneBig := make(chan int, 1)
	go func() {
		fr := jt808.NewFrameReader()
		var frames [][]byte
		for len(frames) < 2 {
			frames = append(frames, fr.ReadFrames(readPacket(t, clientConn, time.Second))...)
		}
		doneBig <- len(frames)
	}()
	if _, err := clientConn.Write(encodeTerminalPacket(t, consts.T0102RegisterAuth, []byte{0x31}, 5)); err != nil {
		t.Fatalf("write T0102 err = %v", err)
	}
	if n := <-doneBig; n != 2 {
		t.Fatalf("multi-packet default reply frames = %d, want 2", n)
	}
	_ = waitChan(t, events.readCmd, time.Second)
	_ = waitChan(t, events.writeCmd, time.Second)

	// 5. 分包上行过滤（第1包被 filter 拦截不回复，第2包重组完成后统一回复）与 0x8003 补传分支
	base := mustDecodeJTMessage(t, heartbeatPacket)
	base.Header.ReplyID = uint16(consts.T0801MultimediaDataUpload)
	base.Header.PlatformSerialNumber = 20
	subPkts := base.Header.EncodePackets(bytes.Repeat([]byte{0xaa}, 1200))
	if _, err := clientConn.Write(subPkts[0]); err != nil {
		t.Fatalf("write subPkts[0] err = %v", err)
	}
	if _, err := clientConn.Write(subPkts[1]); err != nil {
		t.Fatalf("write subPkts[1] err = %v", err)
	}
	_ = readPacket(t, clientConn, time.Second) // 重组完成后的 0x8001
	_ = waitChan(t, events.readCmd, time.Second)
	_ = waitChan(t, events.writeCmd, time.Second)

	// 5a. 触发 parse 产生 0x8003 -> reissuePackChan -> onReissueSubcontractingEvent 正常写出
	if _, err := clientConn.Write(subPkts[0]); err != nil {
		t.Fatalf("write second subPkts[0] err = %v", err)
	}
	p8003Body := (&model.P0x8003{OriginalSerialNumber: 20, AgainPackageCount: 1, AgainPackageList: []uint16{2}}).Encode()
	reissueJT := mustDecodeJTMessage(t, encodeTerminalPacket(t, consts.P8003ReissueSubcontractingRequest, p8003Body, 1))
	reissueMsg := newTerminalMessage(reissueJT, nil)
	reissueMsg.Handler = conn.handles[consts.P8003ReissueSubcontractingRequest]
	conn.handleMessages([]*Message{reissueMsg})
	got8003 := decodeFirstMessage(t, readPacket(t, clientConn, time.Second))
	if consts.JT808CommandType(got8003.Header.ID) != consts.P8003ReissueSubcontractingRequest {
		t.Fatalf("reissue packet ID = 0x%04X, want 0x8003", got8003.Header.ID)
	}
	_ = waitChan(t, events.writeCmd, time.Second)

	// 5b. 0x8003 报文超大(>1000B)被拦截返回 + 连接关闭时补传写失败分支
	oversizeJT := mustDecodeJTMessage(t, heartbeatPacket)
	oversizeJT.Body = bytes.Repeat([]byte{0x01}, 1005)
	conn.onReissueSubcontractingEvent(newTerminalMessage(oversizeJT, nil))

	_ = clientConn.Close()
	_ = waitChan(t, events.left, time.Second)
	<-conn.finallyCompleteChan
	conn.onReissueSubcontractingEvent(reissueMsg)
}

type errNetConn struct {
	net.Conn
	readErr  error
	closeErr error
}

func (e *errNetConn) Read([]byte) (int, error) { return 0, e.readErr }
func (e *errNetConn) Close() error             { return e.closeErr }

func TestConnection_readerEdgeCasesAndDefensiveGuards(t *testing.T) {
	// 1. 解析错误立即断开
	_, conn1, client1, _ := startRawPipeConn(t, nil)
	_, _ = client1.Write(mustHex("7e000200000123456789017fff7e"))
	<-conn1.finallyCompleteChan

	// 2a. KeyFunc 返回 (_, false) (_errKeyInvalid) 时不断开连接，继续处理应答并在后续报文重试 join
	g2, conn2, client2, events2 := startRawPipeConn(t, nil, WithKeyFunc(func(m *Message) (string, bool) {
		if m.Command == consts.T0100Register {
			return m.JTMessage.Header.TerminalPhoneNo, true
		}
		return "", false
	}))
	_, _ = client2.Write(encodeTerminalPacket(t, consts.T0102RegisterAuth, []byte{0x31}, 1))
	_ = readPacket(t, client2, time.Second) // 仍正常收到 0x8001 应答，连接未断开
	_ = waitChan(t, events2.joined, time.Second)
	if reply := g2.SendActiveMessage(NewActiveMessage("12345678901", consts.P8201QueryLocation, nil, 20*time.Millisecond)); !errors.Is(reply.ExtensionFields.Err, ErrNotExistKey) {
		t.Fatalf("before register join, expected ErrNotExistKey, got %v", reply.ExtensionFields.Err)
	}
	_, _ = client2.Write(encodeTerminalPacket(t, consts.T0100Register, []byte{0x00}, 2))
	_ = readPacket(t, client2, time.Second) // 收到 0x8100 应答并正式加入会话
	if gotKey := waitChan(t, events2.joined, time.Second); gotKey != "12345678901" {
		t.Fatalf("second join key = %q, want 12345678901", gotKey)
	}

	// 2b. 重复在线 key (_errKeyExist) 立即断开新连接，且不影响原在线会话；未 join 即断连不触发 OnLeaveEvent
	dupServer, dupClient := net.Pipe()
	t.Cleanup(func() { _ = dupClient.Close(); _ = dupServer.Close() })
	dupEvents := newRecordingTerminalEvent()
	dupConn := newConnection(connectionParams{
		conn:                 dupServer,
		handles:              g2.createCommandHandle(),
		activeRespondHandles: g2.createActiveRespondHandle(),
		terminalEvent:        dupEvents,
		filter:               true,
		onJoinEvent:          g2.sessionManager.join,
		onLeaveEvent:         g2.sessionManager.leave,
	})
	dupConn.run()
	_, _ = dupClient.Write(encodeTerminalPacket(t, consts.T0100Register, []byte{0x00}, 3))
	<-dupConn.finallyCompleteChan
	select {
	case k := <-dupEvents.left:
		t.Fatalf("duplicate unjoined connection should not trigger OnLeaveEvent, got %q", k)
	default:
	}
	_ = client2.Close()
	_ = waitChan(t, events2.left, time.Second)
	<-conn2.finallyCompleteChan

	// 3. IdleTimeout 触发 OnTerminalTimeoutEvent
	timeoutTriggered := make(chan TerminalTimeout, 1)
	_, conn3, _, _ := startRawPipeConn(t, nil, WithTerminalTimeout(20*time.Millisecond, func(tm TerminalTimeout) {
		timeoutTriggered <- tm
	}))
	tm := waitChan(t, timeoutTriggered, time.Second)
	if tm.Address != "pipe" {
		t.Fatalf("timeout address = %q, want pipe", tm.Address)
	}
	<-conn3.finallyCompleteChan

	// 4. 非 EOF/ErrClosed 的底层读取异常与关闭异常
	_, rawPeer := net.Pipe()
	_ = rawPeer.Close()
	events := newRecordingTerminalEvent()
	g := New(WithCustomTerminalEventer(func() TerminalEventer { return events }))
	errConn := newConnection(connectionParams{
		conn: &errNetConn{
			Conn:     rawPeer,
			readErr:  errors.New("custom read failure"),
			closeErr: errors.New("custom close failure"),
		},
		handles:              g.createCommandHandle(),
		activeRespondHandles: g.createActiveRespondHandle(),
		terminalEvent:        events,
		filter:               true,
		onJoinEvent:          g.sessionManager.join,
		onLeaveEvent:         g.sessionManager.leave,
	})
	errConn.run()
	<-errConn.finallyCompleteChan

	// 5. 读写事件过滤分支（Handler==nil 与未完成分包过滤）及超时去重早退
	subJT := mustDecodeJTMessage(t, heartbeatPacket)
	subJT.Header.SubPackageSum = 2
	subMsg := newTerminalMessage(subJT, nil)
	errConn.onReadExecutionEvent(subMsg)
	errConn.onWriteExecutionEvent(subMsg)

	nilHandlerMsg := newTerminalMessage(mustDecodeJTMessage(t, heartbeatPacket), nil)
	errConn.onReadExecutionEvent(nilHandlerMsg)
	errConn.onWriteExecutionEvent(nilHandlerMsg)
	errConn.onActiveTimeoutEvent(&activeRecord{seq: 1}, map[uint16]*activeRecord{})

	// 6. write 协程在回调异常 panic 时 defer 停止所有在途定时器
	sConn, cConn := net.Pipe()
	defer func() { _ = sConn.Close(); _ = cConn.Close() }()
	go func() {
		buf := make([]byte, 256)
		for {
			if _, err := cConn.Read(buf); err != nil {
				return
			}
		}
	}()
	panicConn := newConnection(connectionParams{
		conn:    sConn,
		handles: g.createCommandHandle(),
		activeRespondHandles: map[consts.JT808CommandType]func(*ActiveMessage, *Message) bool{
			consts.T0002HeartBeat: func(*ActiveMessage, *Message) bool { panic("respond panic") },
		},
		terminalEvent: events,
	})
	act := NewActiveMessage("12345678901", consts.P8201QueryLocation, nil, time.Minute)
	act.header = mustDecodeJTMessage(t, heartbeatPacket).Header
	act.replyChan = make(chan *Message, 1)
	panicDone := make(chan struct{})
	go func() {
		defer close(panicDone)
		defer func() { _ = recover() }()
		panicConn.write()
	}()
	panicConn.activeMsgChan <- act
	time.Sleep(10 * time.Millisecond)
	panicConn.terminalUplinkMsgChan <- newTerminalMessage(mustDecodeJTMessage(t, heartbeatPacket), heartbeatPacket)
	<-panicDone
}

func TestConnection_activeLifecycleAndDrain(t *testing.T) {
	const key = "12345678901"

	t.Run("多分包主动下发_无关上行跳过_未注册主动指令与默认超时分支", func(t *testing.T) {
		const customActiveCmd = consts.JT808CommandType(0x8F00)
		g, _, clientConn, events := newPipeTestConnection(t,
			WithCustomActiveRespondHandlerFunc(func() map[consts.JT808CommandType]func(*ActiveMessage, *Message) bool {
				return map[consts.JT808CommandType]func(*ActiveMessage, *Message) bool{
					consts.T0002HeartBeat: func(a *ActiveMessage, _ *Message) bool {
						return a.Command == customActiveCmd
					},
				}
			}),
		)

		// 1. 分包主动下发(>1000B)，中间穿插无关上行(T0100)不误匹配，最后按末包流水号应答成功
		subBody := bytes.Repeat([]byte{0x31}, 1500)
		go func() {
			fr := jt808.NewFrameReader()
			var frames [][]byte
			for len(frames) < 2 {
				frames = append(frames, fr.ReadFrames(readPacket(t, clientConn, time.Second))...)
			}
			// 发送一条不在 activeRespondHandles 中的上行消息，覆盖 onActiveRespondEvent 返回 false
			_, _ = clientConn.Write(encodeTerminalPacket(t, consts.T0100Register, []byte{0x01}, 4))
			_ = readPacket(t, clientConn, time.Second) // 消费 0x8100 回复

			lastMsg := decodeFirstMessage(t, frames[len(frames)-1])
			_, _ = clientConn.Write(encodeGeneralRespond(t, lastMsg.Header.SerialNumber, consts.P8300TextInfoDistribution, 6))
		}()
		if reply := g.SendActiveMessage(NewActiveMessage(key, consts.P8300TextInfoDistribution, subBody, time.Second)); reply.ExtensionFields.Err != nil {
			t.Fatalf("subcontract active send err = %v", reply.ExtensionFields.Err)
		}

		// 2. 未注册的主动指令(0x8F00) + OverTimeDuration=0(走 defaultActiveTimeout)，由自定义应答立即完成
		go func() {
			_ = readPacket(t, clientConn, time.Second)
			_, _ = clientConn.Write(heartbeatPacket)
		}()
		if reply := g.SendActiveMessage(NewActiveMessage(key, customActiveCmd, []byte{0x01}, 0)); reply.ExtensionFields.Err != nil {
			t.Fatalf("custom active cmd err = %v", reply.ExtensionFields.Err)
		}
		if got := waitChan(t, events.notSupported, time.Second); got != customActiveCmd {
			t.Fatalf("notSupported = %s, want 0x8F00", got)
		}

		// 3. 连接保持在线时主动下发正常超时（触发 write 主循环 activeTimeoutChan 分支）
		go func() { _ = readPacket(t, clientConn, time.Second) }()
		if timeoutReply := g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, 15*time.Millisecond)); !errors.Is(timeoutReply.ExtensionFields.Err, ErrWriteDataOverTime) {
			t.Fatalf("in-loop timeout err = %v, want %v", timeoutReply.ExtensionFields.Err, ErrWriteDataOverTime)
		}
	})

	t.Run("流水号回绕冲突与drainBeforeExit完整收尾", func(t *testing.T) {
		g, conn, clientConn, _ := newPipeTestConnection(t)

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

		conn.platformSerialNumber.Store(1)
		go func() {
			pkt := readPacket(t, clientConn, time.Second)
			msg := decodeFirstMessage(t, pkt)
			time.Sleep(80 * time.Millisecond)
			_, _ = clientConn.Write(encodeLocationQueryRespond(t, msg.Header.SerialNumber, 9))
		}()
		if second := g.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, 500*time.Millisecond)); second.ExtensionFields.Err != nil {
			t.Fatalf("second active request failed: %v", second.ExtensionFields.Err)
		}
		if old := <-firstRes; !errors.Is(old.ExtensionFields.Err, ErrWriteDataOverTime) {
			t.Fatalf("collided old request err = %v, want %v", old.ExtensionFields.Err, ErrWriteDataOverTime)
		}

		// 直接验证 drainBeforeExit 三阶段排空：已入队上行应答匹配 + 待发主动消息发送与超时收尾
		sConn, cConn := net.Pipe()
		defer func() { _ = sConn.Close(); _ = cConn.Close() }()
		go func() {
			buf := make([]byte, 512)
			for {
				if _, err := cConn.Read(buf); err != nil {
					return
				}
			}
		}()
		drainConn := newConnection(connectionParams{
			conn:                 sConn,
			handles:              g.createCommandHandle(),
			activeRespondHandles: g.createActiveRespondHandle(),
			terminalEvent:        newRecordingTerminalEvent(),
			filter:               true,
		})
		done := make(chan struct{})
		defer close(done)
		record := map[uint16]*activeRecord{}

		act1 := NewActiveMessage(key, consts.P8201QueryLocation, nil, time.Second)
		act1.header = mustDecodeJTMessage(t, heartbeatPacket).Header
		act1.replyChan = make(chan *Message, 1)
		drainConn.onActiveSendEvent(act1, record, done)

		respJT := mustDecodeJTMessage(t, encodeLocationQueryRespond(t, act1.ExtensionFields.PlatformSeq, 11))
		respMsg := newTerminalMessage(respJT, nil)
		respMsg.Handler = drainConn.handles[consts.T0201QueryLocation]
		drainConn.terminalUplinkMsgChan <- respMsg

		act2 := NewActiveMessage(key, consts.P8201QueryLocation, nil, 15*time.Millisecond)
		act2.header = mustDecodeJTMessage(t, heartbeatPacket).Header
		act2.replyChan = make(chan *Message, 1)
		drainConn.activeMsgChan <- act2

		drainConn.drainBeforeExit(record, done)
		if r1 := <-act1.replyChan; r1.ExtensionFields.Err != nil {
			t.Fatalf("drain matched reply err = %v", r1.ExtensionFields.Err)
		}
		if r2 := <-act2.replyChan; !errors.Is(r2.ExtensionFields.Err, ErrWriteDataOverTime) {
			t.Fatalf("drain timeout reply err = %v, want %v", r2.ExtensionFields.Err, ErrWriteDataOverTime)
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

	const burst = 30
	writeCh := make(chan []byte, burst+2)
	doneWrite := make(chan struct{})
	go func() {
		defer close(doneWrite)
		for pkt := range writeCh {
			if _, err := clientConn.Write(pkt); err != nil {
				t.Errorf("client write err = %v", err)
				return
			}
		}
	}()

	doneRead := make(chan struct{})
	go func() {
		defer close(doneRead)
		fr := jt808.NewFrameReader()
		buf := make([]byte, 4096)
		locReplies := 0
		repliedActive := false
		for locReplies < burst || !repliedActive {
			_ = clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, err := clientConn.Read(buf)
			if err != nil {
				t.Errorf("client read err = %v", err)
				return
			}
			for _, frame := range fr.ReadFrames(buf[:n]) {
				msg := decodeFirstMessage(t, frame)
				if consts.JT808CommandType(msg.Header.ID) == consts.P8201QueryLocation {
					writeCh <- encodeLocationQueryRespond(t, msg.Header.SerialNumber, 99)
					repliedActive = true
				} else {
					locReplies++
				}
			}
		}
		close(writeCh)
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if res := g.SendActiveMessage(NewActiveMessage("12345678901", consts.P8201QueryLocation, nil, 2*time.Second)); res.ExtensionFields.Err != nil {
			t.Errorf("concurrent SendActiveMessage err = %v", res.ExtensionFields.Err)
		}
	}()

	locBody := make([]byte, 28)
	for i := 0; i < burst; i++ {
		writeCh <- encodeTerminalPacket(t, consts.T0200LocationReport, locBody, uint16(i+1))
	}
	wg.Wait()
	<-doneRead
	<-doneWrite
}

func TestConnection_maxFrameBufferSize(t *testing.T) {
	// 1. 默认 1023*2 阈值：未达上限不断开，凑满 2046 字节立即断开
	_, conn, clientConn, events := newPipeTestConnection(t)
	part1 := append([]byte{0x7e}, bytes.Repeat([]byte{0x01}, 999)...)
	part2 := bytes.Repeat([]byte{0x02}, 1000)
	_, _ = clientConn.Write(part1)
	_, _ = clientConn.Write(part2)
	select {
	case <-conn.finallyCompleteChan:
		t.Fatal("closed prematurely before reaching 1023*2")
	case <-time.After(20 * time.Millisecond):
	}
	_, _ = clientConn.Write(bytes.Repeat([]byte{0x03}, 1023*2-len(part1)-len(part2)))
	_ = waitChan(t, events.left, time.Second)
	<-conn.finallyCompleteChan
	if _, err := clientConn.Read(make([]byte, 8)); !errors.Is(err, io.EOF) {
		t.Fatalf("client read err = %v, want io.EOF", err)
	}

	// 2. 自定义阈值(64B)溢出断开 与 WithMaxFrameBufferSize(0) 禁用上限
	_, conn64, client64, _ := startRawPipeConn(t, nil, WithMaxFrameBufferSize(64))
	_, _ = client64.Write(bytes.Repeat([]byte("A"), 64))
	<-conn64.finallyCompleteChan

	_, connUnlimited, clientUnlimited, _ := startRawPipeConn(t, nil, WithMaxFrameBufferSize(0))
	_, _ = clientUnlimited.Write(append([]byte{0x7e}, bytes.Repeat([]byte{0x01}, 1023*2+100)...))
	select {
	case <-connUnlimited.finallyCompleteChan:
		t.Fatal("closed unexpectedly when MaxFrameBufferSize=0")
	case <-time.After(20 * time.Millisecond):
	}
}

func TestConnection_writeTimeoutClosesStuckConnection(t *testing.T) {
	const key = "12345678901"
	shortWrite := func(p *connectionParams) { p.writeTimeout = 40 * time.Millisecond }

	// 1. 上行默认应答写超时 -> 自动关闭连接并移除会话
	g1, conn1, client1, events1 := startRawPipeConn(t, shortWrite)
	_, _ = client1.Write(heartbeatPacket)
	_ = readPacket(t, client1, time.Second)
	_ = waitChan(t, events1.joined, time.Second)
	_, _ = client1.Write(heartbeatPacket) // 客户端不读应答，触发写超时
	_ = waitChan(t, events1.left, time.Second)
	<-conn1.finallyCompleteChan
	if reply := g1.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, 50*time.Millisecond)); !errors.Is(reply.ExtensionFields.Err, ErrNotExistKey) {
		t.Fatalf("after write timeout err = %v, want %v", reply.ExtensionFields.Err, ErrNotExistKey)
	}

	// 2. 主动下发写超时 -> 返回 ErrWriteDataFail + os.ErrDeadlineExceeded 并关闭连接
	g2, conn2, client2, events2 := startRawPipeConn(t, shortWrite)
	_, _ = client2.Write(heartbeatPacket)
	_ = readPacket(t, client2, time.Second)
	_ = waitChan(t, events2.joined, time.Second)
	reply := g2.SendActiveMessage(NewActiveMessage(key, consts.P8201QueryLocation, nil, time.Second))
	if !errors.Is(reply.ExtensionFields.Err, ErrWriteDataFail) || !errors.Is(reply.ExtensionFields.Err, os.ErrDeadlineExceeded) {
		t.Fatalf("active write timeout err = %v", reply.ExtensionFields.Err)
	}
	_ = waitChan(t, events2.left, time.Second)
	<-conn2.finallyCompleteChan
}
