package consts

type PlatformReplyType uint16

const (
	P8001GeneralRespond                   PlatformReplyType = 0x8001 // 平台-通用应答
	P8003ReissueSubcontractingRequest     PlatformReplyType = 0x8003 // 平台-补发分包请求
	P8004QueryTimeRespond                 PlatformReplyType = 0x8004 // 平台-查询时间应答
	P8100RegisterRespond                  PlatformReplyType = 0x8100 // 平台-注册应答
	P8103SetTerminalParams                PlatformReplyType = 0x8103 // 平台-设置终端参数
	P8104QueryTerminalParams              PlatformReplyType = 0x8104 // 平台-查询终端参数
	P8105TerminalControl                  PlatformReplyType = 0x8105 // 平台-终端控制
	P8106QuerySpecifyParam                PlatformReplyType = 0x8106 // 平台-查询指定参数
	P8107QueryTerminalProperties          PlatformReplyType = 0x8107 // 平台-查询终端属性
	P8108DistributeTerminalUpgradePackage PlatformReplyType = 0x8108 // 平台-下发终端升级包
	P8201QueryLocation                    PlatformReplyType = 0x8201 // 平台-查询位置
	P8202TmpLocationTrack                 PlatformReplyType = 0x8202 // 平台-临时位置跟踪
	P8203ManuallyConfirmAlarmInfo         PlatformReplyType = 0x8203 // 平台-人工确认报警信息
	P8300TextInfoDistribution             PlatformReplyType = 0x8300 // 平台-文本信息下发
	P8301EventSetting                     PlatformReplyType = 0x8301 // 平台-事件设置
	P8302QuestionDistribution             PlatformReplyType = 0x8302 // 平台-提问下发
	P8303InfoPlaySetting                  PlatformReplyType = 0x8303 // 平台-信息点播设置
	P8304InfoService                      PlatformReplyType = 0x8304 // 平台-信息服务
	P8400PhoneCallBack                    PlatformReplyType = 0x8400 // 平台-电话回拨
	P8401SetPhoneBook                     PlatformReplyType = 0x8401 // 平台-设置电话本
	P8500VehicleControl                   PlatformReplyType = 0x8500 // 平台-车辆控制
	P8600SetCircularArea                  PlatformReplyType = 0x8600 // 平台-设置圆形区域
	P8601DeleteArea                       PlatformReplyType = 0x8601 // 平台-删除区域
	P8602SetRectArea                      PlatformReplyType = 0x8602 // 平台-设置矩形区域
	P8604PolygonArea                      PlatformReplyType = 0x8604 // 平台-设置多边形区域
	P8606SetRoute                         PlatformReplyType = 0x8606 // 平台-设置路线
	P8608QueryAreaOrRouteData             PlatformReplyType = 0x8608 // 平台-查询区域或路线数据
	P8701DrivingRecordParamDistribution   PlatformReplyType = 0x8701 // 平台-行驶记录仪参数下发
	P8800MultimediaUploadRespond          PlatformReplyType = 0x8800 // 平台-多媒体上传应答
	P8801CameraShootImmediateCommand      PlatformReplyType = 0x8801 // 平台-摄像头立即拍摄命令
	P8802StorageMultimediaDataRetrieval   PlatformReplyType = 0x8802 // 平台-存储多媒体数据检索
	P8803StorageMultimediaDataUpload      PlatformReplyType = 0x8803 // 平台-存储多媒体数据上传
	P8804SoundRecordStartCommand          PlatformReplyType = 0x8804 // 平台-录音开始命令
	P8805SingleMultimediaDataRetrieval    PlatformReplyType = 0x8805 // 平台-单条多媒体数据检索
	P8900DataDownTransparentTransmission  PlatformReplyType = 0x8900 // 平台-数据下行透传
)

func (s PlatformReplyType) String() string {
	switch s {
	case P8001GeneralRespond:
		return "平台-通用应答"
	case P8003ReissueSubcontractingRequest:
		return "平台-补发分包请求"
	case P8004QueryTimeRespond:
		return "平台-查询时间应答"
	case P8100RegisterRespond:
		return "平台-注册应答"
	case P8103SetTerminalParams:
		return "平台-设置终端参数"
	case P8104QueryTerminalParams:
		return "平台-查询终端参数"
	case P8105TerminalControl:
		return "平台-终端控制"
	case P8106QuerySpecifyParam:
		return "平台-查询指定参数"
	case P8107QueryTerminalProperties:
		return "平台-查询终端属性"
	case P8108DistributeTerminalUpgradePackage:
		return "平台-下发终端升级包"
	case P8201QueryLocation:
		return "平台-查询位置"
	case P8202TmpLocationTrack:
		return "平台-临时定位轨迹"
	case P8203ManuallyConfirmAlarmInfo:
		return "平台-人工确认报警信息"
	case P8300TextInfoDistribution:
		return "平台-文本信息下发"
	case P8301EventSetting:
		return "平台-事件设置"
	case P8302QuestionDistribution:
		return "平台-提问下发"
	case P8303InfoPlaySetting:
		return "平台-信息点播设置"
	case P8304InfoService:
		return "平台-信息服务"
	case P8400PhoneCallBack:
		return "平台-电话回拨"
	case P8401SetPhoneBook:
		return "平台-设置电话本"
	case P8500VehicleControl:
		return "平台-车辆控制"
	case P8600SetCircularArea:
		return "平台-设置圆形区域"
	case P8601DeleteArea:
		return "平台-删除区域"
	case P8602SetRectArea:
		return "平台-设置矩形区域"
	case P8604PolygonArea:
		return "平台-设置多边形区域"
	case P8606SetRoute:
		return "平台-设置路线"
	case P8608QueryAreaOrRouteData:
		return "平台-查询区域或路线数据"
	case P8701DrivingRecordParamDistribution:
		return "平台-行驶记录仪参数下发"
	case P8800MultimediaUploadRespond:
		return "平台-多媒体上传应答"
	case P8801CameraShootImmediateCommand:
		return "平台-摄像头立即拍照命令"
	case P8802StorageMultimediaDataRetrieval:
		return "平台-存储多媒体数据检索"
	case P8803StorageMultimediaDataUpload:
		return "平台-存储多媒体数据上传"
	case P8804SoundRecordStartCommand:
		return "平台-录音开始命令"
	case P8805SingleMultimediaDataRetrieval:
		return "平台-单条多媒体数据检索"
	case P8900DataDownTransparentTransmission:
		return "平台-数据下行透传"
	}
	return "平台-暂未实现的命令"
}
