package gb28181

import (
	"fmt"
	"github.com/emiago/sipgo/sip"
	"strconv"
	"strings"
)

// InviteInfo 从INVITE消息中解析需要的信息
// 来源于 https://github.com/langhuihui/monibuca
type InviteInfo struct {
	// 请求者ID
	RequesterId string `json:"requesterId"`
	// 目标通道ID
	TargetChannelId string `json:"targetChannelId"`
	// 源通道ID
	SourceChannelId string `json:"sourceChannelId"`
	// 会话名称
	SessionName string `json:"sessionName"`
	// SSRC
	SSRC string `json:"ssrc"`
	// 是否使用TCP
	TCP bool `json:"tcp"`
	// TCP是否为主动模式
	TCPActive bool `json:"tcpActive"`
	// 呼叫ID
	CallId string `json:"callId"`
	// 开始时间
	StartTime int64 `json:"startTime"`
	// 结束时间
	StopTime int64 `json:"stopTime"`
	// 下载速度
	DownloadSpeed string `json:"downloadSpeed"`
	// IP地址
	IP string `json:"ip"`
	// 端口
	Port       int `json:"port"`
	JT1078Info struct {
		Sim     string `json:"sim"`
		Channel int    `json:"channelId"`
		Port    int    `json:"port"`
	}
}

func decodeSDP(req *sip.Request) (*InviteInfo, error) {
	inviteInfo := &InviteInfo{}

	// 获取请求者ID
	from := req.From()
	if from == nil || from.Address.User == "" {
		return nil, fmt.Errorf("无法从请求中获取来源id")
	}
	inviteInfo.RequesterId = from.Address.User

	// 获取目标通道ID
	channelIDArray := getChannelIDFromRequest(req)

	// 获取CallID
	callID := req.CallID()
	if callID != nil {
		inviteInfo.CallId = callID.Value()
	}

	// 解析SDP消息
	sdpStr := string(req.Body())
	if sdpStr == "" {
		return nil, fmt.Errorf("SDP内容为空")
	}

	// 解析SDP各个字段
	lines := strings.Split(sdpStr, "\r\n")
	var channelIdFromSdp string
	var port int = -1
	var mediaTransmissionTCP bool
	var tcpActive *bool
	var supportedMediaFormat bool
	var sessionName string

	for _, line := range lines {
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "s="):
			sessionName = strings.TrimPrefix(line, "s=")
			inviteInfo.SessionName = sessionName

			// 如果是回放，从URI中获取通道ID
			if strings.EqualFold(sessionName, "Playback") {
				for _, l := range lines {
					if strings.HasPrefix(l, "u=") {
						uriField := strings.TrimPrefix(l, "u=")
						parts := strings.Split(uriField, ":")
						if len(parts) > 0 {
							channelIdFromSdp = parts[0]
						}
						break
					}
				}
			}

		case strings.HasPrefix(line, "c="):
			// c=IN IP4 192.168.1.100
			parts := strings.Split(line, " ")
			if len(parts) >= 3 {
				inviteInfo.IP = parts[2]
			}

		case strings.HasPrefix(line, "t="):
			// t=开始时间 结束时间
			parts := strings.Split(strings.TrimPrefix(line, "t="), " ")
			if len(parts) >= 2 {
				startTime, err := strconv.ParseInt(parts[0], 10, 64)
				if err == nil {
					inviteInfo.StartTime = startTime
				}
				stopTime, err := strconv.ParseInt(parts[1], 10, 64)
				if err == nil {
					inviteInfo.StopTime = stopTime
				}
			}

		case strings.HasPrefix(line, "m="):
			mediaDesc := strings.Split(strings.TrimPrefix(line, "m="), " ")
			if len(mediaDesc) >= 4 { // 必须有足够的元素：类型、端口、传输协议和格式
				portVal, err := strconv.Atoi(mediaDesc[1])
				if err == nil {
					port = portVal
				}

				// 检查传输协议
				if strings.EqualFold(mediaDesc[2], "TCP/RTP/AVP") {
					mediaTransmissionTCP = true
				}

				// 检查是否包含支持的媒体格式：96或8
				for i := 3; i < len(mediaDesc); i++ {
					if mediaDesc[i] == "96" || mediaDesc[i] == "8" {
						supportedMediaFormat = true
						break
					}
				}
			}

		case strings.HasPrefix(line, "a=setup:"):
			val := strings.TrimPrefix(line, "a=setup:")
			if strings.EqualFold(val, "active") {
				activeVal := true
				tcpActive = &activeVal
			} else if strings.EqualFold(val, "passive") {
				passiveVal := false
				tcpActive = &passiveVal
			}

		case strings.HasPrefix(line, "y="):
			inviteInfo.SSRC = strings.TrimPrefix(line, "y=")

		case strings.HasPrefix(line, "a=downloadspeed:"):
			inviteInfo.DownloadSpeed = strings.TrimPrefix(line, "a=downloadspeed:")
		}
	}

	// 确定最终的通道ID，优先使用SDP中的通道ID
	var finalChannelId string
	if channelIdFromSdp != "" {
		finalChannelId = channelIdFromSdp
	} else if len(channelIDArray) > 0 {
		finalChannelId = channelIDArray[0]
	}

	// 验证通道ID和请求者ID
	if inviteInfo.RequesterId == "" || finalChannelId == "" {
		return nil, fmt.Errorf("无法从请求中获取通道id或来源id")
	}

	// 设置目标通道ID
	inviteInfo.TargetChannelId = finalChannelId

	// 设置源通道ID（如果有）
	if len(channelIDArray) >= 2 {
		inviteInfo.SourceChannelId = channelIDArray[1]
	}

	// 验证媒体格式支持
	if port == -1 || !supportedMediaFormat {
		return nil, fmt.Errorf("不支持的媒体格式")
	}

	// 设置传输相关信息
	inviteInfo.TCP = mediaTransmissionTCP
	if tcpActive != nil {
		inviteInfo.TCPActive = *tcpActive
	} else {
		inviteInfo.TCPActive = false // 默认值
	}
	inviteInfo.Port = port

	return inviteInfo, nil
}

// getChannelIDFromRequest 从请求中获取通道ID.
func getChannelIDFromRequest(req *sip.Request) []string {
	subjectHeaders := req.GetHeaders("Subject")
	if len(subjectHeaders) == 0 {
		// 如果缺失subject
		return nil
	}

	// 获取第一个Subject头部的值
	subjectStr := subjectHeaders[0].Value()

	result := make([]string, 2)

	if strings.Contains(subjectStr, ",") {
		subjectSplit := strings.Split(subjectStr, ",")
		result[0] = strings.Split(subjectSplit[0], ":")[0]
		if len(subjectSplit) > 1 {
			result[1] = strings.Split(subjectSplit[1], ":")[0]
		}
	} else {
		result[0] = strings.Split(subjectStr, ":")[0]
	}

	return result
}
