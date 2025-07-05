package command

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
)

type JT1078ToGB28181er interface {
	OnAck(info *InviteInfo)
	ConvertToGB28181(pack *jt1078.Packet) [][]byte
	OnBye(msg string)
}
