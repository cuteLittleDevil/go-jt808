package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"strings"
)

type (
	T0x0200AdditionExtension0x64 struct {
		// AlarmID 报警ID 按照报警先后 从0开始循环 不区分报警类型
		AlarmID uint32 `json:"alarmID"`
		// FlagStatus 标志状态 0x00-不可用 0x01-开始标志 0x02-结束标志
		FlagStatus byte `json:"alarmFlag"`
		// AlarmEventType 报警事件类型 0x01：前向碰撞预警
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
		VehicleStatus VehicleStatus `json:"vehicleStatus"`
		// P9208AlarmSign 报警标识号 苏标
		P9208AlarmSign `json:"p9208AlarmSign"`
	}

	VehicleStatus struct {
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
)

func (vs *VehicleStatus) parse(value uint16) {
	vs.OriginalValue = value
	data := fmt.Sprintf("%.32b", vs.OriginalValue)
	if data[31] == '1' {
		vs.ACC = true
	}
	if data[30] == '1' {
		vs.LeftTurn = true
	}
	if data[29] == '1' {
		vs.RightTurn = true
	}
	if data[28] == '1' {
		vs.Wipers = true
	}
	if data[27] == '1' {
		vs.Brake = true
	}
	if data[26] == '1' {
		vs.Card = true
	}
	if data[21] == '1' {
		vs.Location = true
	}
}

func (vs *VehicleStatus) String() string {
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

func (t T0x0200AdditionExtension0x64) Parse(id uint8, content []byte) (AdditionContent, bool) {
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
		t.VehicleSpeed = content[12]
		t.Altitude = binary.BigEndian.Uint16(content[13:15])
		t.Latitude = binary.BigEndian.Uint32(content[15:19])
		t.Longitude = binary.BigEndian.Uint32(content[19:23])
		t.DateTime = utils.BCD2Time(content[23:29])
		t.VehicleStatus.parse(binary.BigEndian.Uint16(content[29:31]))
		t.P9208AlarmSign.parse(content[31:])
		return AdditionContent{
			Data:        content,
			CustomValue: t,
		}, true
	}
	return AdditionContent{}, false
}

func (t T0x0200AdditionExtension0x64) String() string {
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
		fmt.Sprintf("\t车速:[%d] 单位:km/h 范围0-250", t.VehicleSpeed),
		fmt.Sprintf("\t海拔高度:[%d] 单位米(m)", t.Altitude),
		fmt.Sprintf("\t纬度:[%d] 以度为单位的纬度值乘以 10 的 6 次方，精确到百万分之一度", t.Latitude),
		fmt.Sprintf("\t经度:[%d] 以度为单位的经度值乘以 10 的 6 次方，精确到百万分之一度", t.Longitude),
		fmt.Sprintf("\t时间:[%s] 时间 YY-MM-DD-hh-mm-ss（GMT+8 时间，本标准中之后涉及的时间均采用此时区）", t.DateTime),
		t.VehicleStatus.String(),
		t.P9208AlarmSign.String(),
	}, "\n")
}
