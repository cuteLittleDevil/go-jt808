package command

import "testing"

func TestToGBType_String(t *testing.T) {
	tests := []struct {
		typ  ToGBType
		want string
	}{
		{JT1078ToPS, "jt1078 -> ps流"},
		{RTSPToPS, "rtsp -> ps流"},
		{RTMPToPS, "rtmp -> ps流"},
		{RelayPS, "ps -> ps流"},
		{CustomPS, "自定义ps流处理"},
		{JT1078ToPSFilterPacket, "jt1078[包异常时跳过] -> ps流"},
		{ToGBType(999), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.typ.String(); got != tt.want {
			t.Errorf("ToGBType(%d).String() = %q, want %q", tt.typ, got, tt.want)
		}
	}
}
