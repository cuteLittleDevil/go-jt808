package model

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x9003 struct {
	BaseHandle
}

func (p *P0x9003) Protocol() consts.JT808CommandType {
	return consts.P9003QueryTerminalAudioVideoProperties
}

func (p *P0x9003) ReplyProtocol() consts.JT808CommandType {
	return consts.T1003UploadAudioVideoAttr
}

func (p *P0x9003) Parse(_ *jt808.JTMessage) error {
	return nil
}

func (p *P0x9003) Encode() []byte {
	return nil
}

func (p *P0x9003) HasReply() bool {
	return false
}

func (p *P0x9003) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:null%x", p.Protocol(), p.Encode()),
		"}",
	}, "\n")
}
