package stream

import (
	"bytes"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
)

type stubConverter struct {
	byeCount atomic.Int32
}

func (s *stubConverter) OnAck(_ *command.InviteInfo) {}

func (s *stubConverter) ConvertToGB28181(data []byte) ([][]byte, error) {
	return [][]byte{data}, nil
}

func (s *stubConverter) OnBye(_ string) {
	s.byeCount.Add(1)
}

func TestAdapterServer_forwardAndCleanCloseNoSpin(t *testing.T) {
	conv := &stubConverter{}
	srv := newAdapterServer(&command.InviteInfo{
		IP:   "127.0.0.1",
		Port: 20000,
	}, conv)

	srcServer, srcClient := net.Pipe()
	dstServer, dstClient := net.Pipe()
	defer func() {
		_ = srcClient.Close()
		_ = dstClient.Close()
	}()

	readDone := make(chan struct{})
	writeDone := make(chan struct{})
	go func() {
		defer close(readDone)
		srv.readPacket(srcServer)
	}()
	go func() {
		defer close(writeDone)
		srv.writeGB28181Packet(dstServer)
	}()

	want := []byte("ps-packet-payload")
	gotCh := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 128)
		_ = dstClient.SetReadDeadline(time.Now().Add(time.Second))
		n, err := dstClient.Read(buf)
		if err != nil {
			t.Errorf("dst read error = %v", err)
			return
		}
		gotCh <- buf[:n]
	}()

	if _, err := srcClient.Write(want); err != nil {
		t.Fatalf("src write error = %v", err)
	}
	if got := <-gotCh; !bytes.Equal(got, want) {
		t.Fatalf("forwarded = %q, want %q", got, want)
	}

	// 关闭源端，验证 readPacket 关闭 rtpChan 后 writeGB28181Packet 立刻退出而非忙等空转
	_ = srcClient.Close()
	for name, ch := range map[string]chan struct{}{"readPacket": readDone, "writeGB28181Packet": writeDone} {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatalf("%s did not exit after source closed", name)
		}
	}
	if got := conv.byeCount.Load(); got != 1 {
		t.Fatalf("OnBye called %d times, want 1", got)
	}
}

func TestAdapterServer_stopConcurrentNoRace(t *testing.T) {
	conv := &stubConverter{}
	srv := newAdapterServer(&command.InviteInfo{
		IP:   "127.0.0.1",
		Port: 20000,
	}, conv)

	// 并发调用 stop，验证 stopOnce、listen.Swap 与 OnBye 无数据竞争
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			srv.stop("bye")
		}()
	}
	wg.Wait()
	if got := conv.byeCount.Load(); got != 1 {
		t.Fatalf("OnBye called %d times, want 1", got)
	}
}
