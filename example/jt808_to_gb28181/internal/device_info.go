package internal

import "fmt"

type DeviceInfo struct {
	ID string
	// ip和端口只是sip传输过去 实际不会用到
	IP   string
	Port int
}

func DefaultDeviceInfo(sim string) DeviceInfo {
	// 默认的设备id 就是3402000000132 + sim卡号最后6位 + 0
	sim = "000000" + sim
	return DeviceInfo{
		ID:   fmt.Sprintf("3402000000132%s0", sim[len(sim)-6:]),
		IP:   "127.0.0.1",
		Port: 5060,
	}
}
