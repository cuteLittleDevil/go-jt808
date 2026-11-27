package gb28181

import (
	"context"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	inCommand "github.com/cuteLittleDevil/go-jt808/gb28181/internal/command"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"github.com/emiago/sipgo/sip"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func (c *Client) message(req *sip.Request, tx sip.ServerTransaction) {
	body := req.Body()
	var confirmType inCommand.ConfirmType
	if err := inCommand.ParseXML(body, &confirmType); err != nil {
		_ = tx.Respond(sip.NewResponseFromRequest(req, http.StatusBadRequest, "Bad Request", nil))
		slog.Warn("parse xml fail",
			slog.String("sim", c.Options.Sim),
			slog.String("id", c.Options.DeviceInfo.ID),
			slog.Any("err", err))
		return
	}

	data, err := c.handleMessages(confirmType, body)
	if err != nil {
		if respErr := tx.Respond(sip.NewResponseFromRequest(req, http.StatusNotFound, "暂不支持的指令", nil)); respErr != nil {
			slog.Warn("message respond fail",
				slog.String("sim", c.Options.Sim),
				slog.String("id", c.Options.DeviceInfo.ID),
				slog.Any("err", respErr))
		}
		slog.Warn("unsupported message",
			slog.String("sim", c.Options.Sim),
			slog.String("id", c.Options.DeviceInfo.ID),
			slog.Any("data", req.String()),
			slog.Any("err", err))
		return
	}

	if err := tx.Respond(sip.NewResponseFromRequest(req, http.StatusOK, "OK", nil)); err != nil {
		slog.Warn("message respond fail",
			slog.String("sim", c.Options.Sim),
			slog.String("id", c.Options.DeviceInfo.ID),
			slog.Any("err", err))
	}

	replyReq, _ := c.newSIPRequest(sip.MESSAGE, req.Transport(), false)
	replyReq.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSCDP+xml"))
	replyReq.SetBody(data)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientTx, err := c.client.TransactionRequest(ctx, replyReq)
	if err != nil {
		slog.Warn("transaction request fail",
			slog.String("sim", c.Options.Sim),
			slog.String("id", c.Options.DeviceInfo.ID),
			slog.Any("err", err))
		return
	}
	defer clientTx.Terminate()
	_, _ = c.getResponse(ctx, clientTx)
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

	defaultInfo := c.defaultInvite(inviteInfo)
	c.manage.SubmitInvite(defaultInfo)

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
		c.manage.SubmitBye(inviteInfo.CallId)
		slog.Warn("invite respond fail",
			slog.String("sim", c.Options.Sim),
			slog.String("id", c.Options.DeviceInfo.ID),
			slog.String("callID", inviteInfo.CallId),
			slog.Any("err", err))
	}
}

func (c *Client) ack(req *sip.Request, _ sip.ServerTransaction) {
	if callID := req.CallID(); callID != nil {
		c.manage.SubmitAck(callID.Value())
	}
}

func (c *Client) bye(req *sip.Request, tx sip.ServerTransaction) {
	if callID := req.CallID(); callID != nil {
		c.manage.SubmitBye(callID.Value())
	}
	if err := tx.Respond(sip.NewResponseFromRequest(req, sip.StatusOK, "OK", nil)); err != nil {
		slog.Warn("bye respond fail",
			slog.String("sim", c.Options.Sim),
			slog.String("id", c.Options.DeviceInfo.ID),
			slog.Any("err", err))
	}
}

func (c *Client) defaultInvite(inviteInfo *command.InviteInfo) *command.InviteInfo {
	inviteInfo.Adapter = struct {
		Enable bool             `json:"enable"`
		Port   int              `json:"port"`
		Type   command.ToGBType `json:"type"`
	}{
		Enable: true,
		Port:   0,
		Type:   c.Options.ToGBType,
	}

	switch inviteInfo.Adapter.Type {
	case command.JT1078ToPS, command.JT1078ToPSFilterPacket:
		// 通道默认为通道ID最后一位数字，兜底为 1 防止空串/非数字越界
		channel := 1
		if id := inviteInfo.TargetChannelId; len(id) > 0 {
			if last := id[len(id)-1]; last >= '0' && last <= '9' {
				channel = int(last - '0')
			}
		}
		inviteInfo.JT1078Info = struct {
			Sim            string                                  `json:"sim"`
			Channel        int                                     `json:"channelId"`
			Port           int                                     `json:"port"`
			StreamTypes    []jt1078.PTType                         `json:"streamTypes"`
			RtpTypeConvert func(ptType jt1078.PTType) (byte, bool) `json:"-"`
		}{
			Sim:         c.Options.Sim,
			Channel:     channel,
			Port:        inviteInfo.Port - 100, // 端口默认是连接的流媒体端口-100
			StreamTypes: []jt1078.PTType{jt1078.PTH264},
		}
		inviteInfo.Adapter.Port = inviteInfo.JT1078Info.Port
	case command.CustomPS:
		inviteInfo.Adapter.Enable = false
	default:
	}
	return inviteInfo
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
