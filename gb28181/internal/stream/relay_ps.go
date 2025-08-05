package stream

import (
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"log/slog"
	"strconv"
	"time"
)

type RelayPS struct {
	ssrc32    uint32
	startTime time.Time
}

func NewRelayPS() *RelayPS {
	return &RelayPS{}
}

func (r *RelayPS) OnAck(info *command.InviteInfo) {
	if v, err := strconv.ParseUint(info.SSRC, 10, 32); err == nil {
		r.ssrc32 = uint32(v)
	}
	r.startTime = time.Now()
}

func (r *RelayPS) ConvertToGB28181(data []byte) ([][]byte, error) {
	return [][]byte{data}, nil
}

func (r *RelayPS) OnBye(msg string) {
	slog.Info("relay ps bye",
		slog.Any("ssrc32", r.ssrc32),
		slog.Any("time", time.Since(r.startTime)),
		slog.String("msg", msg))
}
