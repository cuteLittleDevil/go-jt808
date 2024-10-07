package consts

type JT808CommandType uint16

const (
	T0001GeneralRespond                JT808CommandType = 0x0001 // 终端-通用应答
	T0002HeartBeat                     JT808CommandType = 0x0002 // 终端-心跳
	T0100Register                      JT808CommandType = 0x0100 // 终端-注册
	T0102RegisterAuth                  JT808CommandType = 0x0102 // 终端-注册鉴权
	T0104QueryParameter                JT808CommandType = 0x0104 // 终端-查询参数
	T0107QueryAttribute                JT808CommandType = 0x0107 // 终端-查询属性
	T0108UpgradeNotice                 JT808CommandType = 0x0108 // 终端-升级通知
	T0200LocationReport                JT808CommandType = 0x0200 // 终端-位置上报 经纬度坐标是WGS84
	T0201QueryLocation                 JT808CommandType = 0x0201 // 终端-查询位置
	T0301EventReport                   JT808CommandType = 0x0301 // 终端-事件上报
	T0302QuestionAnswer                JT808CommandType = 0x0302 // 终端-提问应答
	T0303MessagePlayCancel             JT808CommandType = 0x0303 // 终端-消息点播取消
	T0608QueryRegionRespond            JT808CommandType = 0x0608 // 终端-查询区域应答
	T0700DrivingRecordUpload           JT808CommandType = 0x0700 // 终端-行驶记录上传
	T0702DriverInfoCollectReport       JT808CommandType = 0x0702 // 终端-驾驶员信息采集上报
	T0704LocationBatchUpload           JT808CommandType = 0x0704 // 终端-位置批量上传
	T0800MultimediaEventInfoUpload     JT808CommandType = 0x0800 // 终端-多媒体事件信息上传
	T0801MultimediaDataUpload          JT808CommandType = 0x0801 // 终端-多媒体数据上传
	T0805CameraShootImmediately        JT808CommandType = 0x0805 // 终端-摄像头立即拍照
	T0900DataUpTransparentTransmission JT808CommandType = 0x0900 // 终端-数据上行透传

	P8001GeneralRespond                   JT808CommandType = 0x8001 // 平台-通用应答
	P8003ReissueSubcontractingRequest     JT808CommandType = 0x8003 // 平台-补发分包请求
	P8004QueryTimeRespond                 JT808CommandType = 0x8004 // 平台-查询时间应答
	P8100RegisterRespond                  JT808CommandType = 0x8100 // 平台-注册应答
	P8103SetTerminalParams                JT808CommandType = 0x8103 // 平台-设置终端参数
	P8104QueryTerminalParams              JT808CommandType = 0x8104 // 平台-查询终端参数
	P8105TerminalControl                  JT808CommandType = 0x8105 // 平台-终端控制
	P8106QuerySpecifyParam                JT808CommandType = 0x8106 // 平台-查询指定参数
	P8107QueryTerminalProperties          JT808CommandType = 0x8107 // 平台-查询终端属性
	P8108DistributeTerminalUpgradePackage JT808CommandType = 0x8108 // 平台-下发终端升级包
	P8201QueryLocation                    JT808CommandType = 0x8201 // 平台-查询位置
	P8202TmpLocationTrack                 JT808CommandType = 0x8202 // 平台-临时位置跟踪
	P8203ManuallyConfirmAlarmInfo         JT808CommandType = 0x8203 // 平台-人工确认报警信息
	P8300TextInfoDistribution             JT808CommandType = 0x8300 // 平台-文本信息下发
	P8301EventSetting                     JT808CommandType = 0x8301 // 平台-事件设置
	P8302QuestionDistribution             JT808CommandType = 0x8302 // 平台-提问下发
	P8303InfoPlaySetting                  JT808CommandType = 0x8303 // 平台-信息点播设置
	P8304InfoService                      JT808CommandType = 0x8304 // 平台-信息服务
	P8400PhoneCallBack                    JT808CommandType = 0x8400 // 平台-电话回拨
	P8401SetPhoneBook                     JT808CommandType = 0x8401 // 平台-设置电话本
	P8500VehicleControl                   JT808CommandType = 0x8500 // 平台-车辆控制
	P8600SetCircularArea                  JT808CommandType = 0x8600 // 平台-设置圆形区域
	P8601DeleteArea                       JT808CommandType = 0x8601 // 平台-删除区域
	P8602SetRectArea                      JT808CommandType = 0x8602 // 平台-设置矩形区域
	P8604PolygonArea                      JT808CommandType = 0x8604 // 平台-设置多边形区域
	P8606SetRoute                         JT808CommandType = 0x8606 // 平台-设置路线
	P8608QueryAreaOrRouteData             JT808CommandType = 0x8608 // 平台-查询区域或路线数据
	P8701DrivingRecordParamDistribution   JT808CommandType = 0x8701 // 平台-行驶记录仪参数下发
	P8800MultimediaUploadRespond          JT808CommandType = 0x8800 // 平台-多媒体上传应答
	P8801CameraShootImmediateCommand      JT808CommandType = 0x8801 // 平台-摄像头立即拍摄命令
	P8802StorageMultimediaDataRetrieval   JT808CommandType = 0x8802 // 平台-存储多媒体数据检索
	P8803StorageMultimediaDataUpload      JT808CommandType = 0x8803 // 平台-存储多媒体数据上传
	P8804SoundRecordStartCommand          JT808CommandType = 0x8804 // 平台-录音开始命令
	P8805SingleMultimediaDataRetrieval    JT808CommandType = 0x8805 // 平台-单条多媒体数据检索
	P8900DataDownTransparentTransmission  JT808CommandType = 0x8900 // 平台-数据下行透传
)

func (s JT808CommandType) String() string {
	switch s {
	case T0001GeneralRespond:
		return "终端-通用应答"
	case T0002HeartBeat:
		return "终端-心跳"
	case T0100Register:
		return "终端-注册"
	case T0102RegisterAuth:
		return "终端-注册鉴权"
	case T0104QueryParameter:
		return "终端-查询参数"
	case T0107QueryAttribute:
		return "终端-查询属性"
	case T0108UpgradeNotice:
		return "终端-升级通知"
	case T0200LocationReport:
		return "终端-位置上报"
	case T0201QueryLocation:
		return "终端-查询位置"
	case T0301EventReport:
		return "终端-事件上报"
	case T0302QuestionAnswer:
		return "终端-提问应答"
	case T0303MessagePlayCancel:
		return "终端-消息点播取消"
	case T0608QueryRegionRespond:
		return "终端-查询区域应答"
	case T0700DrivingRecordUpload:
		return "终端-行驶记录仪上传"
	case T0702DriverInfoCollectReport:
		return "终端-驾驶员信息采集上报"
	case T0704LocationBatchUpload:
		return "终端-位置批量上传"
	case T0800MultimediaEventInfoUpload:
		return "终端-多媒体事件信息上传"
	case T0801MultimediaDataUpload:
		return "终端-多媒体数据上传"
	case T0805CameraShootImmediately:
		return "终端-摄像头立即拍照"
	case T0900DataUpTransparentTransmission:
		return "终端-数据上行透传"
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
