package command

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
)

type Register struct {
	model.T0x0100
	*VerifyInfo
}

func (r *Register) OnReadExecutionEvent(_ *service.Message) {}

func (r *Register) OnWriteExecutionEvent(_ service.Message) {}

func (r *Register) ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error) {
	// 不限制 默认鉴权码用手机号
	code := jtMsg.Header.TerminalPhoneNo
	p8100 := &model.P0x8100{
		RespondSerialNumber: jtMsg.Header.SerialNumber,
		Result:              0,
		AuthCode:            code,
	}
	r.VerifyInfo.Code = code
	return p8100.Encode(), nil
}
