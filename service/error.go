package service

import "errors"

var (
	ErrWriteDataFail     = errors.New("write data fail")
	ErrWriteDataOverTime = errors.New("write data is overtime")
	ErrNotExistKey       = errors.New("key not exist")
)
