package attachment

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type (
	DataHandler interface {
		Parse(jtMsg *jt808.JTMessage) error               // 解析终端上传的body数据
		OnPackageProgressEvent(progress *PackageProgress) // 进度事件 修改当前进度 初始化附件
		ReplyData() ([]byte, error)                       // 终端主动上传的情况 平台回复给终端的指令数据
		CreateStreamDataHandler() StreamDataHandler
	}

	JT808Handler interface {
		Parse(jtMsg *jt808.JTMessage) error
		ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error)
		ReplyProtocol() consts.JT808CommandType
	}

	StreamDataHandler interface {
		HasStreamData(data []byte) bool                 // 是不是流数据
		HasMinHeadLen(data []byte) bool                 // 是不是可处理的最小长度了
		Parse(data []byte) (headLen int, bodyLen int)   // 解析数据 获取数据长度
		OnInitEvent(progress *PackageProgress)          // 数据完整解析 初始化
		GetFileName() string                            // 获取文件名
		GetDataOffsetAndLen() (offset int, dataLen int) // 获取当前传输文件的数据偏移量 和 本次数据的长度
	} // 	}

	FileEventer interface {
		OnEvent(*PackageProgress)
	}
)
