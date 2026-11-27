package service

import (
	"bytes"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

func TestOptions_andDefaultEventer(t *testing.T) {
	// 1. 默认选项与默认事件实现覆盖
	gDefault := New()
	if gDefault.opts.MaxFrameBufferSize != 1023*2 {
		t.Fatalf("default MaxFrameBufferSize = %d, want %d", gDefault.opts.MaxFrameBufferSize, 1023*2)
	}
	msg := newTerminalMessage(mustDecodeJTMessage(t, heartbeatPacket), heartbeatPacket)
	if key, ok := gDefault.opts.KeyFunc(msg); !ok || key != "12345678901" {
		t.Fatalf("default KeyFunc = (%q, %v)", key, ok)
	}

	ev := gDefault.opts.CustomTerminalEventerFunc()
	ev.OnJoinEvent(msg, "12345678901", nil)
	ev.OnNotSupportedEvent(msg)
	ev.OnReadExecutionEvent(msg)
	ev.OnWriteExecutionEvent(*msg)
	ev.OnLeaveEvent("12345678901")

	defHandles := gDefault.createCommandHandle()
	defHandles[consts.T0002HeartBeat].OnReadExecutionEvent(msg)
	defHandles[consts.T0002HeartBeat].OnWriteExecutionEvent(*msg)

	// 2. 所有自定义 Option 覆盖
	custom := &recordingHandler{protocol: consts.T0002HeartBeat}
	gCustom := New(
		WithHostPorts("127.0.0.1:1808"),
		WithNetwork("tcp"),
		WithHasSubcontract(false),
		WithMaxFrameBufferSize(512),
		WithKeyFunc(func(_ *Message) (string, bool) { return "k1", true }),
		WithTerminalTimeout(time.Second, func(TerminalTimeout) {}),
		WithCustomTerminalEventer(func() TerminalEventer { return newRecordingTerminalEvent() }),
		WithCustomHandleFunc(func() map[consts.JT808CommandType]Handler {
			return map[consts.JT808CommandType]Handler{consts.T0002HeartBeat: custom}
		}),
		WithCustomActiveRespondHandlerFunc(func() map[consts.JT808CommandType]func(*ActiveMessage, *Message) bool {
			return map[consts.JT808CommandType]func(*ActiveMessage, *Message) bool{
				consts.T0002HeartBeat: func(*ActiveMessage, *Message) bool { return true },
			}
		}),
	)
	if gCustom.opts.Addr != "127.0.0.1:1808" || gCustom.opts.FilterSubcontract || gCustom.opts.MaxFrameBufferSize != 512 {
		t.Fatalf("custom options mismatch: %+v", gCustom.opts)
	}
	if got := gCustom.createCommandHandle()[consts.T0002HeartBeat]; got != custom {
		t.Fatalf("custom handler = %#v, want %#v", got, custom)
	}
	if _, ok := gCustom.createActiveRespondHandle()[consts.T0002HeartBeat]; !ok {
		t.Fatal("custom active respond handler not registered")
	}
}

func TestCreateDefaultHandle_andRespondMatchers(t *testing.T) {
	g := New()
	handles := g.createCommandHandle()

	for _, cmd := range []consts.JT808CommandType{
		consts.P9105AudioVideoControlStatusNotice,
		consts.P9201SendVideoRecordRequest,
		consts.P9202SendVideoRecordControl,
		consts.P9212FileUploadCompleteRespond,
	} {
		if h, ok := handles[cmd]; !ok || h.Protocol() != cmd {
			t.Fatalf("default handler %s invalid: ok=%v", cmd, ok)
		}
	}

	respond := g.createActiveRespondHandle()

	// 1. T1003 按指令类型匹配 P9003
	t1003Msg := &Message{Command: consts.T1003UploadAudioVideoAttr, JTMessage: jt808.NewJTMessage()}
	if !respond[consts.T1003UploadAudioVideoAttr](&ActiveMessage{Command: consts.P9003QueryTerminalAudioVideoProperties}, t1003Msg) {
		t.Fatal("T1003 should match P9003")
	}
	if respond[consts.T1003UploadAudioVideoAttr](&ActiveMessage{Command: consts.P8201QueryLocation}, t1003Msg) {
		t.Fatal("T1003 should not match P8201")
	}

	// 2. 等待后续业务应答的指令（P8801/P9003/P9205/P9206）应跳过前置 T0001
	t0001Body := (&model.T0x0001{SerialNumber: 9, ID: uint16(consts.P8801CameraShootImmediateCommand)}).Encode()
	t0001Msg := &Message{
		Command:   consts.T0001GeneralRespond,
		JTMessage: &jt808.JTMessage{Header: &jt808.Header{}, Body: t0001Body},
	}
	for _, waitCmd := range []consts.JT808CommandType{
		consts.P8801CameraShootImmediateCommand,
		consts.P9003QueryTerminalAudioVideoProperties,
		consts.P9205QueryResourceList,
		consts.P9206FileUploadInstructions,
	} {
		act := &ActiveMessage{Command: waitCmd}
		act.ExtensionFields.PlatformSeq = 9
		if respond[consts.T0001GeneralRespond](act, t0001Msg) {
			t.Fatalf("waiting command %s should ignore preliminary T0001", waitCmd)
		}
	}

	// 3. 全部基于流水号匹配的默认应答处理器（含正常匹配、流水号不匹配、报文损坏解析失败）
	locItem := model.T0x0200LocationItem{DateTime: "250101120000"}
	cases := []struct {
		terminalCmd consts.JT808CommandType
		platformCmd consts.JT808CommandType
		validBody   []byte
	}{
		{consts.T0001GeneralRespond, consts.P8300TextInfoDistribution, (&model.T0x0001{SerialNumber: 9, ID: 0x8300}).Encode()},
		{consts.T0104QueryParameter, consts.P8104QueryTerminalParams, (&model.T0x0104{RespondSerialNumber: 9}).Encode()},
		{consts.T0201QueryLocation, consts.P8201QueryLocation, (&model.T0x0201{RespondSerialNumber: 9, T0x0200LocationItem: locItem}).Encode()},
		{consts.T0302QuestionAnswer, consts.P8302QuestionDistribution, (&model.T0x0302{SerialNumber: 9, AnswerID: 1}).Encode()},
		{consts.T1205UploadAudioVideoResourceList, consts.P9205QueryResourceList, (&model.T0x1205{SerialNumber: 9}).Encode()},
		{consts.T1206FileUploadCompleteNotice, consts.P9206FileUploadInstructions, (&model.T0x1206{RespondSerialNumber: 9}).Encode()},
		{consts.T0805CameraShootImmediately, consts.P8801CameraShootImmediateCommand, (&model.T0x0805{RespondSerialNumber: 9, Result: 0}).Encode()},
	}
	for _, tc := range cases {
		matchFn := respond[tc.terminalCmd]
		act := &ActiveMessage{Command: tc.platformCmd}
		act.ExtensionFields.PlatformSeq = 9

		okMsg := &Message{Command: tc.terminalCmd, JTMessage: &jt808.JTMessage{Header: &jt808.Header{}, Body: tc.validBody}}
		if !matchFn(act, okMsg) {
			t.Fatalf("%s expected match for seq=9", tc.terminalCmd)
		}

		actMismatch := &ActiveMessage{Command: tc.platformCmd}
		actMismatch.ExtensionFields.PlatformSeq = 10
		if matchFn(actMismatch, okMsg) {
			t.Fatalf("%s expected mismatch for seq=10", tc.terminalCmd)
		}

		badMsg := &Message{Command: tc.terminalCmd, JTMessage: &jt808.JTMessage{Header: &jt808.Header{}, Body: nil}}
		if matchFn(act, badMsg) {
			t.Fatalf("%s expected false on invalid empty body", tc.terminalCmd)
		}
	}
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

func TestService_runAndServe(t *testing.T) {
	// 1. Run 地址解析失败与监听失败分支立即返回
	New(WithHostPorts(":::")).Run()
	New(WithHostPorts("203.0.113.1:8080")).Run()

	// 若当前环境允许绑定本地回环端口，覆盖 Run 正常进入 serve 分支
	if probe, err := net.Listen("tcp", "127.0.0.1:0"); err == nil {
		_ = probe.Close()
		go New(WithHostPorts("127.0.0.1:0")).Run()
		time.Sleep(15 * time.Millisecond)
	}

	// 2. serve 临时错误退避、正常接入连接（传递 MaxFrameBufferSize）、遇到 net.ErrClosed 立即退出
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
	ln.steps <- func() (net.Conn, error) { return nil, errors.New("temporary accept error") }
	ln.steps <- func() (net.Conn, error) { return serverConn, nil }
	ln.steps <- func() (net.Conn, error) { return nil, net.ErrClosed }

	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		g.serve(ln)
	}()

	if _, err := clientConn.Write(heartbeatPacket); err != nil {
		t.Fatalf("write heartbeat err = %v", err)
	}
	_ = readPacket(t, clientConn, time.Second)
	_ = waitChan(t, events.joined, time.Second)

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
