package pkg

import (
	"context"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"m7s.live/v5"
	"time"
)

type Option struct {
	F func(o *Options)
}

type Options struct {
	pubFunc    func(ctx context.Context, pack *jt1078.Packet) (publisher *m7s.Publisher, err error)
	ptsFunc    func(pack *jt1078.Packet) time.Duration
	audio      bool
	sessions   *AudioManager
	onJoinURL  string
	onLeaveURL string
}

func WithPubFunc(pubFunc func(ctx context.Context,
	pack *jt1078.Packet) (publisher *m7s.Publisher, err error)) Option {
	return Option{F: func(o *Options) {
		o.pubFunc = pubFunc
	}}
}

func WithSessions(sessions *AudioManager) Option {
	return Option{F: func(o *Options) {
		o.sessions = sessions
		o.audio = true
	}}
}

func WithURL(onJoinURL, onLeaveURL string) Option {
	return Option{F: func(o *Options) {
		o.onJoinURL = onJoinURL
		o.onLeaveURL = onLeaveURL
	}}
}

func WithPTSFunc(ptsFunc func(pack *jt1078.Packet) time.Duration) Option {
	return Option{F: func(o *Options) {
		o.ptsFunc = ptsFunc
	}}
}
