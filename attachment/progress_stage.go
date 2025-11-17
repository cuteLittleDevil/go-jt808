package attachment

type ProgressStage uint8

const (
	ProgressStageInit ProgressStage = iota + 1
	ProgressStageStart
	ProgressStageStreamData
	ProgressStageSupplementary
	ProgressStageStreamDataComplete
	ProgressStageComplete
	ProgressStageSuccessQuit
	ProgressStageFailQuit
	ProgressStageUnexpectedExit
)

func (p ProgressStage) String() string {
	switch p {
	case ProgressStageInit:
		return "初始化状态-收到0x1210"
	case ProgressStageStart:
		return "开始状态-收到0x1211"
	case ProgressStageStreamData:
		return "文件码流数据收集中"
	case ProgressStageSupplementary:
		return "补传状态-等待最终完成"
	case ProgressStageStreamDataComplete:
		return "文件码流数据收集完成"
	case ProgressStageComplete:
		return "完成状态-收到0x1212 并且没有需要补传的"
	case ProgressStageSuccessQuit:
		return "成功退出状态-所有的文件都接收完成"
	case ProgressStageFailQuit:
		return "异常退出状态-文件没有全部接收完成"
	case ProgressStageUnexpectedExit:
		return "意外退出状态-手机号都未解析"
	}
	return ""
}
