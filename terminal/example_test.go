package terminal

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

func ExampleTerminal_CreateCommandData() {
	t := New(WithHeader(consts.JT808Protocol2013, "1001"))
	data := t.CreateDefaultCommandData(consts.T0100Register)
	fmt.Println(fmt.Sprintf("%x", data))

	// Output:
	// 7e010000300000000010010001001f006e63643132337777772e3830382e636f6d0000000000000000003736353433323101b2e2413132333435363738697e
}

func ExampleTerminal_ExpectedReply() {
	t := New(WithHeader(consts.JT808Protocol2013, "1001"))
	data := t.CreateDefaultCommandData(consts.T0100Register)
	msg := fmt.Sprintf("%x", data)
	fmt.Println(msg)
	replyData := t.ExpectedReply(1, msg)
	fmt.Println(fmt.Sprintf("%x", replyData))

	// Output:
	// 7e010000300000000010010001001f006e63643132337777772e3830382e636f6d0000000000000000003736353433323101b2e2413132333435363738697e
	// 7e81000007000000001001000100010031303031977e
}

func ExampleTerminal_CreateDefaultCommandData() {
	t := New(WithHeader(consts.JT808Protocol2013, "1001"))
	t0x0100 := &model.T0x0100{
		ProvinceID:         31,
		CityID:             110,
		ManufacturerID:     "cd12345678",
		TerminalModel:      "www.808.com",
		TerminalID:         "7654321",
		PlateColor:         1,
		LicensePlateNumber: "测A12345678",
	}
	data := t.CreateCommandData(consts.T0100Register, t0x0100.Encode())
	fmt.Println(fmt.Sprintf("%x", data))

	// Output:
	// 7e010000240000000010010001001f006e63643132337777772e3830382e3736353433323101b2e24131323334353637381c7e
}

func ExampleTerminal_ProtocolDetails() {
	t := New(WithHeader(consts.JT808Protocol2013, "1001"))
	data := t.CreateDefaultCommandData(consts.T0100Register)
	msg := fmt.Sprintf("%x", data)
	fmt.Println(t.ProtocolDetails(msg))

	// Output:
	// [7e]开始: 126
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
	//[000000001001] 终端手机号:[1001]
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
	//[69] 校验码:[105]
	//[7e]结束: 126
}
