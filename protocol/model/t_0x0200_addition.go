package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"sort"
	"strings"
)

type (
	T0x0200AdditionDetails struct {
		// Additions 附加信息
		Additions map[consts.JT808LocationAdditionType]Addition `json:"additions"`
		// 自定义解析信息
		CustomAdditionContentFunc func(id uint8, content []byte) (AdditionContent, bool) `json:"-"`
	}

	Addition struct {
		// ID 附加信息ID
		ID uint8 `json:"id"`
		// Len 附加信息长度
		Len uint8 `json:"len"`
		// Content 附加信息内容
		Content AdditionContent `json:"content"`
	}

	AdditionContent struct {
		// Data 原始数据
		Data []byte `json:"data,omitempty"`
		// CustomValue 未知数据自定义的结果
		CustomValue interface{} `json:"customValue,omitempty"`
		// Mile 里程
		Mile uint32 `json:"mile,omitempty"`
		// Oil 油量
		Oil uint16 `json:"oil,omitempty"`
		// Speed 行驶记录功能获取的速度
		Speed uint16 `json:"speed,omitempty"`
		// ManualAlarm 需要人工确认报警事件的ID
		ManualAlarm uint16 `json:"manualAlarm,omitempty"`
		// TirePressure 胎压
		TirePressure AdditionTirePressure `json:"tirePressure,omitempty"`
		// CarTemperature 车厢温度
		CarTemperature uint16 `json:"carTemperature,omitempty"`
		// OverSpeedAlarm 超速报警
		OverSpeedAlarm AdditionOverSpeedAlarm `json:"overSpeedAlarm,omitempty"`
		// AreaAlarm 进出区域/路线报警
		AreaAlarm AdditionAreaAlarm `json:"areaAlarm,omitempty"`
		// DrivingTimeInsufficientAlarm 路段行驶时间不足/过长报警 详情见表30
		DrivingTimeInsufficientAlarm AdditionDrivingTimeInsufficientAlarm `json:"drivingTimeInsufficientAlarm,omitempty"`
		// ExtendVehicleStatus 扩展车辆信号状态位 详情见表31
		ExtendVehicleStatus AdditionExtendVehicleStatus `json:"extendVehicleStatus,omitempty"`
		// IOStatus IO状态位 详情见表32
		IOStatus AdditionIOStatus `json:"ioStatus,omitempty"`
		// Analog 模拟量 bit0-15 ADD bit16-31 AD1
		Analog uint32 `json:"analog,omitempty"`
		// WIFISignalStrength 无线通信网络信号强度 数据类型位BYTE
		WIFISignalStrength uint8 `json:"WIFISignalStrength,omitempty"`
		// GNSSPositionNum GNSS定位卫星数 数据类型位BYTE
		GNSSPositionNum uint8 `json:"GNSSPositionNum,omitempty"`
	}

	AdditionTirePressure struct {
		Values map[uint8]uint8 `json:"values,omitempty"`
	}

	AdditionOverSpeedAlarm struct {
		// LocationType 位置类型 0-无特定区域 1-圆形 2-矩形 3-多边形 4-路段
		LocationType uint8 `json:"locationType,omitempty"`
		// AreaID  区域或路段ID 若位置类型为0 无该字段
		AreaID uint32 `json:"areaId,omitempty"`
	}

	AdditionAreaAlarm struct {
		// LocationType 位置类型 1-圆形 2-矩形 3-多边形 4-路线
		LocationType uint8 `json:"locationType,omitempty"`
		// AreaID 区域或线路ID
		AreaID uint32 `json:"areaId,omitempty"`
		// Direction 方向 0-进 1-出
		Direction uint8 `json:"direction,omitempty"`
	}

	AdditionDrivingTimeInsufficientAlarm struct {
		// RoadSectionID 路段ID
		RoadSectionID uint32 `json:"roadSectionId,omitempty"`
		// RoadSectionDrivingTimeSecond 路段行驶时间（单位秒)
		RoadSectionDrivingTimeSecond uint16 `json:"roadSectionDrivingTimeSecond,omitempty"`
		// Result 结果 0-不足 1-过长
		Result uint8 `json:"result,omitempty"`
	}

	AdditionExtendVehicleStatus struct {
		// Value 原始值
		Value uint32 `json:"value,omitempty"`
		// LowBeamSignal 近光灯信号
		LowBeamSignal bool `json:"lowBeamSignal,omitempty"`
		// HighBeamSignal 远光灯信号
		HighBeamSignal bool `json:"highBeamSignal,omitempty"`
		// RightTurnSignal 右转向灯信号
		RightTurnSignal bool `json:"rightTurnSignal,omitempty"`
		// LeftTurnSignal 左转向灯信号
		LeftTurnSignal bool `json:"leftTurnSignal,omitempty"`
		// BrakeSignal 制动信号
		BrakeSignal bool `json:"brakeSignal,omitempty"`
		// ReverseGearSignal 倒档信号
		ReverseGearSignal bool `json:"reverseGearSignal,omitempty"`
		// FogLightSignal 雾灯信号
		FogLightSignal bool `json:"fogLightSignal,omitempty"`
		// ClearanceLights 示廓灯
		ClearanceLights bool `json:"clearanceLights,omitempty"`
		// HornSignal 喇叭信号
		HornSignal bool `json:"hornSignal,omitempty"`
		// AirConditionerSignal 空调状态
		AirConditionerSignal bool `json:"airConditionerSignal,omitempty"`
		// NeutralSignal 空挡信号
		NeutralSignal bool `json:"neutralSignal,omitempty"`
		// RetarderWork 缓速器工作
		RetarderWork bool `json:"retarderWork,omitempty"`
		// ABSWork ABS工作
		ABSWork bool `json:"ABSWork,omitempty"`
		// HeaterWork 加热器工作
		HeaterWork bool `json:"heaterWork,omitempty"`
		// ClutchStatus 离合器状态
		ClutchStatus bool `json:"clutchStatus,omitempty"`
	}

	AdditionIOStatus struct {
		// Value 原始值
		Value uint16 `json:"value,omitempty"`
		// DeepSleepStatus 深度休眠状态
		DeepSleepStatus bool `json:"deepSleepStatus,omitempty"`
		// SleepStatus 休眠状态
		SleepStatus bool `json:"sleepStatus,omitempty"`
	}
)

func (a *T0x0200AdditionDetails) parse(body []byte) error {
	index := 0
	contrastFunc := func(id uint8, additionLen uint8) bool {
		switch id {
		case 0x01, 0x25, 0x2B:
			return additionLen == 4
		case 0x02, 0x03, 0x04, 0x06, 0x2A:
			return additionLen == 2
		case 0x05:
			return additionLen == 30
		case 0x11:
			return additionLen == 1 || additionLen == 5
		case 0x12:
			return additionLen == 6
		case 0x13:
			return additionLen == 7
		case 0x30:
			return additionLen == 1
		}
		return true
	}
	if a.Additions == nil {
		a.Additions = make(map[consts.JT808LocationAdditionType]Addition)
	}
	for index < len(body) {
		if index+2 > len(body) {
			return protocol.ErrBodyLengthInconsistency
		}
		id := body[index]
		additionLen := body[index+1]
		start := index + 2
		if ok := contrastFunc(id, additionLen); !ok {
			return protocol.ErrBodyLengthInconsistency
		}
		end := start + int(additionLen)
		if end > len(body) {
			return protocol.ErrBodyLengthInconsistency
		}
		content := body[start:end]
		a.Additions[consts.JT808LocationAdditionType(id)] = Addition{
			ID:      id,
			Len:     additionLen,
			Content: a.decode(id, content),
		}
		index = end
	}
	return nil
}

func (a *T0x0200AdditionDetails) decode(id uint8, content []byte) AdditionContent {
	if a.CustomAdditionContentFunc != nil {
		if v, ok := a.CustomAdditionContentFunc(id, content); ok {
			return v
		}
	}
	tmp := AdditionContent{
		Data: content,
	}
	switch id {
	case 0x01:
		tmp.Mile = binary.BigEndian.Uint32(content)
	case 0x02:
		tmp.Oil = binary.BigEndian.Uint16(content)
	case 0x03:
		tmp.Speed = binary.BigEndian.Uint16(content)
	case 0x04:
		tmp.ManualAlarm = binary.BigEndian.Uint16(content)
	case 0x05:
		tmp.TirePressure = a.parseTirePressure(content)
	case 0x06:
		tmp.CarTemperature = binary.BigEndian.Uint16(content)
	case 0x11:
		tmp.OverSpeedAlarm = AdditionOverSpeedAlarm{
			LocationType: content[0],
		}
		if content[0] != 0 {
			tmp.OverSpeedAlarm.AreaID = binary.BigEndian.Uint32(content)
		}
	case 0x12:
		tmp.AreaAlarm = AdditionAreaAlarm{
			LocationType: content[0],
			AreaID:       binary.BigEndian.Uint32(content[1:5]),
			Direction:    content[5],
		}
	case 0x13:
		tmp.DrivingTimeInsufficientAlarm = AdditionDrivingTimeInsufficientAlarm{
			RoadSectionID:                binary.BigEndian.Uint32(content[0:4]),
			RoadSectionDrivingTimeSecond: binary.BigEndian.Uint16(content[4:6]),
			Result:                       content[6],
		}
	case 0x25:
		tmp.ExtendVehicleStatus = a.parseExtendVehicleStatus(binary.BigEndian.Uint32(content))
	case 0x2A:
		tmp.IOStatus = a.parseIOStatus(binary.BigEndian.Uint16(content))
	case 0x2B:
		tmp.Analog = binary.BigEndian.Uint32(content)
	case 0x30:
		tmp.WIFISignalStrength = content[0]
	case 0x31:
		tmp.GNSSPositionNum = content[0]
	default:
	}
	return tmp
}

func (a *T0x0200AdditionDetails) parseExtendVehicleStatus(value uint32) AdditionExtendVehicleStatus {
	tmp := AdditionExtendVehicleStatus{
		Value: value,
	}
	data := fmt.Sprintf("%.32b", value)
	if data[31] == '1' {
		tmp.LowBeamSignal = true
	}
	if data[30] == '1' {
		tmp.HighBeamSignal = true
	}
	if data[29] == '1' {
		tmp.RightTurnSignal = true
	}
	if data[28] == '1' {
		tmp.LeftTurnSignal = true
	}
	if data[27] == '1' {
		tmp.BrakeSignal = true
	}
	if data[26] == '1' {
		tmp.ReverseGearSignal = true
	}
	if data[25] == '1' {
		tmp.FogLightSignal = true
	}
	if data[24] == '1' {
		tmp.ClearanceLights = true
	}
	if data[23] == '1' {
		tmp.HornSignal = true
	}
	if data[22] == '1' {
		tmp.AirConditionerSignal = true
	}
	if data[21] == '1' {
		tmp.NeutralSignal = true
	}
	if data[20] == '1' {
		tmp.RetarderWork = true
	}
	if data[19] == '1' {
		tmp.ABSWork = true
	}
	if data[18] == '1' {
		tmp.HeaterWork = true
	}
	if data[17] == '1' {
		tmp.ClutchStatus = true
	}
	return tmp
}

func (a *T0x0200AdditionDetails) parseIOStatus(value uint16) AdditionIOStatus {
	tmp := AdditionIOStatus{
		Value: value,
	}
	data := fmt.Sprintf("%.16b", value)
	if data[15] == '1' {
		tmp.DeepSleepStatus = true
	}
	if data[14] == '1' {
		tmp.SleepStatus = true
	}
	return tmp
}

func (a *T0x0200AdditionDetails) String() string {
	mile := a.Additions[consts.A0x01Mile].Content.Mile
	oil := a.Additions[consts.A0x02Oil].Content.Oil
	speed := a.Additions[consts.A0x03Speed].Content.Speed
	alarmID := a.Additions[consts.A0x04ManualAlarm].Content.ManualAlarm
	carTemperature := a.Additions[consts.A0x06CarTemperature].Content.CarTemperature
	analog := a.Additions[consts.A0x2BAnalog].Content.Analog
	wifi := a.Additions[consts.A0x30WIFISignalStrength].Content.WIFISignalStrength
	num := a.Additions[consts.A0x31GNSSPositionNum].Content.GNSSPositionNum
	unknown := ""
	for id, addition := range a.Additions {
		if (id >= 0x07 && id <= 0x0f) || (id >= 0x14 && id <= 0x24) || (id > 0x31) {
			unknown += fmt.Sprintf("\t\t[%02x]未知附加信息[%d] data=[%x]\n", uint8(id), id, addition.Content.Data)
		}
	}
	str := strings.Join([]string{
		"\t{",
		fmt.Sprintf("\t\t[01]附加信息ID:1 里程"),
		fmt.Sprintf("\t\t[04]附加信息长度:4"),
		fmt.Sprintf("\t\t[%08x]里程:[%d]", mile, mile),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[02]附加信息ID:2"),
		fmt.Sprintf("\t\t[02]附加信息长度:2"),
		fmt.Sprintf("\t\t[%04x]油量:[%d]", oil, oil),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[03]附加信息ID:3"),
		fmt.Sprintf("\t\t[02]附加信息长度:2"),
		fmt.Sprintf("\t\t[%04x]行驶记录功能获取速度:[%d]", speed, speed),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[04]附加信息ID:4"),
		fmt.Sprintf("\t\t[02]附加信息长度:2"),
		fmt.Sprintf("\t\t[%04x]需要人工确认报警事件ID:[%d]", alarmID, alarmID),
		"\t}",
		a.Additions[consts.A0x05TirePressure].Content.TirePressure.String(),
		"\t{",
		fmt.Sprintf("\t\t[06]附加信息ID:6 2019版本新增"),
		fmt.Sprintf("\t\t[02]附加信息长度:2"),
		fmt.Sprintf("\t\t[%04x]车厢温度:[%d]", carTemperature, carTemperature),
		"\t}",
		a.Additions[consts.A0x11OverSpeedAlarm].Content.OverSpeedAlarm.String(),
		a.Additions[consts.A0x12AreaAlarm].Content.AreaAlarm.String(),
		a.Additions[consts.A0x13DrivingTimeInsufficientAlarm].Content.DrivingTimeInsufficientAlarm.String(),
		a.Additions[consts.A0x25ExtendVehicleStatus].Content.ExtendVehicleStatus.String(),
		a.Additions[consts.A0x2AIOStatus].Content.IOStatus.String(),
		"\t{",
		fmt.Sprintf("\t\t[2b]附加信息ID:43"),
		fmt.Sprintf("\t\t[04]附加信息长度:4"),
		fmt.Sprintf("\t\t[%08x]模拟量:[%d]", analog, analog),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[30]附加信息ID:48"),
		fmt.Sprintf("\t\t[01]附加信息长度:1"),
		fmt.Sprintf("\t\t[%02x]无线通信网络信号强度:[%d]", wifi, wifi),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[31]附加信息ID:49"),
		fmt.Sprintf("\t\t[01]附加信息长度:1"),
		fmt.Sprintf("\t\t[%02x]GNSS定位卫星:[%d]", num, num),
		"\t}",
	}, "\n")
	if unknown != "" {
		str += strings.Join([]string{
			"\n\t{",
			unknown,
			"\t}",
		}, "\n")
	}
	return str
}

func (a *T0x0200AdditionDetails) parseTirePressure(content []byte) AdditionTirePressure {
	tmp := AdditionTirePressure{
		Values: map[uint8]uint8{},
	}
	for k, v := range content {
		if v != 0x00 {
			tmp.Values[uint8(k)] = v
		}
	}
	return tmp
}

func (a AdditionTirePressure) String() string {
	str := "\t\t胎压 单位为Pa 2019版本新增\n"
	str += "\t\t[05]附加信息ID:5\n"
	str += "\t\t[1e]附加信息长度:30\n"
	values := make([]string, 0, len(a.Values))
	for k, v := range a.Values {
		values = append(values, fmt.Sprintf("\t\t[%02x]轮胎%d胎压:[%d]", v, k, v))
	}
	sort.Strings(values)
	str = strings.TrimRight(str, "\n")
	return strings.Join([]string{
		"\t{",
		str,
		strings.Join(values, "\n"),
		"\t}",
	}, "\n")
}

func (a AdditionOverSpeedAlarm) String() string {
	str := fmt.Sprintf("\t\t[%02x]位置类型:[%d] 0-无特定区域 1-圆形 2-矩形 3-多边形 4-路段", a.LocationType, a.LocationType)
	if a.LocationType != 0x00 {
		str += fmt.Sprintf("\t\t[%08x]区域或路段ID:[%d]", a.AreaID, a.AreaID)
	}
	return strings.Join([]string{
		"\t{",
		fmt.Sprintf("\t\t[11]附加信息ID:17 超速报警 详情见表28"),
		fmt.Sprintf("\t\t[01]附加信息长度:1"),
		str,
		"\t}",
	}, "\n")
}

func (a AdditionAreaAlarm) String() string {
	return strings.Join([]string{
		"\t{",
		fmt.Sprintf("\t\t[12]附加信息ID:18 进出区域/路线报警 详情见表29"),
		fmt.Sprintf("\t\t[06]附加信息长度:6"),
		fmt.Sprintf("\t\t[%02x]位置类型:[%d] 1-圆形 2-矩形 3-多边形 4-路线", a.LocationType, a.LocationType),
		fmt.Sprintf("\t\t[%08x]区域或路段ID:[%d]", a.AreaID, a.AreaID),
		fmt.Sprintf("\t\t[%02x]方向:[%d]", a.Direction, a.Direction),
		"\t}",
	}, "\n")
}

func (a AdditionDrivingTimeInsufficientAlarm) String() string {
	return strings.Join([]string{
		"\t{",
		fmt.Sprintf("\t\t[13]附加信息ID:37 路段行驶时间不足/过长报警 详情见表30"),
		fmt.Sprintf("\t\t[07]附加信息长度:7"),
		fmt.Sprintf("\t\t[%08x]路段ID:[%d]", a.RoadSectionID, a.RoadSectionID),
		fmt.Sprintf("\t\t[%04x]路段行驶时间 单位秒:[%d]", a.RoadSectionDrivingTimeSecond, a.RoadSectionDrivingTimeSecond),
		fmt.Sprintf("\t\t[%02x]结果:[%d]", a.Result, a.Result),
		"\t}",
	}, "\n")
}

func (a AdditionExtendVehicleStatus) String() string {
	return strings.Join([]string{
		"\t{",
		fmt.Sprintf("\t\t[25]附加信息ID:37 扩展车辆信号状态码 详情见表31"),
		fmt.Sprintf("\t\t[04]附加信息长度:4"),
		fmt.Sprintf("\t\t[%032b]扩展车辆信号状态位:[%d]", a.Value, a.Value),
		fmt.Sprintf("\t\t[bit15-31]保留:[%s]", fmt.Sprintf("%032b", a.Value)[:16]),
		fmt.Sprintf("\t\t[bit14]离合器状态:[%t]", a.ClutchStatus),
		fmt.Sprintf("\t\t[bit13]加热器工作:[%t]", a.HeaterWork),
		fmt.Sprintf("\t\t[bit12]ABS工作:[%t]", a.ABSWork),
		fmt.Sprintf("\t\t[bit11]缓速器工作:[%t]", a.RetarderWork),
		fmt.Sprintf("\t\t[bit10]空挡信号:[%t]", a.NeutralSignal),
		fmt.Sprintf("\t\t[bit9]空调状态:[%t]", a.AirConditionerSignal),
		fmt.Sprintf("\t\t[bit8]喇叭信号:[%t]", a.HornSignal),
		fmt.Sprintf("\t\t[bit7]示廓灯:[%t]", a.ClearanceLights),
		fmt.Sprintf("\t\t[bit6]雾灯信号:[%t]", a.FogLightSignal),
		fmt.Sprintf("\t\t[bit5]倒挡信号:[%t]", a.ReverseGearSignal),
		fmt.Sprintf("\t\t[bit4]制动信号:[%t]", a.BrakeSignal),
		fmt.Sprintf("\t\t[bit3]左转向灯信号:[%t]", a.LeftTurnSignal),
		fmt.Sprintf("\t\t[bit2]右转向灯信号:[%t]", a.RightTurnSignal),
		fmt.Sprintf("\t\t[bit1]远光灯信号:[%t]", a.HighBeamSignal),
		fmt.Sprintf("\t\t[bit0]近光灯信号:[%t]", a.LowBeamSignal),
		"\t}",
	}, "\n")
}

func (a AdditionIOStatus) String() string {
	return strings.Join([]string{
		"\t{",
		fmt.Sprintf("\t\t[2A]附加信息ID:42 IO状态 详情见表32"),
		fmt.Sprintf("\t\t[02]附加信息长度:2"),
		fmt.Sprintf("\t\t[%016b]IO状态位:[%d]", a.Value, a.Value),
		fmt.Sprintf("\t\t[bit2-15]保留:[%s]", fmt.Sprintf("%016b", a.Value)[:14]),
		fmt.Sprintf("\t\t[bit1]休眠状态:[%t]", a.SleepStatus),
		fmt.Sprintf("\t\t[bit0]深度休眠状态:[%t]", a.DeepSleepStatus),
		"\t}",
	}, "\n")
}
