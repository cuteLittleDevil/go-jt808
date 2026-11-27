package gb28181

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"github.com/emiago/sipgo/sip"
	"strconv"
	"strings"
)

func (c *Client) decodeSDP(req *sip.Request) (*command.InviteInfo, error) {
	from := req.From()
	if from == nil || from.Address.User == "" {
		return nil, fmt.Errorf("无法从请求中获取来源id")
	}

	sdpStr := string(req.Body())
	if sdpStr == "" {
		return nil, fmt.Errorf("SDP内容为空")
	}

	inviteInfo := &command.InviteInfo{
		RequesterId: from.Address.User,
		Port:        -1,
	}
	if callID := req.CallID(); callID != nil {
		inviteInfo.CallId = callID.Value()
	}

	var (
		uriChannelID         string
		supportedMediaFormat bool
	)
	for _, line := range strings.Split(sdpStr, "\r\n") {
		switch {
		case strings.HasPrefix(line, "s="):
			inviteInfo.SessionName = line[2:]
		case strings.HasPrefix(line, "u="):
			uriChannelID, _, _ = strings.Cut(line[2:], ":")
		case strings.HasPrefix(line, "c="):
			if parts := strings.Split(line, " "); len(parts) >= 3 {
				inviteInfo.IP = parts[2]
			}
		case strings.HasPrefix(line, "t="):
			if parts := strings.Split(line[2:], " "); len(parts) >= 2 {
				if v, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					inviteInfo.StartTime = v
				}
				if v, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					inviteInfo.StopTime = v
				}
			}
		case strings.HasPrefix(line, "m="):
			if mediaDesc := strings.Split(line[2:], " "); len(mediaDesc) >= 4 {
				if portVal, err := strconv.Atoi(mediaDesc[1]); err == nil {
					inviteInfo.Port = portVal
				}
				inviteInfo.TCP = strings.EqualFold(mediaDesc[2], "TCP/RTP/AVP")
				for _, fmtCode := range mediaDesc[3:] {
					if fmtCode == "96" || fmtCode == "8" {
						supportedMediaFormat = true
						break
					}
				}
			}
		case strings.HasPrefix(line, "a=setup:"):
			inviteInfo.TCPActive = strings.EqualFold(line[len("a=setup:"):], "active")
		case strings.HasPrefix(line, "y="):
			inviteInfo.SSRC = line[2:]
		case strings.HasPrefix(line, "a=downloadspeed:"):
			inviteInfo.DownloadSpeed = line[len("a=downloadspeed:"):]
		}
	}

	channelIDArray := c.getChannelIDFromRequest(req)
	if strings.EqualFold(inviteInfo.SessionName, "Playback") && uriChannelID != "" {
		inviteInfo.TargetChannelId = uriChannelID
	} else if len(channelIDArray) > 0 {
		inviteInfo.TargetChannelId = channelIDArray[0]
	}
	if inviteInfo.TargetChannelId == "" {
		return nil, fmt.Errorf("无法从请求中获取通道id或来源id")
	}
	if len(channelIDArray) >= 2 {
		inviteInfo.SourceChannelId = channelIDArray[1]
	}
	if inviteInfo.Port == -1 || !supportedMediaFormat {
		return nil, fmt.Errorf("不支持的媒体格式")
	}
	return inviteInfo, nil
}

// getChannelIDFromRequest 从请求中获取通道ID.
func (c *Client) getChannelIDFromRequest(req *sip.Request) []string {
	subjectHeaders := req.GetHeaders("Subject")
	if len(subjectHeaders) == 0 {
		return nil
	}
	parts := strings.Split(subjectHeaders[0].Value(), ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		id, _, _ := strings.Cut(p, ":")
		result = append(result, id)
	}
	return result
}
