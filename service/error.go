package service

import "errors"

var (
	ErrWriteDataFail     = errors.New("write data fail")
	ErrWriteDataOverTime = errors.New("write data is overtime")
	ErrNotExistKey       = errors.New("key not exist")
)

var (
	_errKeyInvalid = errors.New("key invalid")
	_errKeyExist   = errors.New("key exist")
)
