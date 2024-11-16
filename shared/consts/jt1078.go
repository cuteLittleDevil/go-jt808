package consts

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
