package command

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"web/service/conf"
)

type (
	Auth struct {
		model.T0x0102
		*VerifyInfo
	}

	VerifyInfo struct {
		Code   string
		Verify bool
	}
)

func NewAuth(verifyInfo *VerifyInfo) *Auth {
	return &Auth{VerifyInfo: verifyInfo}
}

func NewVerifyInfo() *VerifyInfo {
	return &VerifyInfo{
		Code:   "",
		Verify: conf.GetData().JTConfig.Verify,
	}
}

func (a *Auth) OnReadExecutionEvent(_ *service.Message) {}

func (a *Auth) OnWriteExecutionEvent(_ service.Message) {}

func (a *Auth) ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error) {
	if err := a.Parse(jtMsg); err != nil {
		return nil, err
	}
	result := byte(2)
	if a.AuthCode == a.VerifyInfo.Code {
		result = 0
	}
	if !a.VerifyInfo.Verify { // 不校验验证码
		result = 0
	}
	//fmt.Println("鉴权", a.AuthInfo.Code, result)
	head := jtMsg.Header
	// 通用应答
	p8001 := &model.P0x8001{
		RespondSerialNumber: head.SerialNumber,
		RespondID:           head.ID,
		Result:              result, // 0-成功 1-失败 2-消息有误 3-不支持 4-报警处理确认 默认成功
	}
	return p8001.Encode(), nil
}
