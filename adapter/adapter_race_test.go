package adapter

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

func TestAdapter_CreateTerminalsDoesNotShareBackingArray(t *testing.T) {
	// 故意预留多余 cap，验证多连接并发 createTerminals 不会踩同一底层数组
	baseCmds := make([]consts.JT808CommandType, 1, 8)
	baseCmds[0] = consts.P9205QueryResourceList

	ad := New(
		WithAllowCommand(consts.P8001GeneralRespond),
		WithTerminals(Terminal{
			Mode:          Follower,
			TargetAddr:    "127.0.0.1:8080",
			AllowCommands: baseCmds,
		}),
	)

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ts := ad.createTerminals()
			if len(ts) != 1 || len(ts[0].AllowCommands) != 2 {
				t.Errorf("unexpected terminals: %+v", ts)
				return
			}
			ts[0].AllowCommands[0] = consts.T0002HeartBeat
		}()
	}
	wg.Wait()

	if baseCmds[0] != consts.P9205QueryResourceList {
		t.Fatalf("original AllowCommands mutated: %v", baseCmds)
	}
}

func TestAdapter_StopConcurrentWithRunNoRace(t *testing.T) {
	ad := New(WithHostPorts("127.0.0.1:0"))
	runDone := make(chan struct{})
	go func() {
		defer close(runDone)
		ad.Run()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ad.Stop()
		}()
	}
	wg.Wait()

	select {
	case <-runDone:
	case <-time.After(time.Second):
		t.Fatal("Run() did not exit after Stop()")
	}
}

func TestBackendClient_ReadLoopDownlinkCopyNoRace(t *testing.T) {
	// 连续下发两种不同报文，验证异步消费 onReply 时不会因 readBuf 覆写导致数据竞争或串包
	pkt1 := testDownlink8001
	pkt2 := testDownlink9205

	clientEnd, serverEnd := net.Pipe()
	defer func() {
		_ = clientEnd.Close()
		_ = serverEnd.Close()
	}()

	const total = 100
	replies := make(chan []byte, total)
	bc := &backendClient{
		terminal:    Terminal{Mode: Leader, TargetAddr: "pipe"},
		sessionStop: make(chan struct{}),
		onReply: func(data []byte) {
			replies <- data
		},
		conn:       clientEnd,
		writeQueue: make(chan []byte, 10),
		stopCh:     make(chan struct{}),
	}

	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		bc.readLoop()
	}()

	go func() {
		for i := 0; i < total; i++ {
			pkt := pkt1
			if i%2 == 1 {
				pkt = pkt2
			}
			if _, err := serverEnd.Write(pkt); err != nil {
				t.Errorf("write[%d] error = %v", i, err)
				return
			}
		}
		_ = serverEnd.Close()
	}()

	for i := 0; i < total; i++ {
		want := pkt1
		if i%2 == 1 {
			want = pkt2
		}
		select {
		case got := <-replies:
			if !bytes.Equal(got, want) {
				t.Fatalf("reply[%d] = %x, want %x", i, got, want)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for reply[%d]", i)
		}
	}

	select {
	case <-readDone:
	case <-time.After(time.Second):
		t.Fatal("readLoop did not exit")
	}
}
