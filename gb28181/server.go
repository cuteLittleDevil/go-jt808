package gb28181

import (
	"context"
	"fmt"
	inCommand "github.com/cuteLittleDevil/go-jt808/gb28181/internal/command"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"github.com/emiago/sipgo/sip"
	"log/slog"
	"net/http"
	"strings"
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
	var confirmType inCommand.ConfirmType
	if err := inCommand.ParseXML(body, &confirmType); err != nil {
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
		_, _ = c.getResponse(tx)
	} else {
		err := tx.Respond(sip.NewResponseFromRequest(req, http.StatusNotFound, "暂不支持的指令", nil))
		slog.Warn("message respond fail",
			slog.String("sim", c.Options.Sim),
			slog.String("id", c.Options.DeviceInfo.ID),
			slog.Any("data", req.String()),
			slog.Any("err", err))
	}
}

func (c *Client) invite(req *sip.Request, tx sip.ServerTransaction) {
	inviteInfo, err := c.decodeSDP(req)
	if err != nil {
		_ = tx.Respond(sip.NewResponseFromRequest(req, sip.StatusBadRequest, err.Error(), nil))
		return
	}
	if inviteInfo.SessionName != "Play" {
		_ = tx.Respond(sip.NewResponseFromRequest(req, sip.StatusInternalServerError, "目前只支持点播", nil))
		return
	}
	// 通道就是通道ID最后一位
	channel := inviteInfo.TargetChannelId[len(inviteInfo.TargetChannelId)-1] - '0'
	inviteInfo.JT1078Info = struct {
		Sim         string          `json:"sim"`
		Channel     int             `json:"channelId"`
		Port        int             `json:"port"`
		StreamTypes []jt1078.PTType `json:"streamTypes"`
	}{
		Sim:         c.Options.Sim,
		Channel:     int(channel),
		Port:        inviteInfo.Port - 100, // 端口默认是连接的流媒体端口-100
		StreamTypes: []jt1078.PTType{jt1078.PTH264, jt1078.PTG711A},
	}

	c.manage.SubmitInvite(inviteInfo)
	platformIP := c.PlatformInfo.IP
	content := []string{
		"v=0",
		fmt.Sprintf("o=%s 0 0 IN IP4 %s", inviteInfo.TargetChannelId, platformIP),
		"s=Play",
		fmt.Sprintf("c=IN IP4 %s", platformIP),
		"t=0 0",
		fmt.Sprintf("m=video %d TCP/RTP/AVP 96", inviteInfo.Port),
		"a=setup:active", // GB28181 2016 95页 只支持TCP被动 也就是客户端主动发起TCP连接
		"a=connection:new",
		"a=sendonly",
		"a=rtpmap:96 PS/90000",
		fmt.Sprintf("y=%s", inviteInfo.SSRC),
	}
	response := sip.NewResponseFromRequest(req, sip.StatusOK, "OK", nil)
	contentType := sip.ContentTypeHeader("application/sdp")
	response.AppendHeader(&contentType)
	response.SetBody([]byte(strings.Join(content, "\r\n") + "\r\n"))
	if err := tx.Respond(response); err != nil {
		panic(err)
	}
}

func (c *Client) ack(req *sip.Request, _ sip.ServerTransaction) {
	c.manage.SubmitAck(req.CallID().Value())
}

func (c *Client) bye(req *sip.Request, _ sip.ServerTransaction) {
	c.manage.SubmitBye(req.CallID().Value())
}

func (c *Client) handleMessages(confirmType inCommand.ConfirmType, body []byte) ([]byte, error) {
	switch confirmType.CmdType {
	case "DeviceInfo":
		var info inCommand.DeviceInfo
		if err := inCommand.ParseXML(body, &info); err != nil {
			return nil, err
		}
		name := fmt.Sprintf("%s-jt808-simulation", c.Options.Sim)
		req := inCommand.NewDeviceInfoResponse(name, info)
		return inCommand.ToXML(req), nil
	case "DeviceStatus":
		var status inCommand.DeviceStatus
		if err := inCommand.ParseXML(body, &status); err != nil {
			return nil, err
		}
		req := inCommand.NewDeviceStatusResponse(status)
		return inCommand.ToXML(req), nil
	case "Catalog":
		var catalog inCommand.Catalog
		if err := inCommand.ParseXML(body, &catalog); err != nil {
			return nil, err
		}
		req := inCommand.NewCatalogResponse(catalog, 4)
		return inCommand.ToXML(req), nil
	default:
		return nil, fmt.Errorf("unknown cmd type: %s", confirmType.CmdType)
	}
}
