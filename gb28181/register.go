package gb28181

import (
	"context"
	"fmt"
	"github.com/emiago/sipgo/sip"
	"github.com/icholy/digest"
	"time"
)

func (c *Client) handleRegister(parentCtx context.Context, hasRegister bool) (bool, error) {
	req, _ := c.newSIPRequest(sip.REGISTER, "", false)
	req.AppendHeader(c.contact)
	req.AppendHeader(sip.NewHeader("User-Agent", c.Options.UserAgent))
	expires := "0" // 注销的情况就是 expires=0
	if hasRegister {
		expires = "3600"
	}
	req.AppendHeader(sip.NewHeader("Expires", expires))
	req.AppendHeader(sip.NewHeader("Content-Length", "0"))
	// 注册wvp 有时候10秒不够 换30秒
	ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
	defer cancel()
	tx, err := c.client.TransactionRequest(ctx, req)
	if err != nil {
		return false, err
	}
	defer tx.Terminate()

	res, err := c.getResponse(ctx, tx)
	if err != nil {
		return false, err
	}
	success := false
	if res.StatusCode == sip.StatusOK {
		success = true
	}
	if hasRegister && res.StatusCode == sip.StatusUnauthorized {
		newReq, err := c.buildAuthRegisterRequest(req, res)
		if err != nil {
			return false, err
		}
		tx, err := c.client.TransactionRequest(ctx, newReq)
		if err != nil {
			return false, err
		}
		defer tx.Terminate()
		newRes, err := c.getResponse(ctx, tx)
		if err != nil {
			return false, err
		}
		if newRes.StatusCode == sip.StatusOK {
			success = true
		}
	}
	return success, nil
}

func (c *Client) buildAuthRegisterRequest(req *sip.Request, res *sip.Response) (*sip.Request, error) {
	wwwAuth := res.GetHeader("WWW-Authenticate")
	if wwwAuth == nil {
		return nil, fmt.Errorf("auth not exist")
	}
	chal, err := digest.ParseChallenge(wwwAuth.Value())
	if err != nil {
		return nil, err
	}
	platform := c.Options.PlatformInfo
	opts := digest.Options{
		Method:   req.Method.String(),
		URI:      "sip:" + platform.Domain,
		Username: c.Options.DeviceInfo.ID,
		Password: platform.Password,
		Cnonce:   sip.GenerateTagN(16),
		Count:    1,
	}

	cred, err := digest.Digest(chal, opts)
	if err != nil {
		return nil, fmt.Errorf("calculating digest failed %w", err)
	}

	// 创建新的带认证信息的请求
	newReq := req.Clone()
	newReq.RemoveHeader("Via")
	newReq.RemoveHeader("CSeq")
	newReq.AppendHeader(&sip.CSeqHeader{
		SeqNo:      c.nextSN(),
		MethodName: sip.REGISTER,
	})
	newReq.AppendHeader(sip.NewHeader("Authorization", cred.String()))
	return newReq, nil
}
