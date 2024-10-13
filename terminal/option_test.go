package terminal

import (
	"fmt"
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
