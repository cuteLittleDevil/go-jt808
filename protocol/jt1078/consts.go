package jt1078

import "fmt"

type PTType uint8

const (
	PTG711A PTType = 6
	PTG711U PTType = 7
	PTAAC   PTType = 19
	PTMP3   PTType = 25
	PTH264  PTType = 98
	PTH265  PTType = 99
)

func (p PTType) String() string {
	switch p {
	case PTG711A:
		return "G711A"
	case PTG711U:
		return "G711U"
	case PTAAC:
		return "AAC"
	case PTMP3:
		return "MP3"
	case PTH264:
		return "H264"
	case PTH265:
		return "H265"
	}
	return fmt.Sprintf("未知类型:%d", uint8(p))
}

type DataType uint8

const (
	DataTypeI DataType = iota
	DataTypeP
	DataTypeB
	DataTypeA
	DataTypePenetrate
)

func (d DataType) String() string {
	switch d {
	case DataTypeI:
		return "视频I祯"
	case DataTypeP:
		return "视频P帧"
	case DataTypeB:
		return "视频B帧"
	case DataTypeA:
		return "音频帧"
	case DataTypePenetrate:
		return "透传数据"
	}
	return fmt.Sprintf("未知类型:%d", uint8(d))
}

type SubcontractType uint8

const (
	SubcontractTypeAtomic SubcontractType = iota
	SubcontractTypeFirst
	SubcontractTypeLast
	SubcontractTypeMiddle
)

func (s SubcontractType) String() string {
	switch s {
	case SubcontractTypeAtomic:
		return "原子祯"
	case SubcontractTypeFirst:
		return "第一祯"
	case SubcontractTypeLast:
		return "最后祯"
	case SubcontractTypeMiddle:
		return "中间祯"
	}
	return fmt.Sprintf("未知类型:%d", uint8(s))
}
