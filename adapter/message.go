package adapter

import "github.com/cuteLittleDevil/go-jt808/protocol/jt808"

type message struct {
	*jt808.JTMessage
	originalData []byte
}
