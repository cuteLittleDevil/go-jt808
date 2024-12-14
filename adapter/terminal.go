package adapter

import (
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type Mode int

const (
	// Leader 可以回复任何命令.
	Leader Mode = iota + 1
	// Follower 只允许回复指定命令 默认是平台下发的0x9101等指令.
	Follower
)

type Terminal struct {
	// Mode 模式 1:leader 2:follower.
	Mode Mode
	// TargetAddr 目标地址 例如 127.0.0.1:8080.
	TargetAddr string
	// AllowCommands 允许的命令 leader-默认全部 follower-只允许指定命令.
	AllowCommands []consts.JT808CommandType
}

func (t *Terminal) allowReply(command consts.JT808CommandType) bool {
	if t.Mode == Leader {
		return true
	}
	for _, v := range t.AllowCommands {
		if v == command {
			return true
		}
	}
	return false
}
