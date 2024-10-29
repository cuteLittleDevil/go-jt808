package consts

// JT808CommandType jt808协议的指令
type JT808CommandType uint16

const (
	// T0001GeneralRespond 终端-通用应答
	T0001GeneralRespond JT808CommandType = 0x0001
	// T0002HeartBeat 终端-心跳
	T0002HeartBeat JT808CommandType = 0x0002
	// T0100Register 终端-注册
	T0100Register JT808CommandType = 0x0100
	// T0102RegisterAuth 终端-注册鉴权
	T0102RegisterAuth JT808CommandType = 0x0102
	// T0104QueryParameter 终端-查询参数
	T0104QueryParameter JT808CommandType = 0x0104
	// T0107QueryAttribute 终端-查询属性
	T0107QueryAttribute JT808CommandType = 0x0107
	// T0108UpgradeNotice 终端-升级通知
	T0108UpgradeNotice JT808CommandType = 0x0108
	// T0200LocationReport 终端-位置上报 经纬度坐标是WGS84
	T0200LocationReport JT808CommandType = 0x0200
	// T0201QueryLocation 终端-查询位置
	T0201QueryLocation JT808CommandType = 0x0201
	// T0301EventReport 终端-事件上报
	T0301EventReport JT808CommandType = 0x0301
	// T0302QuestionAnswer 终端-提问应答
	T0302QuestionAnswer JT808CommandType = 0x0302
	// T0303MessagePlayCancel 终端-消息点播取消
	T0303MessagePlayCancel JT808CommandType = 0x0303
	// T0608QueryRegionRespond 终端-查询区域应答
	T0608QueryRegionRespond JT808CommandType = 0x0608
	// T0700DrivingRecordUpload 终端-行驶记录上传
	T0700DrivingRecordUpload JT808CommandType = 0x0700
	// T0702DriverInfoCollectReport 终端-驾驶员信息采集上报
	T0702DriverInfoCollectReport JT808CommandType = 0x0702
	// T0704LocationBatchUpload 终端-位置批量上传
	T0704LocationBatchUpload JT808CommandType = 0x0704
	// T0800MultimediaEventInfoUpload 终端-多媒体事件信息上传
	T0800MultimediaEventInfoUpload JT808CommandType = 0x0800
	// T0801MultimediaDataUpload 终端-多媒体数据上传
	T0801MultimediaDataUpload JT808CommandType = 0x0801
	// T0805CameraShootImmediately 终端-摄像头立即拍照
	T0805CameraShootImmediately JT808CommandType = 0x0805
	// T0900DataUpTransparentTransmission 终端-数据上行透传
	T0900DataUpTransparentTransmission JT808CommandType = 0x0900

	// P8001GeneralRespond 平台-通用应答
	P8001GeneralRespond JT808CommandType = 0x8001
	// P8003ReissueSubcontractingRequest 平台-补发分包请求
	P8003ReissueSubcontractingRequest JT808CommandType = 0x8003
	// P8004QueryTimeRespond 平台-查询时间应答
	P8004QueryTimeRespond JT808CommandType = 0x8004
	// P8100RegisterRespond 平台-注册应答
	P8100RegisterRespond JT808CommandType = 0x8100
	// P8103SetTerminalParams 平台-设置终端参数
	P8103SetTerminalParams JT808CommandType = 0x8103
	// P8104QueryTerminalParams 平台-查询终端参数
	P8104QueryTerminalParams JT808CommandType = 0x8104
	// P8105TerminalControl 平台-终端控制
	P8105TerminalControl JT808CommandType = 0x8105
	// P8106QuerySpecifyParam 平台-查询指定参数
	P8106QuerySpecifyParam JT808CommandType = 0x8106
	// P8107QueryTerminalProperties 平台-查询终端属性
	P8107QueryTerminalProperties JT808CommandType = 0x8107
	// P8108DistributeTerminalUpgradePackage 平台-下发终端升级包
	P8108DistributeTerminalUpgradePackage JT808CommandType = 0x8108
	// P8201QueryLocation 平台-查询位置
	P8201QueryLocation JT808CommandType = 0x8201
	// P8202TmpLocationTrack 平台-临时位置跟踪
	P8202TmpLocationTrack JT808CommandType = 0x8202
	// P8203ManuallyConfirmAlarmInfo 平台-人工确认报警信息
	P8203ManuallyConfirmAlarmInfo JT808CommandType = 0x8203
	// P8300TextInfoDistribution 平台-文本信息下发
	P8300TextInfoDistribution JT808CommandType = 0x8300
	// P8301EventSetting 平台-事件设置
	P8301EventSetting JT808CommandType = 0x8301
	// P8302QuestionDistribution 平台-提问下发
	P8302QuestionDistribution JT808CommandType = 0x8302
	// P8303InfoPlaySetting 平台-信息点播设置
	P8303InfoPlaySetting JT808CommandType = 0x8303
	// P8304InfoService 平台-信息服务
	P8304InfoService JT808CommandType = 0x8304
	// P8400PhoneCallBack 平台-电话回拨
	P8400PhoneCallBack JT808CommandType = 0x8400
	// P8401SetPhoneBook 平台-设置电话本
	P8401SetPhoneBook JT808CommandType = 0x8401
	// P8500VehicleControl 平台-车辆控制
	P8500VehicleControl JT808CommandType = 0x8500
	// P8600SetCircularArea 平台-设置圆形区域
	P8600SetCircularArea JT808CommandType = 0x8600
	// P8601DeleteArea 平台-删除区域
	P8601DeleteArea JT808CommandType = 0x8601
	// P8602SetRectArea 平台-设置矩形区域
	P8602SetRectArea JT808CommandType = 0x8602
	// P8604PolygonArea 平台-设置多边形区域
	P8604PolygonArea JT808CommandType = 0x8604
	// P8606SetRoute 平台-设置路线
	P8606SetRoute JT808CommandType = 0x8606
	// P8608QueryAreaOrRouteData 平台-查询区域或路线数据
	P8608QueryAreaOrRouteData JT808CommandType = 0x8608
	// P8701DrivingRecordParamDistribution 平台-行驶记录仪参数下发
	P8701DrivingRecordParamDistribution JT808CommandType = 0x8701
	// P8800MultimediaUploadRespond 平台-多媒体上传应答
	P8800MultimediaUploadRespond JT808CommandType = 0x8800
	// P8801CameraShootImmediateCommand 平台-摄像头立即拍摄命令
	P8801CameraShootImmediateCommand JT808CommandType = 0x8801
	// P8802StorageMultimediaDataRetrieval 平台-存储多媒体数据检索
	P8802StorageMultimediaDataRetrieval JT808CommandType = 0x8802
	// P8803StorageMultimediaDataUpload 平台-存储多媒体数据上传
	P8803StorageMultimediaDataUpload JT808CommandType = 0x8803
	// P8804SoundRecordStartCommand 平台-录音开始命令
	P8804SoundRecordStartCommand JT808CommandType = 0x8804
	// P8805SingleMultimediaDataRetrieval 平台-单条多媒体数据检索
	P8805SingleMultimediaDataRetrieval JT808CommandType = 0x8805
	// P8900DataDownTransparentTransmission 平台-数据下行透传
	P8900DataDownTransparentTransmission JT808CommandType = 0x8900
)

// JT1078的
const (
	// T1003UploadAudioVideoAttr 终端-上传音视频属性
	T1003UploadAudioVideoAttr JT808CommandType = 0x1003
	// T1005UploadPassengerFlow 终端-上传乘客流量
	T1005UploadPassengerFlow JT808CommandType = 0x1005
	// T1205UploadAudioVideoResourceList 终端-上传音视频资源列表
	T1205UploadAudioVideoResourceList JT808CommandType = 0x1205
	// T1206FileUploadCompleteNotice 终端-文件上传完成通知
	T1206FileUploadCompleteNotice JT808CommandType = 0x1206

	// P9003QueryTerminalAudioVideoProperties 平台-查询终端音视频属性
	P9003QueryTerminalAudioVideoProperties JT808CommandType = 0x9003
	// P9101RealTimeAudioVideoRequest 平台-实时音视频传输请求
	P9101RealTimeAudioVideoRequest JT808CommandType = 0x9101
	// P9102AudioVideoControl 平台-音视频实时传输控制
	P9102AudioVideoControl JT808CommandType = 0x9102
	// P9105AudioVideoControlStatusNotice 平台-音视频实时传输状态通知
	P9105AudioVideoControlStatusNotice JT808CommandType = 0x9105
	// P9201SendVideoRecordRequest 平台-下发远程录像回放请求
	P9201SendVideoRecordRequest JT808CommandType = 0x9201
	// P9202SendVideoRecordControl 平台-下发远程录像回放控制
	P9202SendVideoRecordControl JT808CommandType = 0x9202
	// P9205QueryResourceList 平台-查询资源列表
	P9205QueryResourceList JT808CommandType = 0x9205
	// P9206FileUploadInstructions 平台-文件上传指令
	P9206FileUploadInstructions JT808CommandType = 0x9206
	// P9207FileUploadControl 平台-文件上传控制
	P9207FileUploadControl JT808CommandType = 0x9207
)

func (j JT808CommandType) String() string {
	switch j {
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

	switch j {
	case T1003UploadAudioVideoAttr:
		return "终端-上传音视频属性"
	case T1005UploadPassengerFlow:
		return "终端-上传乘客流量"
	case T1205UploadAudioVideoResourceList:
		return "终端-上传音视频资源列表"
	case T1206FileUploadCompleteNotice:
		return "终端-文件上传完成通知"
	case P9003QueryTerminalAudioVideoProperties:
		return "平台-查询终端音视频属性"
	case P9101RealTimeAudioVideoRequest:
		return "平台-实时音视频传输请求"
	case P9102AudioVideoControl:
		return "平台-音视频实时传输控制"
	case P9105AudioVideoControlStatusNotice:
		return "平台-音视频实时传输状态通知"
	case P9201SendVideoRecordRequest:
		return "平台-下发远程录像回放请求"
	case P9202SendVideoRecordControl:
		return "平台-下发远程录像回放控制"
	case P9205QueryResourceList:
		return "平台-查询资源列表"
	case P9206FileUploadInstructions:
		return "平台-文件上传指令"
	case P9207FileUploadControl:
		return "平台-文件上传控制"
	}

	return "平台-暂未实现的命令"
}
