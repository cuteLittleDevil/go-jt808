package gb28181

import (
	"context"
	"fmt"
	"github.com/emiago/sipgo/sip"
	"github.com/icholy/digest"
	"time"
)

func (c *Client) handleRegister(hasRegister bool) (bool, error) {
	req := sip.NewRequest(sip.REGISTER, c.recipient)
	platform := c.Options.PlatformInfo
	req.AppendHeader(&sip.ViaHeader{
		ProtocolName:    "SIP",
		ProtocolVersion: "2.0",
		Transport:       c.Options.Transport,
		Host:            platform.IP,
		Port:            platform.Port,
		Params: sip.NewParams().
			Add("branch", sip.GenerateBranchN(16)).
			Add("rport", ""), // 不知道实际使用的端口
	})

	req.AppendHeader(c.from)
	req.AppendHeader(c.to)
	req.AppendHeader(&c.callID)
	req.AppendHeader(&sip.CSeqHeader{
		SeqNo:      c.sn,
		MethodName: sip.REGISTER,
	})
	c.sn++
	req.AppendHeader(c.contact)
	maxForwardsHeader := sip.MaxForwardsHeader(70)
	req.AppendHeader(&maxForwardsHeader)
	req.AppendHeader(sip.NewHeader("User-Agent", c.Options.UserAgent))
	if hasRegister {
		req.AppendHeader(sip.NewHeader("Expires", "3600"))
	} else {
		// 注销的情况就是 expires=0
		req.AppendHeader(sip.NewHeader("Expires", "0"))
	}
	req.AppendHeader(sip.NewHeader("Content-Length", "0"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := c.client.Do(ctx, req)
	if err != nil {
		return false, err
	}

	success := false
	if res.StatusCode == sip.StatusOK {
		success = true
	}
	if hasRegister && res.StatusCode == sip.StatusUnauthorized {
		wwwAuth := res.GetHeader("WWW-Authenticate")
		if wwwAuth == nil {
			return false, fmt.Errorf("auth not exist")
		}
		chal, err := digest.ParseChallenge(wwwAuth.Value())
		if err != nil {
			return false, err
		}
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
			return false, fmt.Errorf("calculating digest failed %v", err)
		}

		// 创建新的带认证信息的请求
		newReq := req.Clone()
		newReq.RemoveHeader("Via")
		newReq.AppendHeader(sip.NewHeader("Authorization", cred.String()))
		newRes, err := c.client.Do(ctx, newReq)
		if err != nil {
			return false, err
		}
		if newRes.StatusCode == sip.StatusOK {
			success = true
		}
	}
	return success, nil
}
