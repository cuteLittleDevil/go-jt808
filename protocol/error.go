package protocol

import "errors"

var (
	ErrUnqualifiedData         = errors.New("unqualified data")
	ErrHeaderLength2Short      = errors.New("header length too short")
	ErrBodyLengthInconsistency = errors.New("body length inconsistency")
	ErrCheckCode               = errors.New("check code fail")
)
