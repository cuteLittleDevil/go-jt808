package consts

type TerminalRequestType uint16

const (
	T0001GeneralRespond                TerminalRequestType = 0x0001 // 终端-通用应答
	T0002HeartBeat                     TerminalRequestType = 0x0002 // 终端-心跳
	T0100Register                      TerminalRequestType = 0x0100 // 终端-注册
	T0102RegisterAuth                  TerminalRequestType = 0x0102 // 终端-注册鉴权
	T0104QueryParameter                TerminalRequestType = 0x0104 // 终端-查询参数
	T0107QueryAttribute                TerminalRequestType = 0x0107 // 终端-查询属性
	T0108UpgradeNotice                 TerminalRequestType = 0x0108 // 终端-升级通知
	T0200LocationReport                TerminalRequestType = 0x0200 // 终端-位置上报 经纬度坐标是WGS84
	T0201QueryLocation                 TerminalRequestType = 0x0201 // 终端-查询位置
	T0301EventReport                   TerminalRequestType = 0x0301 // 终端-事件上报
	T0302QuestionAnswer                TerminalRequestType = 0x0302 // 终端-提问应答
	T0303MessagePlayCancel             TerminalRequestType = 0x0303 // 终端-消息点播取消
	T0608QueryRegionRespond            TerminalRequestType = 0x0608 // 终端-查询区域应答
	T0700DrivingRecordUpload           TerminalRequestType = 0x0700 // 终端-行驶记录上传
	T0702DriverInfoCollectReport       TerminalRequestType = 0x0702 // 终端-驾驶员信息采集上报
	T0704LocationBatchUpload           TerminalRequestType = 0x0704 // 终端-位置批量上传
	T0800MultimediaEventInfoUpload     TerminalRequestType = 0x0800 // 终端-多媒体事件信息上传
	T0801MultimediaDataUpload          TerminalRequestType = 0x0801 // 终端-多媒体数据上传
	T0805CameraShootImmediately        TerminalRequestType = 0x0805 // 终端-摄像头立即拍照
	T0900DataUpTransparentTransmission TerminalRequestType = 0x0900 // 终端-数据上行透传
)

func (t TerminalRequestType) String() string {
	switch t {
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
	}
	return "终端-暂未实现的协议"
}
