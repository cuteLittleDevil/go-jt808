package consts

type ProtocolVersionType uint8

const (
	JT808Protocol2011 ProtocolVersionType = 1
	JT808Protocol2013 ProtocolVersionType = 2
	JT808Protocol2019 ProtocolVersionType = 3
)

func (p ProtocolVersionType) String() string {
	switch p {
	case JT808Protocol2011:
		return "JT2011"
	case JT808Protocol2013:
		return "JT2013"
	case JT808Protocol2019:
		return "JT2019"
	default:
	}

	return "未支持的协议"
}
