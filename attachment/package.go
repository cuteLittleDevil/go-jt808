package attachment

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
)

type (
	PackageProgress struct {
		// ProgressStage 当前进度
		ProgressStage
		// Record 目前上传数据的记录 key=文件名 value=数据包上传情况
		Record map[string]*Package
		// ExtensionFields 扩展字段信息
		ExtensionFields ExtensionFields
		historyData     []byte
		handle          JT808DataHandler
		streamFunc      func() StreamDataHandler
	}

	ExtensionFields struct {
		// CurrentPackage 当前完成的包情况 也在record中的
		CurrentPackage *Package
		// RecentTerminalMessage 终端主动上传的808数据
		RecentTerminalMessage *jt808.JTMessage
		// RecentPlatformData 平台下发的数据
		RecentPlatformData []byte `json:"-"`
		// Err 异常情况
		Err error
	}

	Package struct {
		// FileName 文件名称
		FileName string
		// FileSize 文件大小
		FileSize uint32
		// CurrentSize 已经上传的文件大小
		CurrentSize uint32
		// StreamHead 数据头部 当数据完成时候记录
		StreamHead []byte
		// StreamBody 数据体 当数据完成时候记录
		StreamBody []byte
		// OffsetDataRecord 偏移的数据记录 key=偏移 value=数据
		OffsetDataRecord map[int][]byte
		// OffsetRecord 偏移的记录 key=偏移 value=文件大小
		OffsetRecord map[int]int
	}
)

func (p *PackageProgress) switchState(curData []byte) (bool, error) {
	p.historyData = append(p.historyData, curData...)
	return false, ErrDataInconsistency
}

func (p *PackageProgress) stageStreamData() (bool, error) {
	return false, ErrInsufficientDataLen
}

func (p *PackageProgress) hasJT808Reply() bool {
	switch p.ProgressStage {
	case ProgressStageInit, ProgressStageStart, ProgressStageComplete:
		return true
	default:
	}
	return false
}
