package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"strings"
)

type (
	T0x0200LocationItem struct {
		// AlarmSign 报警标志 JT808.Protocol.Enums.JT808Alarm
		AlarmSign uint32 `json:"alarmSign"`
		// StatusSign 状态标志 JT808.Protocol.Enums.JT808Status
		StatusSign uint32 `json:"statusSign"`
		// Latitude 纬度 以度为单位的纬度值乘以 10 的 6 次方，精确到百万分之一度
		Latitude uint32 `json:"latitude"`
		// Longitude 经度 以度为单位的经度值乘以 10 的 6 次方，精确到百万分之一度
		Longitude uint32 `json:"longitude"`
		// Altitude 海拔高度 单位为米
		Altitude uint16 `json:"altitude"`
		// Speed 速度 1/10km/h
		Speed uint16 `json:"speed"`
		// Direction 方向 0-359，正北为 0，顺时针
		Direction uint16 `json:"direction"`
		// DateTime 时间 YY-MM-DD-hh-mm-ss（GMT+8 时间，本标准中之后涉及的时间均采用此时区）
		DateTime string `json:"dateTime"`
		// AlarmSignDetails 报警标志详情描述
		AlarmSignDetails AlarmSignDetails `json:"alarmSignDetails"`
		// StatusSignDetails 状态标志详情描述
		StatusSignDetails StatusSignDetails `json:"statusSignDetails"`
	}

	AlarmSignDetails struct {
		// EmergencyAlarm 报警类型-紧急报警,触动报警开关后触发
		EmergencyAlarm bool `json:"emergencyAlarm,omitempty,omitempty,omitempty,omitempty"`
		// OverSpeed 报警类型-超速
		OverSpeed bool `json:"overSpeed,omitempty,omitempty,omitempty,omitempty"`
		// FatigueDriving 报警类型-疲劳驾驶
		FatigueDriving bool `json:"fatigueDriving,omitempty,omitempty,omitempty,omitempty"`
		// DangerousAlarm 报警类型-危险预警
		DangerousAlarm bool `json:"dangerousAlarm,omitempty,omitempty,omitempty,omitempty"`
		// GNSSModuleFault 报警类型-GNSS模块发⽣故障
		GNSSModuleFault bool `json:"GNSSModuleFault,omitempty,omitempty,omitempty,omitempty"`
		// GNSSAntennaFault 报警类型-GNSS天线未接或被剪断
		GNSSAntennaFault bool `json:"GNSSAntennaFault,omitempty,omitempty,omitempty,omitempty"`
		// GNSSAntennaShortCircuit GNSS天线短路
		GNSSAntennaShortCircuit bool `json:"GNSSAntennaShortCircuit,omitempty,omitempty,omitempty,omitempty"`
		// TerminalPowerSupply 报警类型-终端主电源⽋压
		TerminalPowerSupply bool `json:"terminalPowerSupply,omitempty,omitempty,omitempty,omitempty"`
		// TerminalPowerSupplyShutdown 报警类型-终端主电源掉电
		TerminalPowerSupplyShutdown bool `json:"terminalPowerSupplyShutdown,omitempty,omitempty,omitempty,omitempty"`
		// TerminalLCDFault 报警类型-终端LCD或显示器故障
		TerminalLCDFault bool `json:"terminalLCDFault,omitempty,omitempty,omitempty"`
		// TTSModuleFault 报警类型-TTS模块故障
		TTSModuleFault bool `json:"TTSModuleFault,omitempty,omitempty,omitempty"`
		// CameraFault 报警类型-摄像头故障
		CameraFault bool `json:"cameraFault,omitempty,omitempty,omitempty"`
		// ICCardModuleFault 报警类型-道路运输证IC卡模块故障
		ICCardModuleFault bool `json:"ICCardModuleFault,omitempty,omitempty,omitempty"`
		// OverSpeedAlarm 报警类型-超速预警
		OverSpeedAlarm bool `json:"overSpeedAlarm,omitempty,omitempty,omitempty"`
		// FatigueDrivingAlarm 报警类型-疲劳驾驶预警
		FatigueDrivingAlarm bool `json:"fatigueDrivingAlarm,omitempty,omitempty,omitempty"`
		// ViolationDrivingAlarm 报警类型-违规行驶预警
		ViolationDrivingAlarm bool `json:"violationDrivingAlarm,omitempty,omitempty,omitempty"`
		// TirePressureAlarm 报警类型-胎压预警
		TirePressureAlarm bool `json:"tirePressureAlarm,omitempty,omitempty"`
		// RightTurnBlindAreaAlarm 报警类型-右转盲区预警
		RightTurnBlindAreaAlarm bool `json:"rightTurnBlindAreaAlarm,omitempty,omitempty"`
		// DrivingTimeout 报警类型-当天累计驾驶超时
		DrivingTimeout bool `json:"drivingTimeout,omitempty,omitempty"`
		// OverTimeStop 报警类型-超时停⻋
		OverTimeStop bool `json:"overTimeStop,omitempty,omitempty"`
		// InOutArea 报警类型-进出区域
		InOutArea bool `json:"inOutArea,omitempty,omitempty"`
		// InOutLine 报警类型-进出路线
		InOutLine bool `json:"inOutLine,omitempty,omitempty"`
		// SectionDrivingTime 报警类型-路段⾏驶时间不⾜/过⻓
		SectionDrivingTime bool `json:"sectionDrivingTime,omitempty"`
		// LineDeviation 报警类型-路线偏离报警
		LineDeviation bool `json:"lineDeviation,omitempty"`
		// VSSFault 报警类型-⻋辆VSS故障
		VSSFault bool `json:"vssFault,omitempty"`
		// OilLevelAbnormality 报警类型-⻋辆油量异常
		OilLevelAbnormality bool `json:"oilLevelAbnormality,omitempty"`
		// StealCar 报警类型-⻋辆被盗(通过⻋辆防盗器)
		StealCar bool `json:"stealCar,omitempty"`
		// LaneDeviation 报警类型-⻋辆⾮法点⽕
		LaneDeviation bool `json:"laneDeviation,omitempty"`
		// LaneOffset 报警类型-⻋辆⾮法位移
		LaneOffset bool `json:"laneOffset,omitempty"`
		// CollisionAlarm 报警类型-碰撞预警
		CollisionAlarm bool `json:"collisionAlarm,omitempty"`
		// SideSlipAlarm 报警类型-侧翻预警
		SideSlipAlarm bool `json:"sideSlipAlarm,omitempty"`
		// LaneOpeningAlarm 报警类型-⾮法开⻔报警
		LaneOpeningAlarm bool `json:"laneOpeningAlarm,omitempty"`
	}

	StatusSignDetails struct {
		// ACC ACC 0-关 1-开
		ACC bool `json:"acc,omitempty"`
		// Location 定位状态 0-未定位 1-定位
		Location bool `json:"location,omitempty"`
		// South 南北纬 0-北纬 1-南纬
		South bool `json:"south,omitempty"`
		// East 东西经 0-东经 1-西经
		East bool `json:"east,omitempty"`
		// Suspended 运营状态 0-运营 1-停运
		Suspended bool `json:"suspended,omitempty"`
		// Encryption 是否加密 0-不加密 1-加密
		Encryption bool `json:"encryption,omitempty"`
		// EmergencyBrake 紧急刹车系统的前撞预警
		EmergencyBrake bool `json:"emergencyBrake,omitempty"`
		// LaneOffset 车道偏移预警
		LaneOffset bool `json:"laneOffset,omitempty"`
		// Cargo 载客情况 00-空车 01-半载 10-保留 11-满载
		Cargo uint8 `json:"cargo,omitempty"`
		// Oil 油路情况 0-正常 1-断开
		Oil bool `json:"oil,omitempty"`
		// Electricity 电路情况 0-正常 1-断开
		Electricity bool `json:"electricity,omitempty"`
		// VehicleDoor 车门情况 0-解锁 1-加锁
		VehicleDoor bool `json:"vehicleDoor,omitempty"`
		// FrontDoor 前门情况 0-门1关 1-门1开
		FrontDoor bool `json:"frontDoor,omitempty"`
		// MiddleDoor 中门情况 0-门2关 1-门2开
		MiddleDoor bool `json:"middleDoor,omitempty"`
		// BackDoor 后门情况 0-门3关 1-门3开
		BackDoor bool `json:"backDoor,omitempty"`
		// DriverDoor 驾驶席门 0-门4关 1-门4开
		DriverDoor bool `json:"driverDoor,omitempty"`
		// CustomDoor 自定义门 0-门5关 1-门5开
		CustomDoor bool `json:"customDoor,omitempty"`
		// UseGPS 是否使用GPS卫星定位 0-不使用 1-使用
		UseGPS bool `json:"useGPS,omitempty"`
		// UseBD 是否使用北斗卫星定位 0-不使用 1-使用
		UseBD bool `json:"useBD,omitempty"`
		// UseGLONASS 是否使用GLONASS卫星定位 0-不使用 1-使用
		UseGLONASS bool `json:"useGLONASS,omitempty"`
		// UseGalileo 是否使用Galileo卫星定位 0-不使用 1-使用
		UseGalileo bool `json:"useGalileo,omitempty"`
		// VehicleRunning 车辆状态 0-停止 1-行驶 2019版本增加的
		VehicleRunning bool `json:"vehicleRunning,omitempty"`
	}
)

func (tl *T0x0200LocationItem) parse(body []byte) error {
	if len(body) < 28 {
		return protocol.ErrBodyLengthInconsistency
	}
	tl.AlarmSign = binary.BigEndian.Uint32(body[:4])
	tl.AlarmSignDetails.parse(tl.AlarmSign)
	tl.StatusSign = binary.BigEndian.Uint32(body[4:8])
	tl.StatusSignDetails.parse(tl.StatusSign)
	tl.Latitude = binary.BigEndian.Uint32(body[8:12])
	tl.Longitude = binary.BigEndian.Uint32(body[12:16])
	tl.Altitude = binary.BigEndian.Uint16(body[16:18])
	tl.Speed = binary.BigEndian.Uint16(body[18:20])
	tl.Direction = binary.BigEndian.Uint16(body[20:22])
	tl.DateTime = utils.BCD2Time(body[22:28])
	return nil
}

func (tl *T0x0200LocationItem) encode() []byte {
	data := make([]byte, 22, 30)
	binary.BigEndian.PutUint32(data[:4], tl.AlarmSign)
	binary.BigEndian.PutUint32(data[4:8], tl.StatusSign)
	binary.BigEndian.PutUint32(data[8:12], tl.Latitude)
	binary.BigEndian.PutUint32(data[12:16], tl.Longitude)
	binary.BigEndian.PutUint16(data[16:18], tl.Altitude)
	binary.BigEndian.PutUint16(data[18:20], tl.Speed)
	binary.BigEndian.PutUint16(data[20:22], tl.Direction)
	bcdTime := strings.ReplaceAll(tl.DateTime, "-", "")
	bcdTime = strings.ReplaceAll(bcdTime, ":", "")
	bcdTime = strings.ReplaceAll(bcdTime, " ", "")
	if len(bcdTime) == 14 {
		bcdTime = bcdTime[2:]
	}
	data = append(data, utils.Time2BCD(bcdTime)...)
	return data
}

func (tl *T0x0200LocationItem) String() string {
	body := tl.encode()
	return strings.Join([]string{
		fmt.Sprintf("\t[%08x] 报警标志:[%d]", tl.AlarmSign, tl.AlarmSign),
		fmt.Sprintf("\t[%08x] 状态标志:[%d]", tl.StatusSign, tl.StatusSign),
		fmt.Sprintf("\t[%08x] 纬度:[%d]", tl.Latitude, tl.Latitude),
		fmt.Sprintf("\t[%08x] 经度:[%d]", tl.Longitude, tl.Longitude),
		fmt.Sprintf("\t[%04x] 海拔高度:[%d]", tl.Altitude, tl.Altitude),
		fmt.Sprintf("\t[%04x] 速度:[%d]", tl.Speed, tl.Speed),
		fmt.Sprintf("\t[%04x] 方向:[%d]", tl.Direction, tl.Direction),
		fmt.Sprintf("\t[%x] 时间:[%s]", body[22:28], tl.DateTime),
	}, "\n")
}

func (a *AlarmSignDetails) parse(alarmSign uint32) {
	data := fmt.Sprintf("%.32b", alarmSign)
	if data[31] == '1' {
		a.EmergencyAlarm = true
	}
	if data[30] == '1' {
		a.OverSpeed = true
	}
	if data[29] == '1' {
		a.FatigueDriving = true
	}
	if data[28] == '1' {
		a.DangerousAlarm = true
	}
	if data[27] == '1' {
		a.GNSSModuleFault = true
	}
	if data[26] == '1' {
		a.GNSSAntennaFault = true
	}
	if data[25] == '1' {
		a.GNSSAntennaShortCircuit = true
	}
	if data[24] == '1' {
		a.TerminalPowerSupply = true
	}
	if data[23] == '1' {
		a.TerminalPowerSupplyShutdown = true
	}
	if data[22] == '1' {
		a.TerminalLCDFault = true
	}
	if data[21] == '1' {
		a.TTSModuleFault = true
	}
	if data[20] == '1' {
		a.CameraFault = true
	}
	if data[19] == '1' {
		a.ICCardModuleFault = true
	}
	if data[18] == '1' {
		a.OverSpeedAlarm = true
	}
	if data[17] == '1' {
		a.FatigueDrivingAlarm = true
	}
	if data[16] == '1' {
		a.ViolationDrivingAlarm = true
	}
	if data[15] == '1' {
		a.TirePressureAlarm = true
	}
	if data[14] == '1' {
		a.RightTurnBlindAreaAlarm = true
	}
	if data[13] == '1' {
		a.DrivingTimeout = true
	}
	if data[12] == '1' {
		a.OverTimeStop = true
	}
	if data[11] == '1' {
		a.InOutArea = true
	}
	if data[10] == '1' {
		a.InOutLine = true
	}
	if data[9] == '1' {
		a.SectionDrivingTime = true
	}
	if data[8] == '1' {
		a.LineDeviation = true
	}
	if data[7] == '1' {
		a.VSSFault = true
	}
	if data[6] == '1' {
		a.OilLevelAbnormality = true
	}
	if data[5] == '1' {
		a.StealCar = true
	}
	if data[4] == '1' {
		a.LaneDeviation = true
	}
	if data[3] == '1' {
		a.LaneOffset = true
	}
	if data[2] == '1' {
		a.CollisionAlarm = true
	}
	if data[1] == '1' {
		a.SideSlipAlarm = true
	}
	if data[0] == '1' {
		a.LaneOpeningAlarm = true
	}
}

func (a *AlarmSignDetails) String() string {
	return strings.Join([]string{
		fmt.Sprintf("\t\t[bit31]非法开门报警:[%t]", a.LaneOpeningAlarm),
		fmt.Sprintf("\t\t[bit30]侧翻预警:[%t]", a.SideSlipAlarm),
		fmt.Sprintf("\t\t[bit29]碰撞预警[%t]", a.CollisionAlarm),
		fmt.Sprintf("\t\t[bit28]车辆非法位移:[%t]", a.LaneOffset),
		fmt.Sprintf("\t\t[bit27]车辆非法点火:[%t]", a.LaneDeviation),
		fmt.Sprintf("\t\t[bit26]车辆被盗(通过车辆防盗器):[%t]", a.StealCar),
		fmt.Sprintf("\t\t[bit25]车辆油量异常:[%t]", a.OilLevelAbnormality),
		fmt.Sprintf("\t\t[bit24]车辆VSS故障:[%t]", a.VSSFault),
		fmt.Sprintf("\t\t[bit23]路线偏离报警:[%t]", a.LineDeviation),
		fmt.Sprintf("\t\t[bit22]路段行驶时间不足/过长:[%t]", a.SectionDrivingTime),
		fmt.Sprintf("\t\t[bit21]进出路线:[%t]", a.InOutLine),
		fmt.Sprintf("\t\t[bit20]进出区域:[%t]", a.InOutArea),
		fmt.Sprintf("\t\t[bit19]超时停车:[%t]", a.OverTimeStop),
		fmt.Sprintf("\t\t[bit18]当天累计驾驶超时:[%t]", a.DrivingTimeout),
		fmt.Sprintf("\t\t[bit17]:右转盲区预警[%t]", a.RightTurnBlindAreaAlarm),
		fmt.Sprintf("\t\t[bit16]:胎压预警[%t]", a.TirePressureAlarm),
		fmt.Sprintf("\t\t[bit15]:违规行驶预警[%t]", a.ViolationDrivingAlarm),
		fmt.Sprintf("\t\t[bit14]疲劳驾驶预警:[%t]", a.FatigueDrivingAlarm),
		fmt.Sprintf("\t\t[bit13]超速预警:[%t]", a.OverSpeedAlarm),
		fmt.Sprintf("\t\t[bit12]道路运输证IC卡模块故障:[%t]", a.ICCardModuleFault),
		fmt.Sprintf("\t\t[bit11]摄像头故障:[%t]", a.CameraFault),
		fmt.Sprintf("\t\t[bit10]TTS模块故障]:[%t]", a.TTSModuleFault),
		fmt.Sprintf("\t\t[bit9]终端LCD或显示器故障:[%t]", a.TerminalLCDFault),
		fmt.Sprintf("\t\t[bit8]终端主电源掉电:[%t]", a.TerminalPowerSupplyShutdown),
		fmt.Sprintf("\t\t[bit7]终端主电源欠压:[%t]", a.TerminalPowerSupply),
		fmt.Sprintf("\t\t[bit6]GNSS天线短路:[%t]", a.GNSSAntennaShortCircuit),
		fmt.Sprintf("\t\t[bit5]GNSS天线未接或被剪断:[%t]", a.GNSSAntennaFault),
		fmt.Sprintf("\t\t[bit4]GNSS模块发生故障:[%t]", a.GNSSModuleFault),
		fmt.Sprintf("\t\t[bit3]危险预警:[%t]", a.DangerousAlarm),
		fmt.Sprintf("\t\t[bit2]疲劳驾驶:[%t]", a.FatigueDriving),
		fmt.Sprintf("\t\t[bit1]超速报警:[%t]", a.OverSpeed),
		fmt.Sprintf("\t\t[bit0]紧急报警,触动报警开关后触发:[%t]", a.EmergencyAlarm),
	}, "\n")
}

func (s *StatusSignDetails) parse(statusSign uint32) {
	data := fmt.Sprintf("%.32b", statusSign)
	if data[31] == '1' {
		s.ACC = true
	}
	if data[30] == '1' {
		s.Location = true
	}
	if data[29] == '1' {
		s.South = true
	}
	if data[28] == '1' {
		s.East = true
	}
	if data[27] == '1' {
		s.Suspended = true
	}
	if data[26] == '1' {
		s.Encryption = true
	}
	if data[25] == '1' {
		s.EmergencyBrake = true
	}
	if data[24] == '1' {
		s.LaneOffset = true
	}

	if data[23] == '0' && data[22] == '0' {
		s.Cargo = uint8(0)
	} else if data[23] == '0' && data[22] == '1' {
		s.Cargo = uint8(1)
	} else if data[23] == '1' && data[22] == '0' {
		s.Cargo = uint8(2)
	} else {
		s.Cargo = uint8(3)
	}

	if data[21] == '1' {
		s.Oil = true
	}
	if data[20] == '1' {
		s.Electricity = true
	}
	if data[19] == '1' {
		s.VehicleDoor = true
	}
	if data[18] == '1' {
		s.FrontDoor = true
	}
	if data[17] == '1' {
		s.MiddleDoor = true
	}
	if data[16] == '1' {
		s.BackDoor = true
	}
	if data[15] == '1' {
		s.DriverDoor = true
	}
	if data[14] == '1' {
		s.CustomDoor = true
	}
	if data[13] == '1' {
		s.UseGPS = true
	}
	if data[12] == '1' {
		s.UseBD = true
	}
	if data[11] == '1' {
		s.UseGLONASS = true
	}
	if data[10] == '1' {
		s.UseGalileo = true
	}
	if data[9] == '1' {
		s.VehicleRunning = true
	}
}

func (s *StatusSignDetails) String() string {
	return strings.Join([]string{
		fmt.Sprintf("\t\t[bit22]车辆状态 是否运行:[%t] 2019版本增加的", s.VehicleRunning),
		fmt.Sprintf("\t\t[bit21]使用Galileo卫星进行定位:[%t]", s.UseGalileo),
		fmt.Sprintf("\t\t[bit20]未使用GLONASS卫星进行定位:[%t]", s.UseGLONASS),
		fmt.Sprintf("\t\t[bit19]未使用北斗卫星进行定位:[%t]", s.UseBD),
		fmt.Sprintf("\t\t[bit18]未使用GPS卫星进行定位:[%t]", s.UseGPS),
		fmt.Sprintf("\t\t[bit17]自定义门 门5:[%t]", s.CustomDoor),
		fmt.Sprintf("\t\t[bit16]驾驶席门 门4:[%t]", s.DriverDoor),
		fmt.Sprintf("\t\t[bit15]后门情况 门3:[%t]", s.BackDoor),
		fmt.Sprintf("\t\t[bit14]中门情况 门2:[%t]", s.MiddleDoor),
		fmt.Sprintf("\t\t[bit13]前门情况 门1:[%t]", s.FrontDoor),
		fmt.Sprintf("\t\t[bit12]车门情况 是否解锁:[%t]", s.VehicleDoor),
		fmt.Sprintf("\t\t[bit11]电路情况 是否断开:[%t]", s.Electricity),
		fmt.Sprintf("\t\t[bit10]油路情况 是否正常:[%t]", s.Oil),
		fmt.Sprintf("\t\t[bit8-bit9]载客情况 0-空车 1-半载 2-保留 3-满载:[%02b]", s.Cargo),
		fmt.Sprintf("\t\t[bit7]车道偏移预警:[%t] 2019版本增加的", s.LaneOffset),
		fmt.Sprintf("\t\t[bit6]紧急刹车系统的前撞预警:[%t] 2019版本增加的", s.EmergencyBrake),
		fmt.Sprintf("\t\t[bit5]经纬度是否加密:[%t]", s.Encryption),
		fmt.Sprintf("\t\t[bit4]运营状态 是否停运:[%t]", s.Suspended),
		fmt.Sprintf("\t\t[bit3]东西经 0-东经 1-西经:[%t]", s.East),
		fmt.Sprintf("\t\t[bit2]南北纬 0-北纬 1-南纬:[%t]", s.South),
		fmt.Sprintf("\t\t[bit1]定位状态 是否开:[%t]", s.Location),
		fmt.Sprintf("\t\t[bit0]ACC 是否开:[%t]", s.ACC),
	}, "\n")
}
