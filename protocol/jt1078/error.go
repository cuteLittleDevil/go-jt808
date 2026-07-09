package jt1078

import "errors"

var (
	ErrUnqualifiedData      = errors.New("unqualified data")
	ErrHeaderLength2Short   = errors.New("header length too short")
	ErrHeaderLengthTooShort = ErrHeaderLength2Short // alias of ErrHeaderLength2Short
	ErrBodyLength2Short     = errors.New("body length too short")
)
