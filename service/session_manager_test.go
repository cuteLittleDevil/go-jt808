package service

import (
	"encoding/hex"
	"errors"
	"strings"
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

func TestSessionManager_lifecycleAndIsolation(t *testing.T) {
	activeStr := NewActiveMessage("12345678901", consts.P8201QueryLocation, []byte{0x01}, time.Second).String()
	if !strings.Contains(activeStr, "12345678901") {
		t.Fatalf("ActiveMessage.String() = %q, missing key", activeStr)
	}

	msg := newTerminalMessage(mustDecodeJTMessage(t, heartbeatPacket), heartbeatPacket)

	// 1. 非法 key 拒绝加入
	badSM := newSessionManager(func(_ *Message) (string, bool) { return "", false })
	go badSM.run()
	if _, err := badSM.join(msg, make(chan *ActiveMessage, 1)); !errors.Is(err, _errKeyInvalid) {
		t.Fatalf("invalid key join err = %v, want %v", err, _errKeyInvalid)
	}

	// 2. 正常加入、重复加入拒绝、下发往返与下线
	keys := []string{"12345678901", "12345678901", "10000000002"}
	idx := 0
	sm := newSessionManager(func(_ *Message) (string, bool) {
		k := keys[idx]
		idx++
		return k, true
	})
	go sm.run()

	if reply := sm.write(NewActiveMessage("missing", consts.P8201QueryLocation, nil, time.Second)); !errors.Is(reply.ExtensionFields.Err, ErrNotExistKey) {
		t.Fatalf("missing key write err = %v, want %v", reply.ExtensionFields.Err, ErrNotExistKey)
	}

	normalChan := make(chan *ActiveMessage, 1)
	normalKey, err := sm.join(msg, normalChan)
	if err != nil || normalKey != "12345678901" {
		t.Fatalf("join got (%q, %v)", normalKey, err)
	}
	if _, err := sm.join(msg, make(chan *ActiveMessage, 1)); !errors.Is(err, _errKeyExist) {
		t.Fatalf("duplicate join err = %v, want %v", err, _errKeyExist)
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

	if got := sm.write(NewActiveMessage(normalKey, consts.P8201QueryLocation, nil, time.Second)); got.ExtensionFields.Err != nil || got.Command != consts.T0001GeneralRespond {
		t.Fatalf("roundtrip got err=%v cmd=%s", got.ExtensionFields.Err, got.Command)
	}

	// 3. 慢终端队列塞满时立即返回 ErrActiveQueueFull，且不阻塞正常终端
	stuckChan := make(chan *ActiveMessage, 1)
	stuckKey, err := sm.join(msg, stuckChan)
	if err != nil {
		t.Fatalf("join stuck terminal err = %v", err)
	}
	inFlight := make(chan *Message, 1)
	go func() {
		inFlight <- sm.write(NewActiveMessage(stuckKey, consts.P8201QueryLocation, nil, time.Second))
	}()
	for len(stuckChan) < 1 {
		time.Sleep(2 * time.Millisecond)
	}

	if full := sm.write(NewActiveMessage(stuckKey, consts.P8201QueryLocation, nil, time.Second)); !errors.Is(full.ExtensionFields.Err, ErrActiveQueueFull) {
		t.Fatalf("full queue err = %v, want %v", full.ExtensionFields.Err, ErrActiveQueueFull)
	}
	if got := sm.write(NewActiveMessage(normalKey, consts.P8201QueryLocation, nil, time.Second)); got.ExtensionFields.Err != nil {
		t.Fatalf("normal terminal blocked or failed: %v", got.ExtensionFields.Err)
	}

	// 清理在途请求与会话
	(<-stuckChan).replyChan <- newErrMessage(ErrWriteDataOverTime)
	<-inFlight

	sm.leave(normalKey)
	if reply := sm.write(NewActiveMessage(normalKey, consts.P8201QueryLocation, nil, time.Second)); !errors.Is(reply.ExtensionFields.Err, ErrNotExistKey) {
		t.Fatalf("after leave write err = %v, want %v", reply.ExtensionFields.Err, ErrNotExistKey)
	}
}
