package service

import (
	"testing"

	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

func TestCreateDefaultHandle_includesDocumentedJT1078AndActiveSafety(t *testing.T) {
	g := New()
	handles := g.createCommandHandle()

	required := []consts.JT808CommandType{
		consts.P9105AudioVideoControlStatusNotice,
		consts.P9201SendVideoRecordRequest,
		consts.P9202SendVideoRecordControl,
		consts.P9212FileUploadCompleteRespond,
	}
	for _, cmd := range required {
		h, ok := handles[cmd]
		if !ok {
			t.Errorf("missing default handler for %s (0x%04X)", cmd, uint16(cmd))
			continue
		}
		if got := h.Protocol(); got != cmd {
			t.Errorf("handler Protocol() = %s, want %s", got, cmd)
		}
	}
}

func TestCreateDefaultHandle_customHandleOverridesDefault(t *testing.T) {
	custom := &recordingHandler{protocol: consts.T0002HeartBeat}
	g := New(WithCustomHandleFunc(func() map[consts.JT808CommandType]Handler {
		return map[consts.JT808CommandType]Handler{
			consts.T0002HeartBeat: custom,
		}
	}))

	handles := g.createCommandHandle()
	got, ok := handles[consts.T0002HeartBeat]
	if !ok {
		t.Fatal("expected heartbeat handler")
	}
	if got != custom {
		t.Fatalf("custom handler was not applied, got %#v", got)
	}
}
