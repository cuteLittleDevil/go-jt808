package gb28181

import (
	"context"
	inCommand "github.com/cuteLittleDevil/go-jt808/gb28181/internal/command"
	"github.com/emiago/sipgo/sip"
	"time"
)

func (c *Client) handleKeepalive(parentCtx context.Context) error {
	req, sn := c.newSIPRequest(sip.MESSAGE, c.Options.Transport, true)
	req.AppendHeader(c.contact)
	req.AppendHeader(sip.NewHeader("User-Agent", c.Options.UserAgent))
	req.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSCDP+xml"))

	keep := inCommand.NewKeepalive(c.Options.DeviceInfo.ID, sn)
	req.SetBody(inCommand.ToXML(keep))
	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()
	_, err := c.client.Do(ctx, req)
	return err
}
