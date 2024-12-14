package model

import (
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
)

func Example() {
	// 收到注册消息 回复注册应答
	msg := "7e0100002c0123456789010000001f0073797a6800007777772e6a74743830382e636f6d0000000000003736353433323101b2e24131323334ca7e"
	data, _ := hex.DecodeString(msg)
	jtMsg := jt808.NewJTMessage()
	_ = jtMsg.Decode(data)

	t0x0100 := &T0x0100{}
	body, _ := t0x0100.ReplyBody(jtMsg)
	jtMsg.Header.ReplyID = uint16(t0x0100.ReplyProtocol())
	jtMsg.Header.PlatformSerialNumber = 1
	replyData := jtMsg.Header.Encode(body)

	fmt.Println(fmt.Sprintf("收到 %s", msg))
	fmt.Println(fmt.Sprintf("应答 %x", replyData))

	_ = t0x0100.Parse(jtMsg)
	fmt.Println(t0x0100.String())
	// Output:
	// 收到 7e0100002c0123456789010000001f0073797a6800007777772e6a74743830382e636f6d0000000000003736353433323101b2e24131323334ca7e
	// 应答 7e8100000e01234567890100010000003132333435363738393031367e
	//数据体对象:{
	//	终端-注册:[001f0073797a6800007777772e6a74743830382e636f6d0000000000003736353433323101b2e24131323334]
	//	[001f] 省域ID:[31]
	//	[0073] 市县域ID:[115]
	//	[797a680000] 制造商ID(5):[yzh]
	//	[7777772e6a74743830382e636f6d000000000000] 终端型号(20):[www.jtt808.com]
	//	[37363534333231] 终端ID(7):[7654321]
	//	[01] 车牌颜色:[1]
	//	[b2e24131323334] 车牌号:[测A1234]
	//}
}

func ExampleT0x0200_Parse() {
	msg := "7E0200407D0201000000000202326095590A4F00002000004C100301E0D7F2073E6EAC0064021400142411271542300104000000CC020201D425040000000030010831010914040000007F150400000001160400000001170200011803000709642F0000001F000201323201000035006401E0D40A073E6AC4241127154210FFFF69643030303033241127154217000500767E"
	data, _ := hex.DecodeString(msg)
	jtMsg := jt808.NewJTMessage()
	_ = jtMsg.Decode(data)
	type Me0x0200 struct {
		T0x0200
		T0x0200AdditionExtension0x64
	}
	tmp := &Me0x0200{}
	tmp.T0x0200AdditionDetails.CustomAdditionContentFunc = func(id uint8, content []byte) (AdditionContent, bool) {
		if id == 0x64 {
			tmp.T0x0200AdditionExtension0x64.Parse(id, content)
			return AdditionContent{
				Data:        content,
				CustomValue: tmp.T0x0200AdditionExtension0x64,
			}, true
		}
		return AdditionContent{}, false
	}
	_ = tmp.T0x0200.Parse(jtMsg)
	fmt.Println(tmp.T0x0200.String())
	fmt.Println(tmp.T0x0200AdditionExtension0x64.String())

	// Output:
	//数据体对象:{
	//	终端-位置上报:[00002000004c100301e0d7f2073e6eac006402140014241127154230]
	//	[00002000] 报警标志:[8192]
	//	[004c1003] 状态标志:[4984835]
	//	[01e0d7f2] 纬度:[31512562]
	//	[073e6eac] 经度:[121532076]
	//	[0064] 海拔高度:[100]
	//	[0214] 速度:[532]
	//	[0014] 方向:[20]
	//	[241127154230] 时间:[2024-11-27 15:42:30]
	//}
	//	报警ID:[31] 从0开始循环 不区分报警类型
	//	标志状态:[0] 0x00-不可用 0x01-开始标志 0x02-结束标志
	//	报警事件类型:[2] 车道偏离报警
	//	报警级别:[1] 0x01:一级报警 0x02:二级报警
	//	前车车速:[50] 单位:km/h 范围0-250 仅报警类型为0x01和0x03时有效 不可用时=0x00
	//	前车或行人距离:[50] 单位100ms 范围0-100 仅报警类型0x01 0x02 0x04时有效 不可用时=0x00
	//	偏离类型:[1] 0x01-左侧偏离 0x02-右侧偏离 仅报警类型为0x02时有效 不可用时=0x00
	//	道路标志识别类型:[0] 0x01-限速 0x02-限高 0x03-限重 仅报警类型为0x06和0x10时有效 不可用时=0x00
	//	道路标志识别数据:[0] 识别到道路标志的数据 不可用时=0x00
	//	车速:[53] 单位:km/h 范围0-250
	//	海拔高度:[100] 单位米(m)
	//	纬度:[31511562] 以度为单位的纬度值乘以 10 的 6 次方，精确到百万分之一度
	//	经度:[121531076] 以度为单位的经度值乘以 10 的 6 次方，精确到百万分之一度
	//	时间:[2024-11-27 15:42:10] 时间 YY-MM-DD-hh-mm-ss（GMT+8 时间，本标准中之后涉及的时间均采用此时区）
	//	车辆状态:[65535] {
	//		 ACC: [true]
	//		 左转向状态: [true]
	//		 右转向状态: [true]
	//		 雨刮器状态: [true]
	//		 制动状态: [true]
	//		 插卡状态: [true]
	//		 定位状态: [true]
	//	}
	//	报警标识 [未知]{ 默认使用苏标
	//		 [69643030303033]终端ID:[id00003]
	//		 [241127154217]时间:[2024-11-27 15:42:17]
	//		 [00]序号:[0]
	//		 [05]附件数量:[5]
	//		 [00]预留
	//	}
}
