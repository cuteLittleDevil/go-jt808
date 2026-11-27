package gb28181

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	inCommand "github.com/cuteLittleDevil/go-jt808/gb28181/internal/command"
	"github.com/emiago/sipgo/sip"
)

func TestClient_nextSNConcurrent(t *testing.T) {
	c := New("1001")
	const workers = 10
	const perWorker = 50
	seen := sync.Map{}
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				sn := c.nextSN()
				if _, dup := seen.LoadOrStore(sn, struct{}{}); dup {
					t.Errorf("duplicate sn = %d", sn)
				}
			}
		}()
	}
	wg.Wait()
	if got := c.sn.Load(); got != workers*perWorker {
		t.Fatalf("final sn = %d, want %d", got, workers*perWorker)
	}
}

func TestClient_defaultInviteChannelBoundary(t *testing.T) {
	c := New("1001", WithToGBType(command.JT1078ToPS))

	tests := []struct {
		name      string
		channelID string
		wantCh    int
	}{
		{name: "标准通道末位", channelID: "34020000001320000334", wantCh: 4},
		{name: "空通道ID不panic兜底1", channelID: "", wantCh: 1},
		{name: "非数字末位不panic兜底1", channelID: "abcX", wantCh: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := c.defaultInvite(&command.InviteInfo{
				TargetChannelId: tt.channelID,
				Port:            10100,
			})
			if info.JT1078Info.Channel != tt.wantCh {
				t.Fatalf("Channel = %d, want %d", info.JT1078Info.Channel, tt.wantCh)
			}
			if info.JT1078Info.Port != 10000 || info.Adapter.Port != 10000 {
				t.Fatalf("Port = (%d, %d), want 10000", info.JT1078Info.Port, info.Adapter.Port)
			}
		})
	}
}

func TestClient_decodeSDP(t *testing.T) {
	c := New("1001")
	req := sip.NewRequest(sip.INVITE, sip.Uri{User: "34020000001320000330", Host: "127.0.0.1", Port: 5060})
	req.AppendHeader(&sip.FromHeader{
		Address: sip.Uri{User: "34020000002000000001", Host: "127.0.0.1", Port: 15060},
		Params:  sip.HeaderParams{{K: "tag", V: "testtag"}},
	})
	callID := sip.CallIDHeader("call-123@127.0.0.1")
	req.AppendHeader(&callID)
	req.AppendHeader(sip.NewHeader("Subject", "34020000001320000332:0,34020000002000000001:0"))
	sdp := strings.Join([]string{
		"v=0",
		"o=34020000002000000001 0 0 IN IP4 192.168.1.10",
		"s=Play",
		"c=IN IP4 192.168.1.10",
		"t=0 0",
		"m=video 20000 TCP/RTP/AVP 96",
		"a=setup:passive",
		"a=recvonly",
		"y=0100000001",
	}, "\r\n") + "\r\n"
	req.SetBody([]byte(sdp))

	info, err := c.decodeSDP(req)
	if err != nil {
		t.Fatalf("decodeSDP() error = %v", err)
	}
	if info.SessionName != "Play" || info.IP != "192.168.1.10" || info.Port != 20000 || !info.TCP || info.SSRC != "0100000001" {
		t.Fatalf("unexpected InviteInfo = %+v", info)
	}
	if info.TargetChannelId != "34020000001320000332" || info.CallId != "call-123@127.0.0.1" {
		t.Fatalf("unexpected IDs: target=%s callID=%s", info.TargetChannelId, info.CallId)
	}
}

func TestClient_handleMessages(t *testing.T) {
	c := New("1001")

	catalogXML := []byte(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>Catalog</CmdType>
<SN>12</SN>
<DeviceID>34020000001320000330</DeviceID>
</Query>`)
	data, err := c.handleMessages(inCommand.ConfirmType{CmdType: "Catalog"}, catalogXML)
	if err != nil {
		t.Fatalf("Catalog handleMessages() error = %v", err)
	}
	if !strings.Contains(string(data), "<SumNum>4</SumNum>") {
		t.Fatalf("Catalog response missing SumNum=4: %s", string(data))
	}

	if _, err := c.handleMessages(inCommand.ConfirmType{CmdType: "UnknownCmd"}, catalogXML); err == nil {
		t.Fatal("expected error for unknown cmd type")
	}

	infoXML := []byte(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>DeviceInfo</CmdType>
<SN>13</SN>
<DeviceID>34020000001320000330</DeviceID>
</Query>`)
	infoResp, err := c.handleMessages(inCommand.ConfirmType{CmdType: "DeviceInfo"}, infoXML)
	if err != nil || !strings.Contains(string(infoResp), "<DeviceName>1001-jt808-simulation</DeviceName>") {
		t.Fatalf("DeviceInfo handleMessages() = (%q, %v)", string(infoResp), err)
	}

	statusXML := []byte(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>DeviceStatus</CmdType>
<SN>14</SN>
<DeviceID>34020000001320000330</DeviceID>
</Query>`)
	statusResp, err := c.handleMessages(inCommand.ConfirmType{CmdType: "DeviceStatus"}, statusXML)
	if err != nil || !strings.Contains(string(statusResp), "<Online>ONLINE</Online>") {
		t.Fatalf("DeviceStatus handleMessages() = (%q, %v)", string(statusResp), err)
	}
}

func TestClient_decodeSDPEdgeCases(t *testing.T) {
	c := New("1001")

	// 1. Playback 从 u= 提取通道 ID、解析 t= 和 a=setup:active、a=downloadspeed
	req := sip.NewRequest(sip.INVITE, sip.Uri{User: "34020000001320000330", Host: "127.0.0.1", Port: 5060})
	req.AppendHeader(&sip.FromHeader{
		Address: sip.Uri{User: "34020000002000000001", Host: "127.0.0.1", Port: 15060},
		Params:  sip.HeaderParams{{K: "tag", V: "t1"}},
	})
	req.SetBody([]byte(strings.Join([]string{
		"v=0",
		"s=Playback",
		"u=34020000001320000333:0",
		"c=IN IP4 10.0.0.2",
		"t=1700000000 1700003600",
		"m=video 21000 TCP/RTP/AVP 96",
		"a=setup:active",
		"a=downloadspeed:2",
		"y=0200000001",
	}, "\r\n") + "\r\n"))

	info, err := c.decodeSDP(req)
	if err != nil {
		t.Fatalf("decodeSDP(Playback) error = %v", err)
	}
	if info.TargetChannelId != "34020000001320000333" || info.StartTime != 1700000000 || info.StopTime != 1700003600 || !info.TCPActive || info.DownloadSpeed != "2" {
		t.Fatalf("unexpected Playback InviteInfo = %+v", info)
	}

	// 2. 空 SDP 报错
	emptyReq := sip.NewRequest(sip.INVITE, sip.Uri{User: "u", Host: "127.0.0.1", Port: 5060})
	emptyReq.AppendHeader(&sip.FromHeader{Address: sip.Uri{User: "from", Host: "127.0.0.1"}})
	if _, err := c.decodeSDP(emptyReq); err == nil {
		t.Fatal("expected error for empty SDP")
	}

	// 3. 缺少 From 报错
	noFromReq := sip.NewRequest(sip.INVITE, sip.Uri{User: "u", Host: "127.0.0.1", Port: 5060})
	if _, err := c.decodeSDP(noFromReq); err == nil {
		t.Fatal("expected error for missing From")
	}
}

func TestClient_StopWithoutRunAndCancelInFlight(t *testing.T) {
	// 1. 未调用 Run 直接 Stop 不死锁
	c1 := New("1001",
		WithToGBType(command.JT1078ToPSFilterPacket),
		WithKeepAliveSecond(1),
	)
	if err := c1.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		c1.Stop()
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop() blocked when Run() was never started")
	}

	// 2. Run 中途 Stop 能立即取消在途 REGISTER，不卡 30s
	c2 := New("1002",
		WithPlatformInfo(PlatformInfo{
			Domain: "34020000002",
			ID:     "34020000002000000001",
			IP:     "127.0.0.1",
			Port:   1, // 不可达端口，模拟无响应
		}),
		WithDeviceInfo(DeviceInfo{ID: "34020000001320000330", IP: "127.0.0.1", Port: 5060}),
		WithToGBType(command.RelayPS),
	)
	if err := c2.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	go c2.Run()
	time.Sleep(30 * time.Millisecond)

	stopDone := make(chan struct{})
	go func() {
		defer close(stopDone)
		c2.Stop()
	}()
	select {
	case <-stopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() took too long to cancel in-flight Run()")
	}
}

type stubServerTx struct {
	sip.ServerTransaction
	mu   sync.Mutex
	resp *sip.Response
}

func (s *stubServerTx) Respond(res *sip.Response) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resp = res
	return nil
}

func TestClient_byeResponds200OK(t *testing.T) {
	c := New("1001", WithToGBType(command.CustomPS))
	if err := c.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer c.Stop()

	req := sip.NewRequest(sip.BYE, sip.Uri{User: "34020000001320000330", Host: "127.0.0.1", Port: 5060})
	req.AppendHeader(&sip.ViaHeader{ProtocolName: "SIP", ProtocolVersion: "2.0", Transport: "UDP", Host: "127.0.0.1", Port: 15060})
	req.AppendHeader(&sip.FromHeader{Address: sip.Uri{User: "srv", Host: "127.0.0.1"}})
	req.AppendHeader(&sip.ToHeader{Address: sip.Uri{User: "dev", Host: "127.0.0.1"}})
	callID := sip.CallIDHeader("bye-call-1")
	req.AppendHeader(&callID)
	req.AppendHeader(&sip.CSeqHeader{SeqNo: 1, MethodName: sip.BYE})

	tx := &stubServerTx{}
	c.bye(req, tx)
	if tx.resp == nil || tx.resp.StatusCode != sip.StatusOK {
		t.Fatalf("bye() response = %v, want 200 OK", tx.resp)
	}

	// 无 Call-ID 时 ack/bye 均不 panic
	noCallIDReq := sip.NewRequest(sip.ACK, sip.Uri{User: "dev", Host: "127.0.0.1", Port: 5060})
	c.ack(noCallIDReq, tx)
}

func TestClient_buildAuthRegisterRequestCSeqIncrement(t *testing.T) {
	c := New("1001",
		WithPlatformInfo(PlatformInfo{
			Domain:   "3402000000",
			ID:       "34020000002000000001",
			Password: "secret",
			IP:       "127.0.0.1",
			Port:     15060,
		}),
		WithDeviceInfo(DeviceInfo{
			ID:   "34020000001320000330",
			IP:   "127.0.0.1",
			Port: 5060,
		}),
		WithTransport("UDP"),
	)

	req := sip.NewRequest(sip.REGISTER, sip.Uri{User: "34020000002000000001", Host: "127.0.0.1", Port: 15060})
	req.AppendHeader(&sip.ViaHeader{ProtocolName: "SIP", ProtocolVersion: "2.0", Transport: "UDP", Host: "127.0.0.1", Port: 15060})
	req.AppendHeader(&sip.FromHeader{Address: sip.Uri{User: c.DeviceInfo.ID, Host: "127.0.0.1"}})
	req.AppendHeader(&sip.ToHeader{Address: sip.Uri{User: c.PlatformInfo.ID, Host: "127.0.0.1"}})
	callID := sip.CallIDHeader("reg-call-1")
	req.AppendHeader(&callID)
	firstSN := c.nextSN()
	req.AppendHeader(&sip.CSeqHeader{SeqNo: firstSN, MethodName: sip.REGISTER})

	// 1. 缺少 WWW-Authenticate 返回错误
	noAuthRes := sip.NewResponseFromRequest(req, sip.StatusUnauthorized, "Unauthorized", nil)
	if _, err := c.buildAuthRegisterRequest(req, noAuthRes); err == nil {
		t.Fatal("expected error when WWW-Authenticate header is missing")
	}

	// 2. 合法 401 响应：返回携带 Authorization 且递增 CSeq 的新请求
	authRes := sip.NewResponseFromRequest(req, sip.StatusUnauthorized, "Unauthorized", nil)
	authRes.AppendHeader(sip.NewHeader("WWW-Authenticate", `Digest realm="3402000000", nonce="testnonce123", algorithm=MD5`))
	newReq, err := c.buildAuthRegisterRequest(req, authRes)
	if err != nil {
		t.Fatalf("buildAuthRegisterRequest() error = %v", err)
	}
	if newReq.GetHeader("Authorization") == nil {
		t.Fatal("expected Authorization header in retry request")
	}
	if newReq.Via() != nil {
		t.Fatal("expected old Via header to be removed in retry request")
	}
	if gotSN := newReq.CSeq().SeqNo; gotSN <= firstSN {
		t.Fatalf("retry CSeq = %d, want > %d", gotSN, firstSN)
	}
}

func TestClient_ConcurrentInitAndZeroKeepAlive(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := New("1001", WithKeepAliveSecond(0))
			if err := c.Init(); err != nil {
				t.Errorf("Init() error = %v", err)
				return
			}
			go c.Run()
			time.Sleep(10 * time.Millisecond)
			c.Stop()
		}()
	}
	wg.Wait()
}
