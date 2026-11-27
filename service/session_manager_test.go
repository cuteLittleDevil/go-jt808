package service

import (
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

// 2013 版心跳：手机号 12345678901
const heartbeatPacketHex = "7e0002000001234567890100008a7e"

var heartbeatPacket = mustHex(heartbeatPacketHex)

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func mustDecodeJTMessage(t *testing.T, data []byte) *jt808.JTMessage {
	t.Helper()
	jtMsg := jt808.NewJTMessage()
	if err := jtMsg.Decode(data); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return jtMsg
}

func TestSessionManager_writeNotExistAndRoundTrip(t *testing.T) {
	sm := newSessionManager(func(message *Message) (string, bool) {
		return message.JTMessage.Header.TerminalPhoneNo, true
	})
	go sm.run()

	reply := sm.write(NewActiveMessage("missing", consts.P8201QueryLocation, nil, time.Second))
	if !errors.Is(reply.ExtensionFields.Err, ErrNotExistKey) {
		t.Fatalf("write() err = %v, want %v", reply.ExtensionFields.Err, ErrNotExistKey)
	}

	activeChan := make(chan *ActiveMessage, 1)
	msg := newTerminalMessage(mustDecodeJTMessage(t, heartbeatPacket), heartbeatPacket)
	key, err := sm.join(msg, activeChan)
	if err != nil {
		t.Fatalf("join() error = %v", err)
	}
	if key != "12345678901" {
		t.Fatalf("join() key = %s, want 12345678901", key)
	}

	// 重复加入应失败
	if _, err := sm.join(msg, make(chan *ActiveMessage, 1)); !errors.Is(err, _errKeyExist) {
		t.Fatalf("second join() error = %v, want %v", err, _errKeyExist)
	}

	go func() {
		active := <-activeChan
		if active.Key != key {
			t.Errorf("active key = %s, want %s", active.Key, key)
		}
		if active.header == nil {
			t.Error("active header is nil")
		}
		active.replyChan <- &Message{Command: consts.T0001GeneralRespond}
	}()

	got := sm.write(NewActiveMessage(key, consts.P8201QueryLocation, nil, time.Second))
	if got.ExtensionFields.Err != nil {
		t.Fatalf("write() err = %v", got.ExtensionFields.Err)
	}
	if got.Command != consts.T0001GeneralRespond {
		t.Fatalf("write() command = %s, want %s", got.Command, consts.T0001GeneralRespond)
	}

	sm.leave(key)
	reply = sm.write(NewActiveMessage(key, consts.P8201QueryLocation, nil, time.Second))
	if !errors.Is(reply.ExtensionFields.Err, ErrNotExistKey) {
		t.Fatalf("after leave write() err = %v, want %v", reply.ExtensionFields.Err, ErrNotExistKey)
	}
}

func TestSessionManager_invalidKey(t *testing.T) {
	sm := newSessionManager(func(_ *Message) (string, bool) {
		return "", false
	})
	go sm.run()

	msg := newTerminalMessage(mustDecodeJTMessage(t, heartbeatPacket), heartbeatPacket)
	if _, err := sm.join(msg, make(chan *ActiveMessage, 1)); !errors.Is(err, _errKeyInvalid) {
		t.Fatalf("join() error = %v, want %v", err, _errKeyInvalid)
	}
}

func TestSessionManager_activeQueueFullDoesNotBlockOtherTerminals(t *testing.T) {
	keyIdx := 0
	keys := []string{"10000000001", "10000000002", "10000000003"}
	sm := newSessionManager(func(_ *Message) (string, bool) {
		k := keys[keyIdx]
		keyIdx++
		return k, true
	})
	go sm.run()

	msg := newTerminalMessage(mustDecodeJTMessage(t, heartbeatPacket), heartbeatPacket)

	// 终端1：队列容量为2，且无人消费，模拟终端写协程阻塞导致下发队列塞满
	stuckChan := make(chan *ActiveMessage, 2)
	stuckKey, err := sm.join(msg, stuckChan)
	if err != nil {
		t.Fatalf("join stuck terminal err = %v", err)
	}

	// 终端2：正常消费并应答
	normalChan := make(chan *ActiveMessage, 2)
	normalKey, err := sm.join(msg, normalChan)
	if err != nil {
		t.Fatalf("join normal terminal err = %v", err)
	}
	stopNormal := make(chan struct{})
	defer close(stopNormal)
	go func() {
		for {
			select {
			case <-stopNormal:
				return
			case active := <-normalChan:
				active.replyChan <- &Message{Command: consts.T0001GeneralRespond}
			}
		}
	}()

	// 先把终端1的队列塞满
	inFlightReplies := make(chan *Message, cap(stuckChan))
	for i := 0; i < cap(stuckChan); i++ {
		go func() {
			inFlightReplies <- sm.write(NewActiveMessage(stuckKey, consts.P8201QueryLocation, nil, time.Second))
		}()
	}
	deadline := time.Now().Add(time.Second)
	for len(stuckChan) < cap(stuckChan) {
		if time.Now().After(deadline) {
			t.Fatalf("stuckChan len = %d, want %d", len(stuckChan), cap(stuckChan))
		}
		time.Sleep(5 * time.Millisecond)
	}

	// 队列已满时继续向终端1下发：应立即返回 ErrActiveQueueFull，不阻塞 sessionManager
	doneFull := make(chan *Message, 1)
	go func() {
		doneFull <- sm.write(NewActiveMessage(stuckKey, consts.P8201QueryLocation, nil, time.Second))
	}()
	select {
	case fullReply := <-doneFull:
		if !errors.Is(fullReply.ExtensionFields.Err, ErrActiveQueueFull) {
			t.Fatalf("full queue err = %v, want %v", fullReply.ExtensionFields.Err, ErrActiveQueueFull)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("sm.write blocked when target terminal activeMsgChan was full")
	}

	// 验证其他终端的下发、上线、下线均不受影响
	normalDone := make(chan *Message, 1)
	go func() {
		normalDone <- sm.write(NewActiveMessage(normalKey, consts.P8201QueryLocation, nil, time.Second))
	}()
	select {
	case got := <-normalDone:
		if got.ExtensionFields.Err != nil || got.Command != consts.T0001GeneralRespond {
			t.Fatalf("normal terminal write got err=%v cmd=%s", got.ExtensionFields.Err, got.Command)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("normal terminal write was blocked by stuck terminal")
	}

	joinDone := make(chan error, 1)
	go func() {
		_, err := sm.join(msg, make(chan *ActiveMessage, 1))
		joinDone <- err
	}()
	select {
	case err := <-joinDone:
		if err != nil {
			t.Fatalf("third terminal join err = %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("join was blocked by stuck terminal")
	}
	sm.leave("10000000003")

	// 排空终端1的在途消息，避免测试残留协程
	for i := 0; i < cap(stuckChan); i++ {
		active := <-stuckChan
		active.replyChan <- newErrMessage(ErrWriteDataOverTime)
		<-inFlightReplies
	}
}
