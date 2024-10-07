package model

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type T0x0002 struct {
	BaseHandle
}

func (t *T0x0002) Encode() []byte {
	return nil
}

func (t *T0x0002) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s: nil", consts.T0002HeartBeat),
		"}",
	}, "\n")
}

func (t *T0x0002) Protocol() consts.JT808CommandType {
	return consts.T0002HeartBeat
}

func (t *T0x0002) ReplyProtocol() consts.JT808CommandType {
	return consts.P8001GeneralRespond
}
