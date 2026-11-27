package gb28181

import (
	"context"
	inCommand "github.com/cuteLittleDevil/go-jt808/gb28181/internal/command"
	"github.com/emiago/sipgo/sip"
	"time"
)

func (c *Client) handleKeepalive(parentCtx context.Context) error {
	platform := c.Options.PlatformInfo
	req := sip.NewRequest(sip.MESSAGE, c.recipient)
	req.SetTransport(c.Options.Transport)
	req.AppendHeader(&sip.ViaHeader{
		ProtocolName:    "SIP",
		ProtocolVersion: "2.0",
		Transport:       c.Options.Transport,
		Host:            platform.IP,
		Port:            platform.Port,
		Params: sip.HeaderParams{
			{K: "branch", V: sip.GenerateBranchN(16)},
			{K: "rport", V: ""}, // 不知道实际使用的端口
		},
	})

	from := *c.from
	from.Params = sip.HeaderParams{{K: "tag", V: sip.GenerateTagN(16)}}
	req.AppendHeader(&from)
	req.AppendHeader(c.to)
	req.AppendHeader(&c.callID)
	sn := c.nextSN()
	req.AppendHeader(&sip.CSeqHeader{
		SeqNo:      sn,
		MethodName: sip.MESSAGE,
	})
	req.AppendHeader(c.contact)
	maxForwardsHeader := sip.MaxForwardsHeader(70)
	req.AppendHeader(&maxForwardsHeader)
	req.AppendHeader(sip.NewHeader("User-Agent", c.Options.UserAgent))
	req.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSCDP+xml"))

	keep := inCommand.NewKeepalive(c.Options.DeviceInfo.ID, sn)
	req.SetBody(inCommand.ToXML(keep))
	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()
	_, err := c.client.Do(ctx, req)
	return err
}
