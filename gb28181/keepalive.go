package gb28181

import (
	"context"
	inCommand "github.com/cuteLittleDevil/go-jt808/gb28181/internal/command"
	"github.com/emiago/sipgo/sip"
	"time"
)

func (c *Client) handleKeepalive() error {
	platform := c.Options.PlatformInfo
	req := sip.NewRequest("OnMessage", c.recipient)
	req.SetTransport(c.Options.Transport)
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

	c.from.Params = sip.NewParams().Add("tag", sip.GenerateTagN(16))
	req.AppendHeader(c.from)
	req.AppendHeader(c.to)
	req.AppendHeader(&c.callID)
	req.AppendHeader(&sip.CSeqHeader{
		SeqNo:      c.sn,
		MethodName: sip.MESSAGE,
	})
	req.AppendHeader(c.contact)
	maxForwardsHeader := sip.MaxForwardsHeader(70)
	req.AppendHeader(&maxForwardsHeader)
	req.AppendHeader(sip.NewHeader("User-Agent", c.Options.UserAgent))
	req.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSCDP+xml"))

	keep := inCommand.NewKeepalive(c.Options.DeviceInfo.ID, c.sn)
	req.SetBody(inCommand.ToXML(keep))
	c.sn++
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := c.client.Do(ctx, req)
	return err
}
