package service

import (
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type recordingHandler struct {
	model.BaseHandle
	protocol consts.JT808CommandType
	mu       sync.Mutex
	reads    int
	writes   int
}

func (r *recordingHandler) Protocol() consts.JT808CommandType { return r.protocol }

func (r *recordingHandler) OnReadExecutionEvent(_ *Message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reads++
}

func (r *recordingHandler) OnWriteExecutionEvent(_ Message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.writes++
}

func (r *recordingHandler) counts() (reads, writes int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reads, r.writes
}

type recordingTerminalEvent struct {
	mu            sync.Mutex
	joins         []string
	leaves        []string
	notSupported  []consts.JT808CommandType
	joinErrs      []error
	readCommands  []consts.JT808CommandType
	writeCommands []consts.JT808CommandType
	writeActive   []bool
	writeErrs     []error
}

func (r *recordingTerminalEvent) OnJoinEvent(msg *Message, key string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.joins = append(r.joins, key)
	r.joinErrs = append(r.joinErrs, err)
	_ = msg
}

func (r *recordingTerminalEvent) OnLeaveEvent(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.leaves = append(r.leaves, key)
}

func (r *recordingTerminalEvent) OnNotSupportedEvent(msg *Message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notSupported = append(r.notSupported, msg.Command)
}

func (r *recordingTerminalEvent) OnReadExecutionEvent(msg *Message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.readCommands = append(r.readCommands, msg.Command)
}

func (r *recordingTerminalEvent) OnWriteExecutionEvent(msg Message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.writeCommands = append(r.writeCommands, msg.Command)
	r.writeActive = append(r.writeActive, msg.ExtensionFields.ActiveSend)
	r.writeErrs = append(r.writeErrs, msg.ExtensionFields.Err)
}

func (r *recordingTerminalEvent) snapshot() recordingTerminalEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := recordingTerminalEvent{
		joins:         append([]string(nil), r.joins...),
		leaves:        append([]string(nil), r.leaves...),
		notSupported:  append([]consts.JT808CommandType(nil), r.notSupported...),
		joinErrs:      append([]error(nil), r.joinErrs...),
		readCommands:  append([]consts.JT808CommandType(nil), r.readCommands...),
		writeCommands: append([]consts.JT808CommandType(nil), r.writeCommands...),
		writeActive:   append([]bool(nil), r.writeActive...),
		writeErrs:     append([]error(nil), r.writeErrs...),
	}
	return cp
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

func TestService_heartbeatJoinAndDefaultReply(t *testing.T) {
	events := &recordingTerminalEvent{}
	hb := &recordingHandler{protocol: consts.T0002HeartBeat}
	g, addr := startTestServer(t,
		WithCustomTerminalEventer(func() TerminalEventer { return events }),
		WithCustomHandleFunc(func() map[consts.JT808CommandType]Handler {
			return map[consts.JT808CommandType]Handler{
				consts.T0002HeartBeat: hb,
			}
		}),
	)
	_ = g

	conn := dialTerminal(t, addr)
	if _, err := conn.Write(heartbeatPacket); err != nil {
		t.Fatalf("Write heartbeat error = %v", err)
	}

	reply := readPacket(t, conn, 2*time.Second)
	jtMsg := decodeFirstMessage(t, reply)
	if consts.JT808CommandType(jtMsg.Header.ID) != consts.P8001GeneralRespond {
		t.Fatalf("reply command = 0x%04X, want 0x8001", jtMsg.Header.ID)
	}

	waitFor(t, 2*time.Second, func() bool {
		s := events.snapshot()
		return len(s.joins) == 1 && s.joins[0] == "12345678901"
	})
	reads, _ := hb.counts()
	if reads == 0 {
		t.Fatal("expected heartbeat OnReadExecutionEvent")
	}
}

func TestService_sendActiveMessageSuccessAndTimeout(t *testing.T) {
	events := &recordingTerminalEvent{}
	g, addr := startTestServer(t,
		WithCustomTerminalEventer(func() TerminalEventer { return events }),
	)

	conn := dialTerminal(t, addr)
	if _, err := conn.Write(heartbeatPacket); err != nil {
		t.Fatalf("Write heartbeat error = %v", err)
	}
	_ = readPacket(t, conn, 2*time.Second) // 消费 0x8001

	waitFor(t, 2*time.Second, func() bool {
		return len(events.snapshot().joins) == 1
	})

	key := "12345678901"

	// 成功路径：下发 0x8201，终端回 0x0201
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

	// 超时路径：下发 0x8300，终端不回复
	go func() {
		_ = readPacket(t, conn, 2*time.Second) // 消费平台下发报文
	}()
	timeoutReply := g.SendActiveMessage(NewActiveMessage(key, consts.P8300TextInfoDistribution, []byte{0x01, 0x31, 0x32, 0x33}, 200*time.Millisecond))
	if !errors.Is(timeoutReply.ExtensionFields.Err, ErrWriteDataOverTime) {
		t.Fatalf("timeout err = %v, want %v", timeoutReply.ExtensionFields.Err, ErrWriteDataOverTime)
	}
}

func TestService_notSupportedCommand(t *testing.T) {
	events := &recordingTerminalEvent{}
	g, addr := startTestServer(t,
		WithCustomTerminalEventer(func() TerminalEventer { return events }),
	)
	_ = g

	conn := dialTerminal(t, addr)
	// 先用心跳加入
	if _, err := conn.Write(heartbeatPacket); err != nil {
		t.Fatalf("Write heartbeat error = %v", err)
	}
	_ = readPacket(t, conn, 2*time.Second)

	// 0x0F00 未实现指令
	unknown := encodeTerminalPacket(t, consts.JT808CommandType(0x0F00), []byte{0x01}, 3)
	if _, err := conn.Write(unknown); err != nil {
		t.Fatalf("Write unknown error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		return len(events.snapshot().notSupported) > 0
	})
	if got := events.snapshot().notSupported[0]; got != consts.JT808CommandType(0x0F00) {
		t.Fatalf("notSupported = %s, want 0x0F00", got)
	}
}

func TestService_defaultHandlersIncludeNewlyRegisteredCommands(t *testing.T) {
	events := &recordingTerminalEvent{}
	g, addr := startTestServer(t,
		WithCustomTerminalEventer(func() TerminalEventer { return events }),
	)

	conn := dialTerminal(t, addr)
	if _, err := conn.Write(heartbeatPacket); err != nil {
		t.Fatalf("Write heartbeat error = %v", err)
	}
	_ = readPacket(t, conn, 2*time.Second)
	waitFor(t, 2*time.Second, func() bool {
		return len(events.snapshot().joins) == 1
	})

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

	// 确认不是走未支持事件
	if len(events.snapshot().notSupported) != 0 {
		t.Fatalf("unexpected notSupported events: %v", events.snapshot().notSupported)
	}
}

func TestPackageParse_completeSubcontract(t *testing.T) {
	base := mustDecodeJTMessage(t, heartbeatPacket)
	header := base.Header
	header.ReplyID = uint16(consts.T0801MultimediaDataUpload)
	header.PlatformSerialNumber = 1

	// 构造超过 1000 字节的 body，触发分包
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
