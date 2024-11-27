package attachment

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

type (
	JT808DataHandler interface {
		Parse(jtMsg *jt808.JTMessage) error           // 解析终端上传的body数据
		OnPackageProgressEvent(pack *PackageProgress) // 进度事件 修改当前进度 初始化附件
		ReplyData() ([]byte, error)                   // 终端主动上传的情况 平台回复给终端的指令数据
	}

	JT808Handler interface {
		Parse(jtMsg *jt808.JTMessage) error
		ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error)
		ReplyProtocol() consts.JT808CommandType
	}

	StreamDataHandler interface {
		HasMinHeadLen(data []byte) bool                 // 是不是可处理的最小长度了
		HasStreamData(data []byte) bool                 // 是不是流数据
		GetLen(data []byte) (headLen int, bodyLen int)  // 数据长度是否解析好
		GetFileName() string                            // 获取文件名
		GetDataOffsetAndLen() (offset int, dataLen int) // 获取当前传输文件的数据偏移量 和 本次数据的长度
	} // 	}

	FileEventer interface {
		OnEvent(*PackageProgress)
	}
)
