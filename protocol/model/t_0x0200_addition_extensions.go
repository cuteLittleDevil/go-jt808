package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"strings"
)

type (
	// T0x0200ExtensionSBBase  苏标扩展的基础数据
	T0x0200ExtensionSBBase struct {
		// VehicleSpeed 车速 单位:km/h 范围0-250
		VehicleSpeed byte `json:"vehicleSpeed"`
		// Altitude 海拔高度 单位米(m)
		Altitude uint16 `json:"altitude"`
		// Latitude 纬度 以度为单位的纬度值乘以 10 的 6 次方，精确到百万分之一度
		Latitude uint32 `json:"latitude"`
		// Longitude 经度 以度为单位的经度值乘以 10 的 6 次方，精确到百万分之一度
		Longitude uint32 `json:"longitude"`
		// DateTime 时间 YY-MM-DD-hh-mm-ss（GMT+8 时间，本标准中之后涉及的时间均采用此时区）bcd[6]
		DateTime string `json:"dateTime"`
		// VehicleStatus 车辆状态 见表18
		VehicleStatus T0x0200ExtensionTable18 `json:"vehicleStatus"`
		// P9208AlarmSign 报警标识号 苏标
		P9208AlarmSign `json:"p9208AlarmSign"`
		// ParseSuccess 解析是否成功
		ParseSuccess bool `json:"parseSuccess"`
	}

	// T0x0200ExtensionTable18 表18-车辆 状态标志位含义
	T0x0200ExtensionTable18 struct {
		// OriginalValue 原始值
		OriginalValue uint16 `json:"originalValue"`
		// ACC ACC 0-关 1-开
		ACC bool `json:"acc,omitempty"`
		// LeftTurn 左转向状态标志 0-关 1-开
		LeftTurn bool `json:"leftTurn,omitempty"`
		// RightTurn 右转向状态标志 0-关 1-开
		RightTurn bool `json:"rightTurn,omitempty"`
		// Wipers 雨刮器状态标志 0-关 1-开
		Wipers bool `json:"wipers,omitempty"`
		// Brake 制动状态标志 0-未制动 1-制动
		Brake bool `json:"brake,omitempty"`
		// Card 插卡状态标志 0-未插卡 1-插卡
		Card bool `json:"card,omitempty"`
		// Location 定位状态标志 0-未定位 1-定位
		Location bool `json:"location,omitempty"`
	}

	// T0x0200AdditionExtension0x64 表17-驾驶辅助功能报警信息 苏标
	T0x0200AdditionExtension0x64 struct {
		// AlarmID 报警ID 按照报警先后 从0开始循环 不区分报警类型
		AlarmID uint32 `json:"alarmID"`
		// FlagStatus 标志状态 0x00-不可用 0x01-开始标志 0x02-结束标志
		FlagStatus byte `json:"alarmFlag"`
		// AlarmEventType 报警事件类型
		//0x01：前向碰撞预警
		//0x02：车道偏离报警
		//0x03：车距过近报警
		//0x04：行人碰撞报警
		//0x05：频繁变道报警
		//0x06：道路标识超限报警
		//0x07：障碍物报警
		//0x08：驾驶辅助功能失效报警
		//0x09~OxOF：用户自定义
		//0x10：道路标志识别事件
		//0x11：主动抓拍事件
		//0x12~0xFF：用户自定义
		AlarmEventType byte `json:"alarmEventType"`
		// AlarmLevel 报警级别 0x01:一级报警 0x02:二级报警
		AlarmLevel byte `json:"alarmLevel"`
		// PreVehicleSpeed 前车车速 单位:km/h 范围0-250 仅报警类型为0x01和0x03时有效 不可用时=0x00
		PreVehicleSpeed byte `json:"preVehicleSpeed"`
		// PreVehicleOrPedestrianDistance 前车或行人距离 单位100ms 范围0-100 仅报警类型0x01 0x02 0x04时有效 不可用时=0x00
		PreVehicleOrPedestrianDistance byte `json:"preVehicleOrPedestrianDistance"`
		// DeviationType 偏离类型 0x01-左侧偏离 0x02-右侧偏离 仅报警类型为0x02时有效 不可用时=0x00
		DeviationType byte `json:"deviationType"`
		// RoadSignRecognitionType 道路标志识别类型 0x01-限速 0x02-限高 0x03-限重 仅报警类型为0x06和0x10时有效 不可用时=0x00
		RoadSignRecognitionType byte `json:"roadSignRecognitionType"`
		// RoadSignRecognitionData 道路标志识别数据 识别到道路标志的数据 不可用时=0x00
		RoadSignRecognitionData byte `json:"roadSignRecognitionData"`
		// T0x0200ExtensionSBBase  苏标扩展的基础数据
		T0x0200ExtensionSBBase
	}

	// T0x0200AdditionExtension0x65 表20-驾驶员行为监测功能报警信息
	T0x0200AdditionExtension0x65 struct {
		// AlarmID 报警ID 按照报警先后 从0开始循环 不区分报警类型
		AlarmID uint32 `json:"alarmID"`
		// FlagStatus 标志状态 0x00-不可用 0x01-开始标志 0x02-结束标志
		FlagStatus byte `json:"alarmFlag"`
		// AlarmEventType 报警事件类型
		//0x01：疲劳驾驶报警
		//0x02：接打手持电话报警
		//0x03：抽烟报警
		//0x04：长时间不目视前方报警
		//0x05：未检测到驾驶员报警
		//0x06：双手同时脱离方向盘报警
		//0x07：驾驶员行为检测功能失效报警
		//0x08-0xFF: 用户自定义
		//0x10：自动抓拍事件
		//0x11：驾驶员变更事件
		//0x12~0xFF：用户自定义
		AlarmEventType byte `json:"alarmEventType"`
		// AlarmLevel 报警级别 0x01:一级报警 0x02:二级报警
		AlarmLevel byte `json:"alarmLevel"`
		// FatigueLevel 疲劳程度 1-10 数值越大越疲劳 仅在报警类型0x01生效 不可用时0x00
		FatigueLevel byte `json:"fatigueLevel"`
		// Reserved 预留 [4]byte
		Reserved [4]byte `json:"reserved"`
		// T0x0200ExtensionSBBase  苏标扩展的基础数据
		T0x0200ExtensionSBBase
	}

	// T0x0200AdditionExtension0x66 表21-轮胎状态监测报警信息
	T0x0200AdditionExtension0x66 struct {
		// AlarmID 报警ID 按照报警先后 从0开始循环 不区分报警类型
		AlarmID uint32 `json:"alarmID"`
		// FlagStatus 标志状态 0x00-不可用 0x01-开始标志 0x02-结束标志
		FlagStatus byte `json:"alarmFlag"`
		T0x0200ExtensionSBBase
		// AlarmOrEventCount 报警或事件列表总数
		AlarmOrEventCount byte `json:"alarmOrEventCount"`
		// AlarmOrEventList 报警或事件列表
		AlarmOrEventList []T0x0200ExtensionTable22 `json:"alarmOrEventList"`
	}

	// T0x0200ExtensionTable22 表22-轮胎状态监测功能报警信息列表格式
	T0x0200ExtensionTable22 struct {
		// TirePressureAlarmLocation 胎压报警位置 报警轮胎位置编号（从左前轮开始以Z字形从00依次编号 编号与是否安装TPMS无关)
		TirePressureAlarmLocation byte `json:"tirePressureAlarmLocation"`
		// AlarmOrEventType 报警或事件类型
		// 0表示无报警，1表示有报警
		// bitO：胎压（定时上报）
		// bitl：胎压过高报警
		// bit2： 胎压过低报警
		// bit3：胎温过高报警
		// bit4：传感器异常报警
		// bit5：胎压不平衡报警
		// bit6：馒漏气报警
		// bit7：电池电量低报警
		// bit8 bit15：自定义
		AlarmOrEventType uint16 `json:"alarmOrEventType"`
		// TirePressure 胎压 单位Kpa
		TirePressure uint16 `json:"tirePressure"`
		// TireTemperature 胎温 单位摄氏度
		TireTemperature uint16 `json:"tireTemperature"`
		// BatteryLevel 电池电量 单位 %
		BatteryLevel uint16 `json:"batteryLevel"`
	}

	// T0x0200AdditionExtension0x67 表23-变道决策辅助报警信息
	T0x0200AdditionExtension0x67 struct {
		// AlarmID 报警ID 按照报警先后 从0开始循环 不区分报警类型
		AlarmID uint32 `json:"alarmID"`
		// FlagStatus 标志状态 0x00-不可用 0x01-开始标志 0x02-结束标志
		FlagStatus byte `json:"alarmFlag"`
		// AlarmEventType 报警/事件类型 0x01-后方接近报警 0x02-左侧后方接近报警 0x03-右侧后方接近报警
		AlarmEventType byte `json:"alarmEventType"`
		// T0x0200ExtensionSBBase  苏标扩展的基础数据
		T0x0200ExtensionSBBase
	}
)

func (t *T0x0200AdditionExtension0x64) Parse(id uint8, content []byte) (AdditionContent, bool) {
	if id == 0x64 && len(content) == 31+16 {
		t.AlarmID = binary.BigEndian.Uint32(content[0:4])
		t.FlagStatus = content[4]
		t.AlarmEventType = content[5]
		t.AlarmLevel = content[6]
		t.PreVehicleSpeed = content[7]
		t.PreVehicleOrPedestrianDistance = content[8]
		t.DeviationType = content[9]
		t.RoadSignRecognitionType = content[10]
		t.RoadSignRecognitionData = content[11]
		t.T0x0200ExtensionSBBase.parse(content[12:47])
		return AdditionContent{
			Data:        content,
			CustomValue: t,
		}, true
	}
	return AdditionContent{}, false
}

func (t *T0x0200AdditionExtension0x65) Parse(id uint8, content []byte) (AdditionContent, bool) {
	if id == 0x65 && len(content) == 31+16 {
		t.AlarmID = binary.BigEndian.Uint32(content[0:4])
		t.FlagStatus = content[4]
		t.AlarmEventType = content[5]
		t.AlarmLevel = content[6]
		t.FatigueLevel = content[7]
		t.Reserved = [4]byte(content[8:12])
		t.T0x0200ExtensionSBBase.parse(content[12:47])
		return AdditionContent{
			Data:        content,
			CustomValue: t,
		}, true
	}
	return AdditionContent{}, false
}

func (t *T0x0200AdditionExtension0x66) Parse(id uint8, content []byte) (AdditionContent, bool) {
	if id == 0x66 && len(content) >= 40 {
		t.AlarmID = binary.BigEndian.Uint32(content[0:4])
		t.FlagStatus = content[4]
		t.T0x0200ExtensionSBBase.parse(content[5:40])
		t.AlarmOrEventCount = content[40]
		if len(content) == 40+int(t.AlarmOrEventCount)*9 {
			for i := 0; i < int(t.AlarmOrEventCount); i++ {
				start := 41 + i*9
				t.AlarmOrEventList = append(t.AlarmOrEventList, T0x0200ExtensionTable22{
					TirePressureAlarmLocation: content[start],
					AlarmOrEventType:          binary.BigEndian.Uint16(content[start+1 : start+3]),
					TirePressure:              binary.BigEndian.Uint16(content[start+3 : start+5]),
					TireTemperature:           binary.BigEndian.Uint16(content[start+5 : start+7]),
					BatteryLevel:              binary.BigEndian.Uint16(content[start+7 : start+9]),
				})
			}
			return AdditionContent{
				Data:        content,
				CustomValue: t,
			}, true
		}
	}
	return AdditionContent{}, false
}

func (t *T0x0200AdditionExtension0x67) Parse(id uint8, content []byte) (AdditionContent, bool) {
	if id == 0x67 && len(content) == 25+16 {
		t.AlarmID = binary.BigEndian.Uint32(content[0:4])
		t.FlagStatus = content[4]
		t.AlarmEventType = content[5]
		t.T0x0200ExtensionSBBase.parse(content[6:41])
		return AdditionContent{
			Data:        content,
			CustomValue: t,
		}, true
	}
	return AdditionContent{}, false
}

func (t *T0x0200AdditionExtension0x64) String() string {
	alarmEventDetails := func() string {
		infos := map[uint8]string{
			0x01: "前向碰撞预警",
			0x02: "车道偏离报警",
			0x03: "车距过近报警",
			0x04: "行人碰撞报警",
			0x05: "频繁变道报警",
			0x06: "道路标识超限报警",
			0x07: "障碍物报警",
			0x08: "驾驶辅助功能失效报警",
			0x10: "道路标志识别事件",
			0x11: "主动抓拍事件",
		}
		str := fmt.Sprintf("自定义报警类型[%d]", t.AlarmEventType)
		if v, ok := infos[t.AlarmEventType]; ok {
			str = v
		}
		return str
	}
	return strings.Join([]string{
		fmt.Sprintf("\t报警ID:[%d] 从0开始循环 不区分报警类型", t.AlarmID),
		fmt.Sprintf("\t标志状态:[%d] 0x00-不可用 0x01-开始标志 0x02-结束标志", t.FlagStatus),
		fmt.Sprintf("\t报警事件类型:[%d] %s", t.AlarmEventType, alarmEventDetails()),
		fmt.Sprintf("\t报警级别:[%d] 0x01:一级报警 0x02:二级报警", t.AlarmLevel),
		fmt.Sprintf("\t前车车速:[%d] 单位:km/h 范围0-250 仅报警类型为0x01和0x03时有效 不可用时=0x00", t.PreVehicleSpeed),
		fmt.Sprintf("\t前车或行人距离:[%d] 单位100ms 范围0-100 仅报警类型0x01 0x02 0x04时有效 不可用时=0x00", t.PreVehicleOrPedestrianDistance),
		fmt.Sprintf("\t偏离类型:[%d] 0x01-左侧偏离 0x02-右侧偏离 仅报警类型为0x02时有效 不可用时=0x00", t.DeviationType),
		fmt.Sprintf("\t道路标志识别类型:[%d] 0x01-限速 0x02-限高 0x03-限重 仅报警类型为0x06和0x10时有效 不可用时=0x00", t.RoadSignRecognitionType),
		fmt.Sprintf("\t道路标志识别数据:[%d] 识别到道路标志的数据 不可用时=0x00", t.RoadSignRecognitionData),
		t.T0x0200ExtensionSBBase.String(),
	}, "\n")
}

func (t *T0x0200AdditionExtension0x65) String() string {
	alarmEventDetails := func() string {
		infos := map[uint8]string{
			0x01: "疲劳驾驶报警",
			0x02: "接打手持电话报警",
			0x03: "抽烟报警",
			0x04: "长时间不目视前方报警",
			0x05: "未检测到驾驶员报警",
			0x06: "双手同时脱离方向盘报警",
			0x07: "驾驶员行为检测功能失效报警",
			0x10: "自动抓拍事件",
			0x11: "驾驶员变更事件",
		}
		str := fmt.Sprintf("自定义报警类型[%d]", t.AlarmEventType)
		if v, ok := infos[t.AlarmEventType]; ok {
			str = v
		}
		return str
	}
	return strings.Join([]string{
		fmt.Sprintf("\t报警ID:[%d] 从0开始循环 不区分报警类型", t.AlarmID),
		fmt.Sprintf("\t标志状态:[%d] 0x00-不可用 0x01-开始标志 0x02-结束标志", t.FlagStatus),
		fmt.Sprintf("\t报警事件类型:[%d] %s", t.AlarmEventType, alarmEventDetails()),
		fmt.Sprintf("\t报警级别:[%d] 0x01:一级报警 0x02:二级报警", t.AlarmLevel),
		fmt.Sprintf("\t疲劳程度:[%d] 单位:km/h 1-10 数值越大越疲劳 仅在报警类型0x01生效 不可用时0x00", t.FatigueLevel),
		fmt.Sprintf("\t预留:[%x]", t.Reserved),
		t.T0x0200ExtensionSBBase.String(),
	}, "\n")
}

func (t *T0x0200AdditionExtension0x66) String() string {
	str := fmt.Sprintf("\t报警或事件列表总数:[%x]\n", t.AlarmOrEventCount)
	for _, v := range t.AlarmOrEventList {
		str += fmt.Sprintf("\t\t胎压报警位置:[%d] 报警轮胎位置编号（从左前轮开始以Z字形从00依次编号 编号与是否安装TPMS无关)\n", v.TirePressureAlarmLocation)
		str += fmt.Sprintf("\t\t报警或事件类型:[%d]\n", v.AlarmOrEventType)
		str += fmt.Sprintf("\t\t胎压:[%d] 单位Kpa\n", v.TirePressure)
		str += fmt.Sprintf("\t\t胎温:[%d] 单位摄氏度\n", v.TireTemperature)
		str += fmt.Sprintf("\t\t电池电量:[%d] 单位%%\n", v.BatteryLevel)
	}

	return strings.Join([]string{
		fmt.Sprintf("\t报警ID:[%d] 从0开始循环 不区分报警类型", t.AlarmID),
		fmt.Sprintf("\t标志状态:[%d] 0x00-不可用 0x01-开始标志 0x02-结束标志", t.FlagStatus),
		t.T0x0200ExtensionSBBase.String(),
		str,
	}, "\n")
}

func (t *T0x0200AdditionExtension0x67) String() string {
	return strings.Join([]string{
		fmt.Sprintf("\t报警ID:[%d] 从0开始循环 不区分报警类型", t.AlarmID),
		fmt.Sprintf("\t标志状态:[%d] 0x00-不可用 0x01-开始标志 0x02-结束标志", t.FlagStatus),
		fmt.Sprintf("\t报警/事件类型:[%d] 0x01-后方接近报警 0x02-左侧后方接近报警 0x03-右侧后方接近报警", t.AlarmEventType),
		t.T0x0200ExtensionSBBase.String(),
	}, "\n")
}

func (t *T0x0200ExtensionSBBase) parse(data []byte) {
	t.VehicleSpeed = data[0]
	t.Altitude = binary.BigEndian.Uint16(data[1:3])
	t.Latitude = binary.BigEndian.Uint32(data[3:7])
	t.Longitude = binary.BigEndian.Uint32(data[7:11])
	t.DateTime = utils.BCD2Time(data[11:17])
	t.VehicleStatus.parse(binary.BigEndian.Uint16(data[17:19]))
	t.P9208AlarmSign.parse(data[19:])
	t.ParseSuccess = true
	return
}

func (vs *T0x0200ExtensionTable18) parse(value uint16) {
	vs.OriginalValue = value
	data := fmt.Sprintf("%.16b", vs.OriginalValue)
	if data[15] == '1' {
		vs.ACC = true
	}
	if data[14] == '1' {
		vs.LeftTurn = true
	}
	if data[13] == '1' {
		vs.RightTurn = true
	}
	if data[12] == '1' {
		vs.Wipers = true
	}
	if data[11] == '1' {
		vs.Brake = true
	}
	if data[10] == '1' {
		vs.Card = true
	}
	if data[5] == '1' {
		vs.Location = true
	}
}

func (vs *T0x0200ExtensionTable18) String() string {
	return strings.Join([]string{
		fmt.Sprintf("\t车辆状态:[%d] {", vs.OriginalValue),
		fmt.Sprintf("\t\t ACC: [%v]", vs.ACC),
		fmt.Sprintf("\t\t 左转向状态: [%v]", vs.LeftTurn),
		fmt.Sprintf("\t\t 右转向状态: [%v]", vs.RightTurn),
		fmt.Sprintf("\t\t 雨刮器状态: [%v]", vs.Wipers),
		fmt.Sprintf("\t\t 制动状态: [%v]", vs.Brake),
		fmt.Sprintf("\t\t 插卡状态: [%v]", vs.Card),
		fmt.Sprintf("\t\t 定位状态: [%v]", vs.Location),
		"\t}",
	}, "\n")
}

func (t *T0x0200ExtensionSBBase) String() string {
	return strings.Join([]string{
		fmt.Sprintf("\t车速:[%d] 单位:km/h 范围0-250", t.VehicleSpeed),
		fmt.Sprintf("\t海拔高度:[%d] 单位米(m)", t.Altitude),
		fmt.Sprintf("\t纬度:[%d] 以度为单位的纬度值乘以 10 的 6 次方，精确到百万分之一度", t.Latitude),
		fmt.Sprintf("\t经度:[%d] 以度为单位的经度值乘以 10 的 6 次方，精确到百万分之一度", t.Longitude),
		fmt.Sprintf("\t时间:[%s] 时间 YY-MM-DD-hh-mm-ss（GMT+8 时间，本标准中之后涉及的时间均采用此时区）", t.DateTime),
		t.VehicleStatus.String(),
		t.P9208AlarmSign.String(),
	}, "\n")
}
