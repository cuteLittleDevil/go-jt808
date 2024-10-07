package consts

type HandleEventType int

const (
	AllEvent           HandleEventType = -1 // 全部事件 全部指令
	AllReadEvent       HandleEventType = -2 // 读事件 指收到终端的指令
	AllWriteEvent      HandleEventType = -3 // 写事件 指下发给终端的指令
	AllNonsupportEvent HandleEventType = -4 // 不支持事件 收到不支持的指令
)
