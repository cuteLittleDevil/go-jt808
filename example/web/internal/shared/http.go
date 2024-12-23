package shared

import "github.com/cuteLittleDevil/go-jt808/shared/consts"

type (
	PlatformHandler interface {
		Encode() []byte
		Protocol() consts.JT808CommandType
		ReplyProtocol() consts.JT808CommandType
	}

	Request[T PlatformHandler] struct {
		Key     string                  `json:"key" binding:"required"`
		Command consts.JT808CommandType `json:"command" binding:"required"`
		Data    T                       `json:"data" binding:"required"`
	}

	Response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data any    `json:"data"`
	}

	Notice struct {
		Command      consts.JT808CommandType `json:"command"`
		TerminalData string                  `json:"terminalData"`
		PlatformData string                  `json:"platformData"`
		Remark       string                  `json:"remark"`
	}
)
