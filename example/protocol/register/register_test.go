package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
)

func Example() {
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, "1"))
	data := t.CreateDefaultCommandData(consts.T0100Register)
	fmt.Println(fmt.Sprintf("模拟器生成的指令 [%x]", data))

	jtMsg := jt808.NewJTMessage()
	_ = jtMsg.Decode(data)

	var t0x0100 model.T0x0100
	_ = t0x0100.Parse(jtMsg)
	fmt.Println(jtMsg.Header.String())
	fmt.Println(t0x0100.String())

	// Output:
	//模拟器生成的指令 [7e010000300000000000010001001f006e63643132337777772e3830382e636f6d0000000000000000003736353433323101b2e2413132333435363738797e]
	//[0100] 消息ID:[256] [终端-注册]
	//消息体属性对象: {
	//	[0000000000110000] 消息体属性对象:[48]
	//	版本号:[JT2013]
	//	[bit15] [0]
	//	[bit14] 协议版本标识:[0]
	//	[bit13] 是否分包:[false]
	//	[bit10-12] 加密标识:[0] 0-不加密 1-RSA
	//	[bit0-bit9] 消息体长度:[48]
	//}
	//[000000000001] 终端手机号:[1]
	//[0001] 消息流水号:[1]
	//数据体对象:{
	//	终端-注册:[001f006e63643132337777772e3830382e636f6d0000000000000000003736353433323101b2e2413132333435363738]
	//	[001f] 省域ID:[31]
	//	[006e] 市县域ID:[110]
	//	[6364313233] 制造商ID(5):[cd123]
	//	[7777772e3830382e636f6d000000000000000000] 终端型号(20):[www.808.com]
	//	[37363534333231] 终端ID(7):[7654321]
	//	[01] 车牌颜色:[1]
	//	[b2e2413132333435363738] 车牌号:[测A12345678]
	//}
}
