package gb28181

import (
	"context"
	"gb28181/internal"
	"github.com/emiago/sipgo/sip"
	"log/slog"
	"net/http"
	"time"
)

func (c *Client) message(req *sip.Request, tx sip.ServerTransaction) {
	if err := tx.Respond(sip.NewResponseFromRequest(req, http.StatusOK, "OK", nil)); err != nil {
		slog.Warn("message respond fail",
			slog.String("sim", c.Options.Sim),
			slog.String("id", c.Options.DeviceInfo.ID),
			slog.Any("err", err))
	}

	body := req.Body()
	var confirmType internal.ConfirmType
	if err := internal.ParseXML(body, &confirmType); err != nil {
		slog.Warn("parse xml fail",
			slog.String("sim", c.Options.Sim),
			slog.String("id", c.Options.DeviceInfo.ID),
			slog.Any("err", err))
		return
	}

	replyReq := sip.NewRequest(sip.MESSAGE, c.recipient)
	replyReq.AppendHeader(&sip.ViaHeader{
		ProtocolName:    "SIP",
		ProtocolVersion: "2.0",
		Transport:       c.Options.Transport,
		Host:            c.PlatformInfo.IP,
		Port:            c.PlatformInfo.Port,
		Params: sip.NewParams().
			Add("branch", sip.GenerateBranchN(16)).
			Add("rport", ""), // 不知道实际使用的端口
	})
	replyReq.AppendHeader(c.from)
	replyReq.AppendHeader(c.to)
	replyReq.SetTransport(req.Transport())
	replyReq.AppendHeader(&c.callID)
	replyReq.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSCDP+xml"))
	maxForwardsHeader := sip.MaxForwardsHeader(70)
	replyReq.AppendHeader(&maxForwardsHeader)
	req.AppendHeader(&sip.CSeqHeader{
		SeqNo:      c.sn,
		MethodName: sip.REGISTER,
	})
	c.sn++

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if data, err := c.handleMessages(confirmType, body); err == nil {
		replyReq.SetBody(data)
		tx, err := c.client.TransactionRequest(ctx, replyReq)
		if err != nil {
			slog.Warn("transaction request fail",
				slog.String("sim", c.Options.Sim),
				slog.String("id", c.Options.DeviceInfo.ID),
				slog.Any("err", err))
			return
		}
		defer tx.Terminate()
	} else {
		if err := tx.Respond(sip.NewResponseFromRequest(req, http.StatusNotFound, "暂不支持的指令", nil)); err != nil {
			slog.Warn("message respond fail",
				slog.String("sim", c.Options.Sim),
				slog.String("id", c.Options.DeviceInfo.ID),
				slog.Any("err", err))
		}
	}
}

func (c *Client) invite(req *sip.Request, tx sip.ServerTransaction) {

}

func (c *Client) ack(req *sip.Request, tx sip.ServerTransaction) {

}

func (c *Client) bye(req *sip.Request, tx sip.ServerTransaction) {

}
