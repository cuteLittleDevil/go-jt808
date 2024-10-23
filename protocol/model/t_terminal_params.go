package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"sort"
	"strings"
)

type (
	TerminalParamDetails struct {
		// T0x001HeartbeatInterval 终端心跳发送间隔,单位为秒(s)
		T0x001HeartbeatInterval ParamContent[uint32] `json:"t0X001HeartbeatInterval"`
		// T0x002TCPRespondOverTime TCP消息应答超时时间,单位为秒(s)
		T0x002TCPRespondOverTime ParamContent[uint32] `json:"t0X002TCPRespondOverTime"`
		// T0x003TCPRetransmissionCount TCP消息重传次数
		T0x003TCPRetransmissionCount ParamContent[uint32] `json:"t0X003TCPRetransmissionCount"`
		// T0x004UDPRespondOverTime UDP消息应答超时时间,单位为秒(s)
		T0x004UDPRespondOverTime ParamContent[uint32] `json:"t0X004UDPRespondOverTime"`
		// T0x005UDPRetransmissionCount UDP消息重传次数
		T0x005UDPRetransmissionCount ParamContent[uint32] `json:"t0X005UDPRetransmissionCount"`
		// T0x006SMSRetransmissionCount SMS消息应答超时时间,单位为秒(s)
		T0x006SMSRetransmissionCount ParamContent[uint32] `json:"t0X006SMSRetransmissionCount"`
		// T0x007SMSRetransmissionCount SMS消息重传次数
		T0x007SMSRetransmissionCount ParamContent[uint32] `json:"t0X007SMSRetransmissionCount"`
		// T0x010APN 主服务器APN,无线通信拨号访问点.若网络制式为CDMA,则该处为PPP拨号号码
		T0x010APN ParamContent[string] `json:"t0X010APN"`
		// T0x011WIFIUsername 主服务器无线通信拨号用户名
		T0x011WIFIUsername ParamContent[string] `json:"t0X011WIFIUsername"`
		// T0x012WIFIPassword 主服务器无线通信拨号密码
		T0x012WIFIPassword ParamContent[string] `json:"t0X012WIFIPassword"`
		// T0x013Address 主服务器地址,IP或域名,以冒号分割主机和端口,多个服务器使用分号分割
		T0x013Address ParamContent[string] `json:"t0X013Address"`
		// T0x014BackupServerAPN 备份服务器APN
		T0x014BackupServerAPN ParamContent[string] `json:"t0X014BackupServerAPN"`
		// T0x015BackupServerWIFIUsername 备份服务器无线通信拨号用户名
		T0x015BackupServerWIFIUsername ParamContent[string] `json:"t0X015BackupServerWIFIUsername"`
		// T0x016BackupServerWIFIPassword 备份服务器无线通信拨号密码
		T0x016BackupServerWIFIPassword ParamContent[string] `json:"t0X016BackupServerWIFIPassword"`
		// T0x017BackupServerAddress 备份服务器地址,IP或域名,以冒号分割主机和端口,多个服务器使用分号分割
		T0x017BackupServerAddress ParamContent[string] `json:"t0X017BackupServerAddress"`
		// T0x018TCPPort 服务器TCP端口（2013版本)
		T0x018TCPPort ParamContent[uint32] `json:"t0X018TCPPort"`
		// T0x019UDPPort 服务器UDP端口 (2013版本)
		T0x019UDPPort ParamContent[uint32] `json:"t0X019UDPPort"`
		// T0x01AICCardAddress 道路运输证IC卡认证主服务器IP地址或域名
		T0x01AICCardAddress ParamContent[string] `json:"t0X01AICCardAddress"`
		// T0x01BICCardTCPPort 道路运输证IC卡认证主服务器TCP端口
		T0x01BICCardTCPPort ParamContent[uint32] `json:"t0X01BICCardTCPPort"`
		// T0x01CICCardUDPPort 道路运输证IC卡认证主服务器UDP端口
		T0x01CICCardUDPPort ParamContent[uint32] `json:"t0X01CICCardUDPPort"`
		// T0x01DICCardAddress 道路运输证IC卡认证主服务器IP地址或域名,端口同主服务器
		T0x01DICCardAddress ParamContent[string] `json:"t0X01DICCardAddress"`
		// T0x020PositionReportingStrategy 位置汇报策略:0.定时汇报 1.定距汇报 2.定时和定距汇报
		T0x020PositionReportingStrategy ParamContent[uint32] `json:"t0X020PositionReportingStrategy"`
		// T0x021PositionReportingPlan 位置汇报方案:0.根据ACC状态 1.根据登录状态和ACC状态,先判断登录状态,若登录再根据ACC状态
		T0x021PositionReportingPlan ParamContent[uint32] `json:"t0X021PositionReportingPlan"`
		// T0x022DriverReportingInterval 驾驶员未登录汇报时间间隔,单位为秒(s),值大于0
		T0x022DriverReportingInterval ParamContent[uint32] `json:"t0X022DriverReportingInterval"`
		// T0x023FromServerAPN 从服务器APN.该值为空时,终端应使用主服务器相同配置 (2019版本)
		T0x023FromServerAPN ParamContent[string] `json:"t0X023FromServerAPN"`
		// T0x024FromServerAPNWIFIUsername 从服务器无线通信拨号用户名。该值为空时,终端应使用主服务器相同配置 (2019版本)
		T0x024FromServerAPNWIFIUsername ParamContent[string] `json:"t0X024FromServerAPNWIFIUsername"`
		// T0x025FromServerAPNWIFIPassword 从服务器无线通信拨号密码.该值为空时,终端应使用主服务器相同配置 (2019版本)
		T0x025FromServerAPNWIFIPassword ParamContent[string] `json:"t0X025FromServerAPNWIFIPassword"`
		// T0x026FromServerAPNWIFIAddress 从服务器备份地址、IP或域名,主机和端口用冒号分割,多个服务器使用分号分割 (2019版本)
		T0x026FromServerAPNWIFIAddress ParamContent[string] `json:"t0X026FromServerAPNWIFIAddress"`
		// T0x027ReportingTimeInterval 休眠时汇报时间间隔,单位为秒(s),值大于0
		T0x027ReportingTimeInterval ParamContent[uint32] `json:"t0X027ReportingTimeInterval"`
		// T0x028EmergencyReportingTimeInterval 紧急报警时汇报时间间隔,单位为秒(s),值大于0
		T0x028EmergencyReportingTimeInterval ParamContent[uint32] `json:"t0X028EmergencyReportingTimeInterval"`
		// T0x029DefaultReportingTimeInterval 缺省时间汇报间隔,单位为秒(s),值大于0
		T0x029DefaultReportingTimeInterval ParamContent[uint32] `json:"t0X029DefaultReportingTimeInterval"`
		// T0x02CDefaultDistanceReportingTimeInterval 缺省距离汇报间隔,单位为米(m),值大于0
		T0x02CDefaultDistanceReportingTimeInterval ParamContent[uint32] `json:"t0X02CDefaultDistanceReportingTimeInterval"`
		// T0x02DDrivingReportingDistanceInterval 驾驶员未登录汇报距离间隔,单位为米(m),值大于0
		T0x02DDrivingReportingDistanceInterval ParamContent[uint32] `json:"t0X02DDrivingReportingDistanceInterval"`
		// T0x02ESleepReportingDistanceInterval 休眠时汇报距离间隔,单位为米(m),值大于0
		T0x02ESleepReportingDistanceInterval ParamContent[uint32] `json:"t0X02ESleepReportingDistanceInterval"`
		// T0x02FAlarmReportingDistanceInterval 紧急报警时汇报距离间隔,单位为米(m),值大于0
		T0x02FAlarmReportingDistanceInterval ParamContent[uint32] `json:"t0X02FAlarmReportingDistanceInterval"`
		// T0x030InflectionPointSupplementaryPassAngle 拐点补传角度,值小于180度
		T0x030InflectionPointSupplementaryPassAngle ParamContent[uint32] `json:"t0X030InflectionPointSupplementaryPassAngle"`
		// T0x031GeofenceRadius 电子围栏半径(非法位移阈值),单位为米(m)
		T0x031GeofenceRadius ParamContent[uint16] `json:"t0X031GeofenceRadius"`
		// T0x032IllegalDrivingTime 违规行驶时段范围,精确到分。(2019版本)
		//   byte1:违规行驶开始时间的小时部分；
		//   byte2:违规行驶开始的分钟部分；
		//   byte3:违规行驶结束时间的小时部分；
		//   byte4:违规行驶结束时间的分钟部分。
		// 示例: 0x16320ALE 表示当天晚上10点50分到第二天早上10点30分属于违规行驶时段
		T0x032IllegalDrivingTime ParamContent[[4]byte] `json:"t0X032IllegalDrivingTime"`
		// T0x040MonitoringPlatformPhone 监控平台电话号码
		T0x040MonitoringPlatformPhone ParamContent[string] `json:"t0X040MonitoringPlatformPhone"`
		// T0x041ResetPhone 复位电话号码,可采用此电话号码拨打终端电话让终端复位
		T0x041ResetPhone ParamContent[string] `json:"t0X041ResetPhone"`
		// T0x042RestoreFactoryPhone 恢复出厂设置电话号码,可采用此电话号码拨打终端电话让终端恢复出厂设置
		T0x042RestoreFactoryPhone ParamContent[string] `json:"t0X042RestoreFactoryPhone"`
		// T0x043SMSPhone 监控平台SMS电话号码
		T0x043SMSPhone ParamContent[string] `json:"t0X043SMSPhone"`
		// T0x044SMSTxtPhone 接收终端SMS文本报警号码
		T0x044SMSTxtPhone ParamContent[string] `json:"t0X044SMSTxtPhone"`
		// T0x045TerminalTelephoneStrategy 终端电话接听策略,0-自动接听 1-ACC ON时自动接听,OFF时手动接听
		T0x045TerminalTelephoneStrategy ParamContent[uint32] `json:"t0X045TerminalTelephoneStrategy"`
		// T0x046MaximumCallTime 每次最长通话时间,单位为秒(s),0为不允许通话,0xFFFFFFFF为不限制
		T0x046MaximumCallTime ParamContent[uint32] `json:"t0X046MaximumCallTime"`
		// T0x047MonthMaximumCallTime 当月最长通话时间,单位为秒(s),0为不允许通话,0xFFFFFFFF为不限制
		T0x047MonthMaximumCallTime ParamContent[uint32] `json:"t0X047MonthMaximumCallTime"`
		// T0x048MonitorPhone 监听电话号码
		T0x048MonitorPhone ParamContent[string] `json:"t0X048MonitorPhone"`
		// T0x049MonitorPrivilegedSMS 监管平台特权短信号码
		T0x049MonitorPrivilegedSMS ParamContent[string] `json:"t0X049MonitorPrivilegedSMS"`
		// T0x050AlarmBlockingWords 报警屏蔽字.与位置信息汇报消息中的报警标志相对应,相应位为1则相应报警被屏蔽
		T0x050AlarmBlockingWords ParamContent[uint32] `json:"t0X050AlarmBlockingWords"`
		// T0x051AlarmSendTextSMSSwitch 报警发送文本SMS开关,与位置信息汇报消息中的报警标志相对应,相应位为1则相应报警时发送文本SMS
		T0x051AlarmSendTextSMSSwitch ParamContent[uint32] `json:"t0X051AlarmSendTextSMSSwitch"`
		// T0x052AlarmShootingSwitch 报警拍摄开关,与位置信息汇报消息中的报警标志相对应,相应位为1则相应报警时摄像头拍摄
		T0x052AlarmShootingSwitch ParamContent[uint32] `json:"t0X052AlarmShootingSwitch"`
		// T0x053AlarmShootingStorageSign 报警拍摄存储标志,与位置信息汇报消息中的报警标志相对应,相应位为1则对相应报警时牌的照片进行存储,否则实时长传
		T0x053AlarmShootingStorageSign ParamContent[uint32] `json:"t0X053AlarmShootingStorageSign"`
		// T0x054KeySign 关键标志,与位置信息汇报消息中的报警标志相对应,相应位为1则对相应报警为关键报警
		T0x054KeySign ParamContent[uint32] `json:"t0X054KeySign"`
		// T0x055MaxSpeed 最高速度,单位为千米每小时(km/h)
		T0x055MaxSpeed ParamContent[uint32] `json:"t0X055MaxSpeed"`
		// T0x056DurationOverSpeed 超速持续时间,单位为秒(s)
		T0x056DurationOverSpeed ParamContent[uint32] `json:"t0X056DurationOverSpeed"`
		// T0x057ContinuousDrivingTimeLimit 连续驾驶时间门限,单位为秒(s)
		T0x057ContinuousDrivingTimeLimit ParamContent[uint32] `json:"t0X057ContinuousDrivingTimeLimit"`
		// T0x058CumulativeDayDrivingTime 当天累计驾驶时间门限,单位为秒(s)
		T0x058CumulativeDayDrivingTime ParamContent[uint32] `json:"t0X058CumulativeDayDrivingTime"`
		// T0x059MinimumRestTime 最小休息时间,单位为秒(s)
		T0x059MinimumRestTime ParamContent[uint32] `json:"t0X059MinimumRestTime"`
		// T0x05AMaximumParkingTime 最长停车时间,单位为秒(s)
		T0x05AMaximumParkingTime ParamContent[uint32] `json:"t0X05AMaximumParkingTime"`
		// T0x05BSpeedWarningDifference 超速预警差值,单位1/10千米每小时(1/10 km/h)
		T0x05BSpeedWarningDifference ParamContent[uint16] `json:"t0X05BSpeedWarningDifference"`
		// T0x05CFatigueDrivingWarningInterpolation 疲劳驾驶预警插值,单位为秒(s),值大于0
		T0x05CFatigueDrivingWarningInterpolation ParamContent[uint16] `json:"t0X05CFatigueDrivingWarningInterpolation"`
		// T0x05DCollisionAlarmParam 碰撞报警参数设置 b7-b0: 为碰撞时间,单位为毫秒(ms) b15-18 为碰撞加速度,单位为0.1g;设置范围0-79,默认10
		T0x05DCollisionAlarmParam ParamContent[uint16] `json:"t0X05DCollisionAlarmParam"`
		// T0x05ERolloverAlarmParam 侧翻报警参数设置:侧翻角度,单位为度,默认为30度
		T0x05ERolloverAlarmParam ParamContent[uint16] `json:"t0X05ERolloverAlarmParam"`
		// T0x064TimedPhotographyParam 定时拍照参数,参数项格式和定义见表14
		T0x064TimedPhotographyParam ParamContent[uint32] `json:"t0X064TimedPhotographyParam"`
		// T0x065FixedDistanceShootingParam 定距拍照参数,参数项格式和定义见表15
		T0x065FixedDistanceShootingParam ParamContent[uint32] `json:"t0X065FixedDistanceShootingParam"`
		// T0x070ImageVideoQuality 图像/视频质量,设置范围为1-10,1表示最优质量
		T0x070ImageVideoQuality ParamContent[uint32] `json:"t0X070ImageVideoQuality"`
		// T0x071Brightness 亮度,设置范围为0-255
		T0x071Brightness ParamContent[uint32] `json:"t0X071Brightness"`
		// T0x072Contrast 对比度,设置范围为0-127
		T0x072Contrast ParamContent[uint32] `json:"t0X072Contrast"`
		// T0x073Saturation 饱和度,设置范围为0-127
		T0x073Saturation ParamContent[uint32] `json:"t0X073Saturation"`
		// T0x074Chrominance 色度,设置范围为0-255
		T0x074Chrominance ParamContent[uint32] `json:"t0X074Chrominance"`
		// T0x080VehicleOdometerReadings 车辆里程表读数,单位:1/10km
		T0x080VehicleOdometerReadings ParamContent[uint32] `json:"t0X080VehicleOdometerReadings"`
		// T0x081VehicleProvinceID 车辆所在的省域ID
		T0x081VehicleProvinceID ParamContent[uint16] `json:"t0X081VehicleProvinceID"`
		// T0x082VehicleCityID 车辆所在的市域ID
		T0x082VehicleCityID ParamContent[uint16] `json:"t0X082VehicleCityID"`
		// T0x083MotorVehicleLicensePlate 公安交通管理部门颁发的机动车号牌
		T0x083MotorVehicleLicensePlate ParamContent[string] `json:"t0X083MotorVehicleLicensePlate"`
		// T0x084licensePlateColor 车牌颜色,值按照JT/T 797.7-2014中的规定,未上牌车辆填0
		T0x084licensePlateColor ParamContent[byte] `json:"t0X084LicensePlateColor"`
		// T0x090GNSSPositionMode GNSS定位模式,定义如下:
		//   bit0: 0-禁用GPS定位,1-启用 GPS 定位;
		//   bit1: 0-禁用北斗定位,1-启用北斗定位;
		//   bit2: 0-禁用GLONASS定位,1-启用GLONASS定位;
		//   bit3: 0-禁用Galileo定位,1-启用Galileo定位
		T0x090GNSSPositionMode ParamContent[byte] `json:"t0X090GNSSPositionMode"`
		// T0x091GNSSBaudRate GNSS波特率,定义如下:
		//   0x00:4800;
		//   0x01:9600;
		//   0x02:19200;
		//   0x03:38400;
		//   0x04:57600;
		//   0x05:115200
		T0x091GNSSBaudRate ParamContent[byte] `json:"t0X091GNSSBaudRate"`
		// T0x092GNSSModePositionOutputFrequency GNSS模块详细定位数据输出频率,定义如下:
		//   0x00:500ms;
		//   0x01:1000ms(默认值);
		//   0x02:2000ms;
		//   0x03:3000ms;
		//   0x04:4000ms
		T0x092GNSSModePositionOutputFrequency ParamContent[byte] `json:"t0X092GNSSModePositionOutputFrequency"`
		// T0x093GNSSModePositionAcquisitionFrequency GNSS模块详细定位数据采集频率,单位为秒(s),默认为 1。
		T0x093GNSSModePositionAcquisitionFrequency ParamContent[uint32] `json:"t0X093GNSSModePositionAcquisitionFrequency"`
		// T0x094GNSSModePositionUploadMethod GNSS模块详细定位数据上传方式:
		//   0x00,本地存储,不上传(默认值);
		//   0x01,按时间间隔上传;
		//   0x02,按距离间隔上传;
		//   0x0B,按累计时间上传,达到传输时间后自动停止上传;
		//   0x0C,按累计距离上传,达到距离后自动停止上传;
		//   0x0D,按累计条数上传,达到上传条数后自动停止上传。
		T0x094GNSSModePositionUploadMethod ParamContent[byte] `json:"t0X094GNSSModePositionUploadMethod"`
		// T0x095GNSSModeSetPositionUpload GNSS模块详细定位数据上传设置, 关联0x0094:
		// 上传方式为 0x01 时,单位为秒;
		// 上传方式为 0x02 时,单位为米;
		// 上传方式为 0x0B 时,单位为秒;
		// 上传方式为 0x0C 时,单位为米;
		// 上传方式为 0x0D 时,单位为条。
		T0x095GNSSModeSetPositionUpload ParamContent[uint32] `json:"t0X095GNSSModeSetPositionUpload"`
		// T0x100CANCollectionTimeInterval CAN总线通道1采集时间间隔(ms),0表示不采集
		T0x100CANCollectionTimeInterval ParamContent[uint32] `json:"t0X100CANCollectionTimeInterval"`
		// T0x101CAN1UploadTimeInterval CAN总线通道1上传时间间隔(s),0表示不上传
		T0x101CAN1UploadTimeInterval ParamContent[uint16] `json:"t0X101CAN1UploadTimeInterval"`
		// T0x102CAN2CollectionTimeInterval CAN总线通道2采集时间间隔(ms),0表示不采集
		T0x102CAN2CollectionTimeInterval ParamContent[uint32] `json:"t0X102CAN2CollectionTimeInterval"`
		// T0x103CAN2UploadTimeInterval CAN总线通道2上传时间间隔(s),0表示不上传
		T0x103CAN2UploadTimeInterval ParamContent[uint16] `json:"t0X103CAN2UploadTimeInterval"`
		// T0x110CANIDSetIndividualAcquisition CAN总线ID单独采集设置:
		//   bit63-bit32: 表示此 ID 采集时间间隔(ms),0 表示不采集;
		//   bit31: 表示 CAN 通道号,0:CAN1,1:CAN2;
		//   bit30: 表示帧类型,0:标准帧,1:扩展帧;
		//   bit29: 表示数据采集方式,0:原始数据,1:采集区间的计算值;
		//   bit28-bit0: 表示 CAN 总线 ID。
		T0x110CANIDSetIndividualAcquisition ParamContent[[8]byte] `json:"t0X0110CANIDSetIndividualAcquisition"`

		// ParamParseBeforeFunc 参数解析前 用于自定义消息处理
		ParamParseBeforeFunc func(id uint32, content []byte) `json:"-"`
		// OtherContent 未知的解析内容
		OtherContent map[uint32]ParamContent[[]byte] `json:"otherContent"`
		//// AuxiliaryFields 辅助字段列表 用于2013和2019版本的部分不同处
		//AuxiliaryFields
	}
	ParamContent[T byte | uint16 | uint32 | string | []byte | [4]byte | [8]byte] struct {
		// ID 参数ID
		ID uint32 `json:"id"`
		// Len 参数长度
		Len byte `json:"len"`
		// Value 参数内容
		Value T `json:"value"`
	}
)

func (t *TerminalParamDetails) parse(count uint8, body []byte) error {
	index := 0
	if len(t.OtherContent) == 0 {
		t.OtherContent = make(map[uint32]ParamContent[[]byte])
	}
	for index < len(body) {
		if index+5 > len(body) {
			return protocol.ErrBodyLengthInconsistency
		}
		id := binary.BigEndian.Uint32(body[index : index+4])
		paramLen := body[index+4]
		start := index + 5
		end := start + int(paramLen)
		if end > len(body) {
			return protocol.ErrBodyLengthInconsistency
		}
		content := body[start:end]
		if err := t.parseParam(id, paramLen, content); err != nil {
			return err
		}
		index = end
		count--
	}
	if count != 0 {
		return protocol.ErrBodyLengthInconsistency
	}
	return nil
}

func (t *TerminalParamDetails) parseParam(id uint32, paramLen byte, content []byte) error {
	if t.ParamParseBeforeFunc != nil {
		t.ParamParseBeforeFunc(id, content)
	}
	switch id {
	case 0x001, 0x002, 0x003, 0x004, 0x005, 0x006, 0x007, 0x01b, 0x01c, 0x020,
		0x022, 0x027, 0x028, 0x029, 0x02a, 0x02b, 0x02c, 0x02d, 0x02e, 0x02f,
		0x030, 0x045, 0x046, 0x047, 0x050, 0x051, 0x052, 0x053, 0x054, 0x055,
		0x056, 0x057, 0x058, 0x059, 0x05a, 0x064, 0x065, 0x070, 0x071, 0x072,
		0x073, 0x074, 0x080, 0x093, 0x095, 0x100, 0x102:
		if paramLen != 4 {
			return protocol.ErrBodyLengthInconsistency
		}
		tmp := ParamContent[uint32]{
			ID:    id,
			Len:   paramLen,
			Value: binary.BigEndian.Uint32(content),
		}
		t.parseParamDWORD(id, tmp)
	case 0x031, 0x05b, 0x05c, 0x05d, 0x05e, 0x081, 0x082, 0x101, 0x103:
		if paramLen != 2 {
			return protocol.ErrBodyLengthInconsistency
		}
		tmp := ParamContent[uint16]{
			ID:    id,
			Len:   paramLen,
			Value: binary.BigEndian.Uint16(content),
		}
		t.parseParamWORD(id, tmp)
	case 0x010, 0x011, 0x012, 0x013, 0x014, 0x015, 0x016, 0x017, 0x01a, 0x01d,
		0x023, 0x024, 0x025, 0x026, 0x040, 0x041, 0x042, 0x043, 0x044, 0x048,
		0x049, 0x083:
		tmp := ParamContent[string]{
			ID:    id,
			Len:   paramLen,
			Value: string(utils.GBK2UTF8(content)),
		}
		t.parseParamString(id, tmp)
	case 0x032:
		if paramLen != 4 {
			return protocol.ErrBodyLengthInconsistency
		}
		t.T0x032IllegalDrivingTime = ParamContent[[4]byte]{
			ID:    id,
			Len:   paramLen,
			Value: [4]byte(content),
		}
	case 0x084, 0x090, 0x091, 0x092, 0x094:
		if paramLen != 1 {
			return protocol.ErrBodyLengthInconsistency
		}
		tmp := ParamContent[byte]{
			ID:    id,
			Len:   paramLen,
			Value: content[0],
		}
		t.parseParamByte(id, tmp)
	case 0x110:
		if paramLen != 8 {
			return protocol.ErrBodyLengthInconsistency
		}
		t.T0x110CANIDSetIndividualAcquisition = ParamContent[[8]byte]{
			ID:    id,
			Len:   paramLen,
			Value: [8]byte(content),
		}
	default:
		t.OtherContent[id] = ParamContent[[]byte]{
			ID:    id,
			Len:   paramLen,
			Value: content,
		}
	}
	return nil
}

func (t *TerminalParamDetails) parseParamDWORD(id uint32, dwordContent ParamContent[uint32]) {
	switch id {
	case 0x001:
		t.T0x001HeartbeatInterval = dwordContent
	case 0x002:
		t.T0x002TCPRespondOverTime = dwordContent
	case 0x003:
		t.T0x003TCPRetransmissionCount = dwordContent
	case 0x004:
		t.T0x004UDPRespondOverTime = dwordContent
	case 0x005:
		t.T0x005UDPRetransmissionCount = dwordContent
	case 0x006:
		t.T0x006SMSRetransmissionCount = dwordContent
	case 0x007:
		t.T0x007SMSRetransmissionCount = dwordContent
	case 0x01b:
		t.T0x01BICCardTCPPort = dwordContent
	case 0x01c:
		t.T0x01CICCardUDPPort = dwordContent
	case 0x020:
		t.T0x020PositionReportingStrategy = dwordContent
	case 0x022:
		t.T0x022DriverReportingInterval = dwordContent
	case 0x027:
		t.T0x027ReportingTimeInterval = dwordContent
	case 0x028:
		t.T0x028EmergencyReportingTimeInterval = dwordContent
	case 0x029:
		t.T0x029DefaultReportingTimeInterval = dwordContent
	case 0x02c:
		t.T0x02CDefaultDistanceReportingTimeInterval = dwordContent
	case 0x02d:
		t.T0x02DDrivingReportingDistanceInterval = dwordContent
	case 0x02e:
		t.T0x02ESleepReportingDistanceInterval = dwordContent
	case 0x02f:
		t.T0x02FAlarmReportingDistanceInterval = dwordContent
	case 0x030:
		t.T0x030InflectionPointSupplementaryPassAngle = dwordContent
	case 0x045:
		t.T0x045TerminalTelephoneStrategy = dwordContent
	case 0x046:
		t.T0x046MaximumCallTime = dwordContent
	case 0x047:
		t.T0x047MonthMaximumCallTime = dwordContent
	case 0x050:
		t.T0x050AlarmBlockingWords = dwordContent
	case 0x051:
		t.T0x051AlarmSendTextSMSSwitch = dwordContent
	case 0x052:
		t.T0x052AlarmShootingSwitch = dwordContent
	case 0x053:
		t.T0x053AlarmShootingStorageSign = dwordContent
	case 0x054:
		t.T0x054KeySign = dwordContent
	case 0x055:
		t.T0x055MaxSpeed = dwordContent
	case 0x056:
		t.T0x056DurationOverSpeed = dwordContent
	case 0x057:
		t.T0x057ContinuousDrivingTimeLimit = dwordContent
	case 0x058:
		t.T0x058CumulativeDayDrivingTime = dwordContent
	case 0x059:
		t.T0x059MinimumRestTime = dwordContent
	case 0x05a:
		t.T0x05AMaximumParkingTime = dwordContent
	case 0x064:
		t.T0x064TimedPhotographyParam = dwordContent
	case 0x065:
		t.T0x065FixedDistanceShootingParam = dwordContent
	case 0x070:
		t.T0x070ImageVideoQuality = dwordContent
	case 0x071:
		t.T0x071Brightness = dwordContent
	case 0x072:
		t.T0x072Contrast = dwordContent
	case 0x073:
		t.T0x073Saturation = dwordContent
	case 0x074:
		t.T0x074Chrominance = dwordContent
	case 0x080:
		t.T0x080VehicleOdometerReadings = dwordContent
	case 0x093:
		t.T0x093GNSSModePositionAcquisitionFrequency = dwordContent
	case 0x095:
		t.T0x095GNSSModeSetPositionUpload = dwordContent
	case 0x100:
		t.T0x100CANCollectionTimeInterval = dwordContent
	case 0x102:
		t.T0x102CAN2CollectionTimeInterval = dwordContent
	}
}

func (t *TerminalParamDetails) parseParamWORD(id uint32, wordContent ParamContent[uint16]) {
	switch id {
	case 0x031:
		t.T0x031GeofenceRadius = wordContent
	case 0x05b:
		t.T0x05BSpeedWarningDifference = wordContent
	case 0x05c:
		t.T0x05CFatigueDrivingWarningInterpolation = wordContent
	case 0x05d:
		t.T0x05DCollisionAlarmParam = wordContent
	case 0x05e:
		t.T0x05ERolloverAlarmParam = wordContent
	case 0x081:
		t.T0x081VehicleProvinceID = wordContent
	case 0x082:
		t.T0x082VehicleCityID = wordContent
	case 0x101:
		t.T0x101CAN1UploadTimeInterval = wordContent
	case 0x103:
		t.T0x103CAN2UploadTimeInterval = wordContent
	}
}

func (t *TerminalParamDetails) parseParamByte(id uint32, byteContent ParamContent[byte]) {
	switch id {
	case 0x084:
		t.T0x084licensePlateColor = byteContent
	case 0x090:
		t.T0x090GNSSPositionMode = byteContent
	case 0x091:
		t.T0x091GNSSBaudRate = byteContent
	case 0x092:
		t.T0x092GNSSModePositionOutputFrequency = byteContent
	case 0x094:
		t.T0x094GNSSModePositionUploadMethod = byteContent
	}
}

func (t *TerminalParamDetails) parseParamString(id uint32, stringContent ParamContent[string]) {
	switch id {
	case 0x010:
		t.T0x010APN = stringContent
	case 0x011:
		t.T0x011WIFIUsername = stringContent
	case 0x012:
		t.T0x012WIFIPassword = stringContent
	case 0x013:
		t.T0x013Address = stringContent
	case 0x014:
		t.T0x014BackupServerAPN = stringContent
	case 0x015:
		t.T0x015BackupServerWIFIUsername = stringContent
	case 0x016:
		t.T0x016BackupServerWIFIPassword = stringContent
	case 0x017:
		t.T0x017BackupServerAddress = stringContent
	case 0x01a:
		t.T0x01AICCardAddress = stringContent
	case 0x01d:
		t.T0x01DICCardAddress = stringContent
	case 0x023:
		t.T0x023FromServerAPN = stringContent
	case 0x024:
		t.T0x024FromServerAPNWIFIUsername = stringContent
	case 0x025:
		t.T0x025FromServerAPNWIFIPassword = stringContent
	case 0x026:
		t.T0x026FromServerAPNWIFIAddress = stringContent
	case 0x040:
		t.T0x040MonitoringPlatformPhone = stringContent
	case 0x041:
		t.T0x041ResetPhone = stringContent
	case 0x042:
		t.T0x042RestoreFactoryPhone = stringContent
	case 0x043:
		t.T0x043SMSPhone = stringContent
	case 0x044:
		t.T0x044SMSTxtPhone = stringContent
	case 0x048:
		t.T0x048MonitorPhone = stringContent
	case 0x049:
		t.T0x049MonitorPrivilegedSMS = stringContent
	case 0x083:
		t.T0x083MotorVehicleLicensePlate = stringContent
	}
}

func (t *TerminalParamDetails) String() string {
	str := strings.Join([]string{
		"\t{",
		fmt.Sprintf("\t\t[0001]终端参数ID:1 终端心跳发送间隔,单位为秒(s)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x001HeartbeatInterval.Len, t.T0x001HeartbeatInterval.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x001HeartbeatInterval.Value, t.T0x001HeartbeatInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0002]终端参数ID:2 TCP消息应答超时时间,单位为秒(s)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x002TCPRespondOverTime.Len, t.T0x002TCPRespondOverTime.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x002TCPRespondOverTime.Value, t.T0x002TCPRespondOverTime.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0003]终端参数ID:3 TCP消息重传次数"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x003TCPRetransmissionCount.Len, t.T0x003TCPRetransmissionCount.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x003TCPRetransmissionCount.Value, t.T0x003TCPRetransmissionCount.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0004]终端参数ID:4 UDP消息应答超时时间,单位为秒(s)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x004UDPRespondOverTime.Len, t.T0x004UDPRespondOverTime.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x004UDPRespondOverTime.Value, t.T0x004UDPRespondOverTime.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0005]终端参数ID:5 UDP消息重传次数"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x005UDPRetransmissionCount.Len, t.T0x005UDPRetransmissionCount.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x005UDPRetransmissionCount.Value, t.T0x005UDPRetransmissionCount.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0006]终端参数ID:6 SMS消息应答超时时间,单位为秒(s)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x006SMSRetransmissionCount.Len, t.T0x006SMSRetransmissionCount.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x006SMSRetransmissionCount.Value, t.T0x006SMSRetransmissionCount.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0007]终端参数ID:7 SMS消息重传次数"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x007SMSRetransmissionCount.Len, t.T0x007SMSRetransmissionCount.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x007SMSRetransmissionCount.Value, t.T0x007SMSRetransmissionCount.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0010]终端参数ID:16 主服务器APN,无线通信拨号访问点.若网络制式为CDMA,则该处为PPP拨号号码"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x010APN.Len, t.T0x010APN.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x010APN.Value, t.T0x010APN.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0011]终端参数ID:17 主服务器无线通信拨号用户名"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x011WIFIUsername.Len, t.T0x011WIFIUsername.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x011WIFIUsername.Value, t.T0x011WIFIUsername.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0012]终端参数ID:18 主服务器无线通信拨号密码"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x012WIFIPassword.Len, t.T0x012WIFIPassword.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x012WIFIPassword.Value, t.T0x012WIFIPassword.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0013]终端参数ID:19 主服务器地址,IP或域名,以冒号分割主机和端口,多个服务器使用分号分割"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x013Address.Len, t.T0x013Address.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x013Address.Value, t.T0x013Address.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0014]终端参数ID:20 备份服务器APN"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x014BackupServerAPN.Len, t.T0x014BackupServerAPN.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x014BackupServerAPN.Value, t.T0x014BackupServerAPN.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0015]终端参数ID:21 备份服务器无线通信拨号用户名"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x015BackupServerWIFIUsername.Len, t.T0x015BackupServerWIFIUsername.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x015BackupServerWIFIUsername.Value, t.T0x015BackupServerWIFIUsername.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0016]终端参数ID:22 备份服务器无线通信拨号密码"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x016BackupServerWIFIPassword.Len, t.T0x016BackupServerWIFIPassword.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x016BackupServerWIFIPassword.Value, t.T0x016BackupServerWIFIPassword.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0017]终端参数ID:23 备份服务器地址,IP或域名,以冒号分割主机和端口,多个服务器使用分号分割"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x017BackupServerAddress.Len, t.T0x017BackupServerAddress.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x017BackupServerAddress.Value, t.T0x017BackupServerAddress.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0018]终端参数ID:24 服务器TCP端口（2013版本)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x018TCPPort.Len, t.T0x018TCPPort.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x018TCPPort.Value, t.T0x018TCPPort.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0019]终端参数ID:25 服务器UDP端口 (2013版本)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x019UDPPort.Len, t.T0x019UDPPort.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x019UDPPort.Value, t.T0x019UDPPort.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[001a]终端参数ID:26 道路运输证IC卡认证主服务器IP地址或域名"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x01AICCardAddress.Len, t.T0x01AICCardAddress.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x01AICCardAddress.Value, t.T0x01AICCardAddress.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[001b]终端参数ID:27 道路运输证IC卡认证主服务器TCP端口"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x01BICCardTCPPort.Len, t.T0x01BICCardTCPPort.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x01BICCardTCPPort.Value, t.T0x01BICCardTCPPort.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[001c]终端参数ID:28 路运输证IC卡认证主服务器UDP端口"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x01CICCardUDPPort.Len, t.T0x01CICCardUDPPort.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x01CICCardUDPPort.Value, t.T0x01CICCardUDPPort.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[001d]终端参数ID:29 道路运输证IC卡认证主服务器IP地址或域名,端口同主服务器"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x01DICCardAddress.Len, t.T0x01DICCardAddress.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x01DICCardAddress.Value, t.T0x01DICCardAddress.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0020]终端参数ID:32 位置汇报策略:0.定时汇报 1.定距汇报 2.定时和定距汇报"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x020PositionReportingStrategy.Len, t.T0x020PositionReportingStrategy.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x020PositionReportingStrategy.Value, t.T0x020PositionReportingStrategy.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0021]终端参数ID:33 位置汇报方案:0.根据ACC状态 1.根据登录状态和ACC状态,先判断登录状态,若登录再根据ACC状态"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x021PositionReportingPlan.Len, t.T0x021PositionReportingPlan.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x021PositionReportingPlan.Value, t.T0x021PositionReportingPlan.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0022]终端参数ID:34 驾驶员未登录汇报时间间隔,单位为秒(s),值大于0"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x022DriverReportingInterval.Len, t.T0x022DriverReportingInterval.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x022DriverReportingInterval.Value, t.T0x022DriverReportingInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0023]终端参数ID:35 从服务器APN.该值为空时,终端应使用主服务器相同配置 (2019版本)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x023FromServerAPN.Len, t.T0x023FromServerAPN.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x023FromServerAPN.Value, t.T0x023FromServerAPN.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0024]终端参数ID:36 从服务器无线通信拨号用户名。该值为空时,终端应使用主服务器相同配置 (2019版本)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x024FromServerAPNWIFIUsername.Len, t.T0x024FromServerAPNWIFIUsername.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x024FromServerAPNWIFIUsername.Value, t.T0x024FromServerAPNWIFIUsername.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0025]终端参数ID:37 从服务器无线通信拨号密码.该值为空时,终端应使用主服务器相同配置 (2019版本)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x025FromServerAPNWIFIPassword.Len, t.T0x025FromServerAPNWIFIPassword.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x025FromServerAPNWIFIPassword.Value, t.T0x025FromServerAPNWIFIPassword.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0026]终端参数ID:38 从服务器备份地址、IP或域名,主机和端口用冒号分割,多个服务器使用分号分割 (2019版本)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x026FromServerAPNWIFIAddress.Len, t.T0x026FromServerAPNWIFIAddress.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x026FromServerAPNWIFIAddress.Value, t.T0x026FromServerAPNWIFIAddress.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0027]终端参数ID:39 休眠时汇报时间间隔,单位为秒(s),值大于0"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x027ReportingTimeInterval.Len, t.T0x027ReportingTimeInterval.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x027ReportingTimeInterval.Value, t.T0x027ReportingTimeInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0028]终端参数ID:40 紧急报警时汇报时间间隔,单位为秒(s),值大于0"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x028EmergencyReportingTimeInterval.Len, t.T0x028EmergencyReportingTimeInterval.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x028EmergencyReportingTimeInterval.Value, t.T0x028EmergencyReportingTimeInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0029]终端参数ID:41 缺省时间汇报间隔,单位为秒(s),值大于0"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x029DefaultReportingTimeInterval.Len, t.T0x029DefaultReportingTimeInterval.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x029DefaultReportingTimeInterval.Value, t.T0x029DefaultReportingTimeInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[002C]终端参数ID:44 缺省距离汇报间隔,单位为米(m),值大于0"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x02CDefaultDistanceReportingTimeInterval.Len, t.T0x02CDefaultDistanceReportingTimeInterval.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x02CDefaultDistanceReportingTimeInterval.Value, t.T0x02CDefaultDistanceReportingTimeInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[002D]终端参数ID:45 驾驶员未登录汇报距离间隔,单位为米(m),值大于0"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x02DDrivingReportingDistanceInterval.Len, t.T0x02DDrivingReportingDistanceInterval.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x02DDrivingReportingDistanceInterval.Value, t.T0x02DDrivingReportingDistanceInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[002E]终端参数ID:46 休眠时汇报距离间隔,单位为米(m),值大于0"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x02ESleepReportingDistanceInterval.Len, t.T0x02ESleepReportingDistanceInterval.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x02ESleepReportingDistanceInterval.Value, t.T0x02ESleepReportingDistanceInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[002F]终端参数ID:47 紧急报警时汇报距离间隔,单位为米(m),值大于0"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x02FAlarmReportingDistanceInterval.Len, t.T0x02FAlarmReportingDistanceInterval.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x02FAlarmReportingDistanceInterval.Value, t.T0x02FAlarmReportingDistanceInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0030]终端参数ID:48 拐点补传角度,值小于180度"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x030InflectionPointSupplementaryPassAngle.Len, t.T0x030InflectionPointSupplementaryPassAngle.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x030InflectionPointSupplementaryPassAngle.Value, t.T0x030InflectionPointSupplementaryPassAngle.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0031]终端参数ID:49 电子围栏半径(非法位移阈值),单位为米(m)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x031GeofenceRadius.Len, t.T0x031GeofenceRadius.ID != 0),
		fmt.Sprintf("\t\t[%04x]参数值:[%d]", t.T0x031GeofenceRadius.Value, t.T0x031GeofenceRadius.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0032]终端参数ID:50 违规行驶时段范围,精确到分。(2019版本)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x032IllegalDrivingTime.Len, t.T0x032IllegalDrivingTime.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x032IllegalDrivingTime.Value, t.T0x032IllegalDrivingTime.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0040]终端参数ID:64 监控平台电话号码"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x040MonitoringPlatformPhone.Len, t.T0x040MonitoringPlatformPhone.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x040MonitoringPlatformPhone.Value, t.T0x040MonitoringPlatformPhone.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0041]终端参数ID:65 复位电话号码,可采用此电话号码拨打终端电话让终端复位"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x041ResetPhone.Len, t.T0x041ResetPhone.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x041ResetPhone.Value, t.T0x041ResetPhone.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0042]终端参数ID:65 恢复出厂设置电话号码,可采用此电话号码拨打终端电话让终端恢复出厂设置"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x042RestoreFactoryPhone.Len, t.T0x042RestoreFactoryPhone.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x042RestoreFactoryPhone.Value, t.T0x042RestoreFactoryPhone.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0043]终端参数ID:66 监控平台SMS电话号码"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x043SMSPhone.Len, t.T0x043SMSPhone.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x043SMSPhone.Value, t.T0x043SMSPhone.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0044]终端参数ID:67 接收终端SMS文本报警号码"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x044SMSTxtPhone.Len, t.T0x044SMSTxtPhone.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x044SMSTxtPhone.Value, t.T0x044SMSTxtPhone.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0045]终端参数ID:69 终端电话接听策略,0-自动接听 1-ACC ON时自动接听,OFF时手动接听"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x045TerminalTelephoneStrategy.Len, t.T0x045TerminalTelephoneStrategy.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x045TerminalTelephoneStrategy.Value, t.T0x045TerminalTelephoneStrategy.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0046]终端参数ID:70 每次最长通话时间,单位为秒(s),0为不允许通话,0xFFFFFFFF为不限制"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x046MaximumCallTime.Len, t.T0x046MaximumCallTime.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x046MaximumCallTime.Value, t.T0x046MaximumCallTime.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0047]终端参数ID:71 当月最长通话时间,单位为秒(s),0为不允许通话,0xFFFFFFFF为不限制"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x047MonthMaximumCallTime.Len, t.T0x047MonthMaximumCallTime.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x047MonthMaximumCallTime.Value, t.T0x047MonthMaximumCallTime.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0048]终端参数ID:72 监听电话号码"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x048MonitorPhone.Len, t.T0x048MonitorPhone.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x048MonitorPhone.Value, t.T0x048MonitorPhone.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0049]终端参数ID:73 监管平台特权短信号码"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x049MonitorPrivilegedSMS.Len, t.T0x049MonitorPrivilegedSMS.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x049MonitorPrivilegedSMS.Value, t.T0x049MonitorPrivilegedSMS.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0050]终端参数ID:80 报警屏蔽字.与位置信息汇报消息中的报警标志相对应,相应位为1则相应报警被屏蔽"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x050AlarmBlockingWords.Len, t.T0x050AlarmBlockingWords.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x050AlarmBlockingWords.Value, t.T0x050AlarmBlockingWords.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0051]终端参数ID:81 报警发送文本SMS开关,与位置信息汇报消息中的报警标志相对应,相应位为1则相应报警时发送文本SMS"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x051AlarmSendTextSMSSwitch.Len, t.T0x051AlarmSendTextSMSSwitch.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x051AlarmSendTextSMSSwitch.Value, t.T0x051AlarmSendTextSMSSwitch.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0052]终端参数ID:82 报警拍摄开关,与位置信息汇报消息中的报警标志相对应,相应位为1则相应报警时摄像头拍摄"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x052AlarmShootingSwitch.Len, t.T0x052AlarmShootingSwitch.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x052AlarmShootingSwitch.Value, t.T0x052AlarmShootingSwitch.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0053]终端参数ID:83 报警拍摄存储标志,与位置信息汇报消息中的报警标志相对应,相应位为1则对相应报警时牌的照片进行存储,否则实时长传"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x053AlarmShootingStorageSign.Len, t.T0x053AlarmShootingStorageSign.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x053AlarmShootingStorageSign.Value, t.T0x053AlarmShootingStorageSign.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0054]终端参数ID:84 关键标志,与位置信息汇报消息中的报警标志相对应,相应位为1则对相应报警为关键报警"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x054KeySign.Len, t.T0x054KeySign.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x054KeySign.Value, t.T0x054KeySign.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0055]终端参数ID:85 最高速度,单位为千米每小时(km/h)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x055MaxSpeed.Len, t.T0x055MaxSpeed.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x055MaxSpeed.Value, t.T0x055MaxSpeed.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0056]终端参数ID:86 超速持续时间,单位为秒(s)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x056DurationOverSpeed.Len, t.T0x056DurationOverSpeed.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x056DurationOverSpeed.Value, t.T0x056DurationOverSpeed.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0057]终端参数ID:87 连续驾驶时间门限,单位为秒(s)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x057ContinuousDrivingTimeLimit.Len, t.T0x057ContinuousDrivingTimeLimit.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x057ContinuousDrivingTimeLimit.Value, t.T0x057ContinuousDrivingTimeLimit.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0058]终端参数ID:88 当天累计驾驶时间门限,单位为秒(s)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x058CumulativeDayDrivingTime.Len, t.T0x058CumulativeDayDrivingTime.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x058CumulativeDayDrivingTime.Value, t.T0x058CumulativeDayDrivingTime.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0059]终端参数ID:89 最小休息时间,单位为秒(s)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x059MinimumRestTime.Len, t.T0x059MinimumRestTime.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x059MinimumRestTime.Value, t.T0x059MinimumRestTime.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[005a]终端参数ID:90 最长停车时间,单位为秒(s)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x05AMaximumParkingTime.Len, t.T0x05AMaximumParkingTime.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x05AMaximumParkingTime.Value, t.T0x05AMaximumParkingTime.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[005b]终端参数ID:91 超速预警差值,单位1/10千米每小时(1/10 km/h)"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x05BSpeedWarningDifference.Len, t.T0x05BSpeedWarningDifference.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x05BSpeedWarningDifference.Value, t.T0x05BSpeedWarningDifference.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[005c]终端参数ID:92 疲劳驾驶预警插值,单位为秒(s),值大于0"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x05CFatigueDrivingWarningInterpolation.Len, t.T0x05CFatigueDrivingWarningInterpolation.ID != 0),
		fmt.Sprintf("\t\t[%04x]参数值:[%d]", t.T0x05CFatigueDrivingWarningInterpolation.Value, t.T0x05CFatigueDrivingWarningInterpolation.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[005d]终端参数ID:93 碰撞报警参数设置 b7-b0: 为碰撞时间,单位为毫秒(ms) b15-18 为碰撞加速度,单位为0.1g;设置范围0-79,默认10"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x05DCollisionAlarmParam.Len, t.T0x05DCollisionAlarmParam.ID != 0),
		fmt.Sprintf("\t\t[%04x]参数值:[%d]", t.T0x05DCollisionAlarmParam.Value, t.T0x05DCollisionAlarmParam.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[005e]终端参数ID:94 侧翻报警参数设置:侧翻角度,单位为度,默认为30度"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x05ERolloverAlarmParam.Len, t.T0x05ERolloverAlarmParam.ID != 0),
		fmt.Sprintf("\t\t[%04x]参数值:[%d]", t.T0x05ERolloverAlarmParam.Value, t.T0x05ERolloverAlarmParam.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0064]终端参数ID:100 定时拍照参数,参数项格式和定义见表14"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x064TimedPhotographyParam.Len, t.T0x064TimedPhotographyParam.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x064TimedPhotographyParam.Value, t.T0x064TimedPhotographyParam.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0065]终端参数ID:101 定距拍照参数,参数项格式和定义见表15"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x065FixedDistanceShootingParam.Len, t.T0x065FixedDistanceShootingParam.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x065FixedDistanceShootingParam.Value, t.T0x065FixedDistanceShootingParam.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0070]终端参数ID:112 图像/视频质量,设置范围为1-10,1表示最优质量"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x070ImageVideoQuality.Len, t.T0x070ImageVideoQuality.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x070ImageVideoQuality.Value, t.T0x070ImageVideoQuality.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0071]终端参数ID:113 亮度,设置范围为0-255"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x071Brightness.Len, t.T0x071Brightness.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x071Brightness.Value, t.T0x071Brightness.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0072]终端参数ID:114 对比度,设置范围为0-127"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x072Contrast.Len, t.T0x072Contrast.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x072Contrast.Value, t.T0x072Contrast.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0073]终端参数ID:115 饱和度,设置范围为0-127"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x073Saturation.Len, t.T0x073Saturation.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x073Saturation.Value, t.T0x073Saturation.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0074]终端参数ID:116 色度,设置范围为0-255"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x074Chrominance.Len, t.T0x074Chrominance.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x074Chrominance.Value, t.T0x074Chrominance.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0080]终端参数ID:128 车辆里程表读数,单位:1/10km"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x080VehicleOdometerReadings.Len, t.T0x080VehicleOdometerReadings.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x080VehicleOdometerReadings.Value, t.T0x080VehicleOdometerReadings.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0081]终端参数ID:129 车辆所在的省域ID"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x081VehicleProvinceID.Len, t.T0x081VehicleProvinceID.ID != 0),
		fmt.Sprintf("\t\t[%04x]参数值:[%d]", t.T0x081VehicleProvinceID.Value, t.T0x081VehicleProvinceID.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0082]终端参数ID:130 车辆所在的市域ID"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x082VehicleCityID.Len, t.T0x082VehicleCityID.ID != 0),
		fmt.Sprintf("\t\t[%04x]参数值:[%d]", t.T0x082VehicleCityID.Value, t.T0x082VehicleCityID.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0083]终端参数ID:131 公安交通管理部门颁发的机动车号牌"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x083MotorVehicleLicensePlate.Len, t.T0x083MotorVehicleLicensePlate.ID != 0),
		fmt.Sprintf("\t\t[%x]参数值:[%s]", t.T0x083MotorVehicleLicensePlate.Value, t.T0x083MotorVehicleLicensePlate.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0084]终端参数ID:132 车牌颜色,值按照JT/T 797.7-2014中的规定,未上牌车辆填0"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x084licensePlateColor.Len, t.T0x084licensePlateColor.ID != 0),
		fmt.Sprintf("\t\t[%02x]参数值:[%d]", t.T0x084licensePlateColor.Value, t.T0x084licensePlateColor.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0090]终端参数ID:144 GNSS定位模式 bit0: 0-禁用GPS定位,1-启用 GPS 定位;..."),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x090GNSSPositionMode.Len, t.T0x090GNSSPositionMode.ID != 0),
		fmt.Sprintf("\t\t[%02x]参数值:[%d]", t.T0x090GNSSPositionMode.Value, t.T0x090GNSSPositionMode.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0091]终端参数ID:144 GNSS波特率,定义如下: 0x00:4800;..."),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x091GNSSBaudRate.Len, t.T0x091GNSSBaudRate.ID != 0),
		fmt.Sprintf("\t\t[%02x]参数值:[%d]", t.T0x091GNSSBaudRate.Value, t.T0x091GNSSBaudRate.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0092]终端参数ID:146 GNSS模块详细定位数据输出频率,定义如下:0x01:1000ms(默认值);..."),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x092GNSSModePositionOutputFrequency.Len, t.T0x092GNSSModePositionOutputFrequency.ID != 0),
		fmt.Sprintf("\t\t[%02x]参数值:[%d]", t.T0x092GNSSModePositionOutputFrequency.Value, t.T0x092GNSSModePositionOutputFrequency.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0093]终端参数ID:147 GNSS模块详细定位数据采集频率,单位为秒(s),默认为 1。"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x092GNSSModePositionOutputFrequency.Len, t.T0x092GNSSModePositionOutputFrequency.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x092GNSSModePositionOutputFrequency.Value, t.T0x092GNSSModePositionOutputFrequency.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0094]终端参数ID:148 GNSS模块详细定位数据上传方式 0x00,本地存储,不上传(默认值);..."),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x094GNSSModePositionUploadMethod.Len, t.T0x094GNSSModePositionUploadMethod.ID != 0),
		fmt.Sprintf("\t\t[%02x]参数值:[%d]", t.T0x094GNSSModePositionUploadMethod.Value, t.T0x094GNSSModePositionUploadMethod.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0095]终端参数ID:149 GNSS模块详细定位数据上传设置, 关联0x0094 上传方式为 0x01 时,单位为秒;..."),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x095GNSSModeSetPositionUpload.Len, t.T0x095GNSSModeSetPositionUpload.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x095GNSSModeSetPositionUpload.Value, t.T0x095GNSSModeSetPositionUpload.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0100]终端参数ID:256 CAN总线通道1采集时间间隔(ms),0表示不采集"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x100CANCollectionTimeInterval.Len, t.T0x100CANCollectionTimeInterval.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x100CANCollectionTimeInterval.Value, t.T0x100CANCollectionTimeInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0101]终端参数ID:257 CAN总线通道1上传时间间隔(s),0表示不上传"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x101CAN1UploadTimeInterval.Len, t.T0x101CAN1UploadTimeInterval.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x101CAN1UploadTimeInterval.Value, t.T0x101CAN1UploadTimeInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0102]终端参数ID:258 CAN总线通道2采集时间间隔(ms),0表示不采集"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x102CAN2CollectionTimeInterval.Len, t.T0x102CAN2CollectionTimeInterval.ID != 0),
		fmt.Sprintf("\t\t[%08x]参数值:[%d]", t.T0x102CAN2CollectionTimeInterval.Value, t.T0x102CAN2CollectionTimeInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0103]终端参数ID:259 CAN总线通道2上传时间间隔(s),0表示不上传"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x103CAN2UploadTimeInterval.Len, t.T0x103CAN2UploadTimeInterval.ID != 0),
		fmt.Sprintf("\t\t[%04x]参数值:[%d]", t.T0x103CAN2UploadTimeInterval.Value, t.T0x103CAN2UploadTimeInterval.Value),
		"\t}",
		"\t{",
		fmt.Sprintf("\t\t[0110]终端参数ID:272 CAN总线ID单独采集设置:"),
		fmt.Sprintf("\t\t参数长度[%d] 是否存在[%t]", t.T0x110CANIDSetIndividualAcquisition.Len, t.T0x110CANIDSetIndividualAcquisition.ID != 0),
		fmt.Sprintf("\t\t[%04x]参数值:[%d]", t.T0x110CANIDSetIndividualAcquisition.Value, t.T0x110CANIDSetIndividualAcquisition.Value),
		"\t}",
	}, "\n")

	if len(t.OtherContent) > 0 {
		ids := make([]int, 0, len(t.OtherContent))
		for id, _ := range t.OtherContent {
			ids = append(ids, int(id))
		}
		sort.Ints(ids)
		str += fmt.Sprintf("\n\t未知终端参数id:%v\n", ids)
	}
	return str
}
