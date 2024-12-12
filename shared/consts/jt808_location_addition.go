package consts

// JT808LocationAdditionType 终端-位置附加信息.
type JT808LocationAdditionType uint8

const (
	// A0x01Mile 里程.
	A0x01Mile JT808LocationAdditionType = 0x01
	// A0x02Oil 油量.
	A0x02Oil JT808LocationAdditionType = 0x02
	// A0x03Speed 行驶记录功能获取的速度.
	A0x03Speed JT808LocationAdditionType = 0x03
	// A0x04ManualAlarm 需要人工确认报警事件的ID.
	A0x04ManualAlarm JT808LocationAdditionType = 0x04
	// A0x05TirePressure 胎压 2019版本新增.
	A0x05TirePressure JT808LocationAdditionType = 0x05
	// A0x06CarTemperature 车厢温度 2019版本新增.
	A0x06CarTemperature JT808LocationAdditionType = 0x06
	// A0x11OverSpeedAlarm 超速报警 详情见表28.
	A0x11OverSpeedAlarm JT808LocationAdditionType = 0x11
	// A0x12AreaAlarm 进出区域/路线报警 详情见表29.
	A0x12AreaAlarm JT808LocationAdditionType = 0x12
	// A0x13DrivingTimeInsufficientAlarm 路段行驶时间不足/过长报警 详情见表30.
	A0x13DrivingTimeInsufficientAlarm JT808LocationAdditionType = 0x13
	// A0x25ExtendVehicleStatus 扩展车辆信号状态位 详情见表31.
	A0x25ExtendVehicleStatus JT808LocationAdditionType = 0x25
	// A0x2AIOStatus IO状态位 详情见表32.
	A0x2AIOStatus JT808LocationAdditionType = 0x2A
	// A0x2BAnalog 模拟量 bit0-15 ADD bit16-31 AD1.
	A0x2BAnalog JT808LocationAdditionType = 0x2B
	// A0x30WIFISignalStrength 无线通信网络信号强度 数据类型位BYTE.
	A0x30WIFISignalStrength JT808LocationAdditionType = 0x30
	// A0x31GNSSPositionNum GNSS定位卫星数 数据类型位BYTE.
	A0x31GNSSPositionNum JT808LocationAdditionType = 0x31
	// A0xE0Custom 厂商自定义.
	A0xE0Custom JT808LocationAdditionType = 0xE0
)

func (t JT808LocationAdditionType) String() string {
	switch t {
	case A0x01Mile:
		return "里程"
	case A0x02Oil:
		return "油量"
	case A0x03Speed:
		return "行驶记录功能获取的速度"
	case A0x04ManualAlarm:
		return "需要人工确认报警事件的ID"
	case A0x05TirePressure:
		return "胎压"
	case A0x06CarTemperature:
		return "车厢温度"
	case A0x11OverSpeedAlarm:
		return "超速报警"
	case A0x12AreaAlarm:
		return "进出区域/路线报警"
	case A0x13DrivingTimeInsufficientAlarm:
		return "路段行驶时间不足/过长报警"
	case A0x25ExtendVehicleStatus:
		return "扩展车辆信号状态位"
	case A0x2AIOStatus:
		return "IO状态位"
	case A0x2BAnalog:
		return "模拟量"
	case A0x30WIFISignalStrength:
		return "无线通信网络信号强度"
	case A0x31GNSSPositionNum:
		return "GNSS定位卫星数"
	case A0xE0Custom:
		return "厂商自定义"
	}

	return "非标准的附加信息"
}
