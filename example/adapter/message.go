package main

import "github.com/cuteLittleDevil/go-jt808/protocol/jt808"

type Message struct {
	*jt808.JTMessage
	OriginalData []byte
}
