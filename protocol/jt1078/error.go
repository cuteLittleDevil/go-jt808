package jt1078

import "errors"

var (
	ErrUnqualifiedData    = errors.New("unqualified data")
	ErrHeaderLength2Short = errors.New("header length too short")
	ErrBodyLength2Short   = errors.New("body length too short")
)
