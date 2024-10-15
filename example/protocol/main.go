package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
)

func main() {
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, "1"))
	data := t.CreateDefaultCommandData(consts.T0100Register)
	fmt.Println(fmt.Sprintf("模拟器生成的[%x]", data))

	jtMsg := jt808.NewJTMessage()
	_ = jtMsg.Decode(data)

	var t0x0100 model.T0x0100
	_ = t0x0100.Parse(jtMsg)
	fmt.Println(jtMsg.Header.String())
	fmt.Println(t0x0100.String())
}
