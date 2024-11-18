package consts

// 主动安全扩展的
const (
	// T1210AlarmAttachInfoMessage 终端-报警附件信息消息
	T1210AlarmAttachInfoMessage JT808CommandType = 0x1210
	// T1211FileInfoUpload 终端-文件信息上传
	T1211FileInfoUpload JT808CommandType = 0x1211
	// T1212FileUploadComplete 终端-文件上传完成消息
	T1212FileUploadComplete JT808CommandType = 0x1212

	// P9208AlarmAttachUpload 平台-报警附件上传指令
	P9208AlarmAttachUpload JT808CommandType = 0x9208
	// P9212FileUploadCompleteRespond 平台-文件上传完成消息应答
	P9212FileUploadCompleteRespond JT808CommandType = 0x9212
	// T1FC4TerminalUpgradeProgressReport 终端-升级进度上报
	T1FC4TerminalUpgradeProgressReport JT808CommandType = 0x1FC4
)
