package service

import "errors"

var (
	ErrWriteDataFail     = errors.New("write data fail")
	ErrWriteDataOverTime = errors.New("write data is overtime")
	ErrNotExistKey       = errors.New("key not exist")
	// ErrActiveQueueFull 终端主动下发队列已满 通常是终端网络阻塞写不出去.
	ErrActiveQueueFull = errors.New("active message queue full")
)

var (
	_errKeyInvalid = errors.New("key invalid")
	_errKeyExist   = errors.New("key exist")
)
