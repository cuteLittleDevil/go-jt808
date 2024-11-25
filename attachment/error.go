package attachment

import "errors"

var (
	ErrUnknownCommand      = errors.New("unknown command")
	ErrDataInconsistency   = errors.New("data inconsistency")
	ErrInsufficientDataLen = errors.New("insufficient data len")
)
