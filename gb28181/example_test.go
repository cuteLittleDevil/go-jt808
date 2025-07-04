package gb28181

import (
	"time"
)

func Example() {
	client := New("1001", WithPlatformInfo(PlatformInfo{
		Domain:   "34020000002",
		ID:       "34020000002000000001",
		Password: "123456",
		IP:       "127.0.0.1",
		Port:     15060,
	}), WithDeviceInfo(DeviceInfo{
		ID: "34020000002000000330",
		// 实际不会用到设备的IP和端口 只是sip传输过去
		IP:   "127.0.0.1",
		Port: 5060,
	}),
		WithTransport("UDP"),
		WithKeepAliveSecond(10),
	)
	if err := client.Init(); err != nil {
		panic(err)
	}
	go client.Run()
	time.Sleep(30 * time.Second)
	client.Stop()

	// Output:
}
