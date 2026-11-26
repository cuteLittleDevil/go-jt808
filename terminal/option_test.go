package terminal

import (
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"testing"
)

func TestWithHeader(t *testing.T) {
	tmp := New(WithHeader(consts.JT808Protocol2013, "678"))
	data := tmp.CreateDefaultCommandData(consts.T0002HeartBeat)
	msg := fmt.Sprintf("%x", data)
	want := "7e0002000000000000067800017d017e"
	if msg != want {
		t.Errorf("WithHeader()=%s\n want %s", msg, want)
		return
	}
}

func TestWithHeader2019(t *testing.T) {
	tmp := New(WithHeader(consts.JT808Protocol2019, "678"))
	data := tmp.CreateDefaultCommandData(consts.T0002HeartBeat)
	jtMsg := jt808.NewJTMessage()
	if err := jtMsg.Decode(data); err != nil {
		t.Fatalf("Decode() err = %v", err)
	}
	if jtMsg.Header.ProtocolVersion != consts.JT808Protocol2019 {
		t.Errorf("ProtocolVersion = %v, want %v", jtMsg.Header.ProtocolVersion, consts.JT808Protocol2019)
	}
	if jtMsg.Header.TerminalPhoneNo != "678" {
		t.Errorf("TerminalPhoneNo = %q, want %q", jtMsg.Header.TerminalPhoneNo, "678")
	}
	if jtMsg.Header.SerialNumber != 1 {
		t.Errorf("SerialNumber = %d, want 1", jtMsg.Header.SerialNumber)
	}
}

func TestWithHeaderVerifyCodeEscapeAndDecodeError(t *testing.T) {
	// 0x02 ^ 0x78 ^ 0x04 = 0x7e (triggers 0x7e escape)
	t7e := New(WithHeader(consts.JT808Protocol2013, "7804"))
	if t7e.TerminalPhoneNo != "000000007804" {
		t.Errorf("t7e TerminalPhoneNo = %q, want 000000007804", t7e.TerminalPhoneNo)
	}
	jtMsg7e := jt808.NewJTMessage()
	if err := jtMsg7e.Decode(t7e.CreateDefaultCommandData(consts.T0002HeartBeat)); err != nil {
		t.Fatalf("t7e decode err = %v", err)
	}

	// 0x02 ^ 0x78 ^ 0x07 = 0x7d (triggers 0x7d escape)
	t7d := New(WithHeader(consts.JT808Protocol2013, "7807"))
	if t7d.TerminalPhoneNo != "000000007807" {
		t.Errorf("t7d TerminalPhoneNo = %q, want 000000007807", t7d.TerminalPhoneNo)
	}
	jtMsg7d := jt808.NewJTMessage()
	if err := jtMsg7d.Decode(t7d.CreateDefaultCommandData(consts.T0002HeartBeat)); err != nil {
		t.Fatalf("t7d decode err = %v", err)
	}

	// Invalid hex phone triggers decode error path in WithHeader
	tErr := New(WithHeader(consts.JT808Protocol2013, "zzzz"))
	if tErr.TerminalPhoneNo != "00000000zzzz" {
		t.Errorf("tErr TerminalPhoneNo = %q, want 00000000zzzz", tErr.TerminalPhoneNo)
	}
}

func TestWithCustomHeaderAndProtocolHandleFunc(t *testing.T) {
	jtMsg := jt808.NewJTMessage()
	raw, _ := hex.DecodeString("7e000200001234567820130001387e")
	if err := jtMsg.Decode(raw); err != nil {
		t.Fatalf("Decode err = %v", err)
	}
	customCmd := consts.JT808CommandType(0x9998)
	term := New(
		WithCustomHeader(jtMsg.Header),
		WithCustomProtocolHandleFunc(func() map[consts.JT808CommandType]Handler {
			return map[consts.JT808CommandType]Handler{
				customCmd: &model.T0x0002{},
			}
		}),
	)
	if term.TerminalPhoneNo != "123456782013" {
		t.Errorf("TerminalPhoneNo = %q, want 123456782013", term.TerminalPhoneNo)
	}
	if data := term.CreateDefaultCommandData(customCmd); len(data) == 0 {
		t.Errorf("CreateDefaultCommandData(customCmd) returned empty")
	}
}
