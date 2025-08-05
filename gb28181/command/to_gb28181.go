package command

// ToGBType 适配流的方式 1-JT1078ToPS 2-RTSPToPS 3-RTMPToPS 4-RelayPS.
type ToGBType int

const (
	JT1078ToPS ToGBType = iota + 1
	RTSPToPS
	RTMPToPS
	RelayPS
	CustomPS
)

type ToGB28181er interface {
	OnAck(info *InviteInfo)
	ConvertToGB28181(data []byte) ([][]byte, error)
	OnBye(msg string)
}

func (t ToGBType) String() string {
	switch t {
	case JT1078ToPS:
		return "jt1078 -> ps流"
	case RTSPToPS:
		return "rtsp -> ps流"
	case RTMPToPS:
		return "rtmp -> ps流"
	case RelayPS:
		return "ps -> ps流"
	case CustomPS:
		return "自定义ps流处理"
	default:
		return "unknown"
	}
}
