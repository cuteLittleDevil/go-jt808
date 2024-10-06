package consts

type HandleEventType int

const (
	AllEvent           HandleEventType = -1 // 全部事件 全部指令
	AllReadEvent       HandleEventType = -2 // 读事件 指收到终端的指令
	AllWriteEvent      HandleEventType = -3 // 写事件 指下发给终端的指令
	AllNonsupportEvent HandleEventType = -4 // 不支持事件 收到不支持的指令
)

type ProtocolVersionType uint8

const (
	JT808Protocol2011 ProtocolVersionType = 1
	JT808Protocol2013 ProtocolVersionType = 2
	JT808Protocol2019 ProtocolVersionType = 3
)

func (p ProtocolVersionType) String() string {
	switch p {
	case JT808Protocol2011:
		return "JT2011"
	case JT808Protocol2013:
		return "JT2013"
	case JT808Protocol2019:
		return "JT2019"
	default:
	}
	return "未支持的协议"
}
