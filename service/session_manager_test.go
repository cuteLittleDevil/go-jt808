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
