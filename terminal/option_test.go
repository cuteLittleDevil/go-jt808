package terminal

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
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
