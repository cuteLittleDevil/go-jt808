package jt808

import (
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

func ExampleJTMessage_Decode() {
	msg := "7e0100002c0123456789010000001f0073797a6800007777772e6a74743830382e636f6d0000000000003736353433323101b2e24131323334ca7e"
	data, _ := hex.DecodeString(msg)
	jtMsg := NewJTMessage()
	_ = jtMsg.Decode(data)
	fmt.Println(jtMsg.Header.String())

	// Output:
	//[0100] 消息ID:[256] [终端-注册]
	//消息体属性对象: {
	//	[0000000000101100] 消息体属性对象:[44]
	//	版本号:[JT2013]
	//	[bit15] [0]
	//	[bit14] 协议版本标识:[0]
	//	[bit13] 是否分包:[false]
	//	[bit10-12] 加密标识:[0] 0-不加密 1-RSA
	//	[bit0-bit9] 消息体长度:[44]
	//}
	//[012345678901] 终端手机号:[12345678901]
	//[0000] 消息流水号:[0]
}

func ExampleHeader_Encode() {

	jtMsg := NewJTMessage()
	{
		msg := "7e0100002c0123456789010000001f0073797a6800007777772e6a74743830382e636f6d0000000000003736353433323101b2e24131323334ca7e"
		data, _ := hex.DecodeString(msg)
		_ = jtMsg.Decode(data)
	}
	msg := "7fff000200"
	body, _ := hex.DecodeString(msg)
	head := jtMsg.Header
	head.ReplyID = uint16(consts.P8001GeneralRespond)
	head.PlatformSerialNumber = 1
	fmt.Println(fmt.Sprintf("%x", jtMsg.Header.Encode(body)))

	// Output:
	// 7e8001000501234567890100017fff0002008f7e
}
