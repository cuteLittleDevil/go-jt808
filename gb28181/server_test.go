package gb28181

import (
	"strings"
	"sync"
	"testing"

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
}
