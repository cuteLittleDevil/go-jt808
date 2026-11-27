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

func TestAdapterServer_stopUnblocksBlockedRead(t *testing.T) {
	conv := &stubConverter{}
	srv := newAdapterServer(&command.InviteInfo{
		IP:   "127.0.0.1",
		Port: 20000,
	}, conv)

	srcServer, srcClient := net.Pipe()
	defer func() {
		_ = srcClient.Close()
	}()

	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		srv.readPacket(srcServer)
	}()

	// 等待 readPacket 进入阻塞 conn.Read
	time.Sleep(20 * time.Millisecond)

	// 模拟收到 BYE 调用 stop()，即使对端未关闭连接，readPacket 也必须立刻退出
	srv.stop("收到bye")
	select {
	case <-readDone:
	case <-time.After(time.Second):
		t.Fatal("readPacket remained blocked after stop()")
	}
}

func TestManage_Lifecycle(t *testing.T) {
	var ackCount, byeCount atomic.Int32
	m := NewManage(func(info *command.InviteInfo) *command.InviteInfo {
		if info.CallId == "drop-me" {
			return nil
		}
		info.Adapter.Enable = false
		return info
	}, func() command.ToGB28181er {
		return &manageStubConverter{ackCount: &ackCount, byeCount: &byeCount}
	})
	go m.Run()

	waitFor := func(cond func() bool, msg string) {
		t.Helper()
		deadline := time.Now().Add(time.Second)
		for time.Now().Before(deadline) {
			if cond() {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		t.Fatalf("%s (ack=%d bye=%d)", msg, ackCount.Load(), byeCount.Load())
	}

	// 1. OnInviteEvent 返回 nil 时不 panic
	m.SubmitInvite(&command.InviteInfo{CallId: "drop-me"})
	m.SubmitAck("drop-me")

	// 2. 正常 Invite -> Ack -> 同 CallId 覆盖 -> Bye（按 SIP 顺序等待处理完成）
	m.SubmitInvite(&command.InviteInfo{CallId: "call-1"})
	time.Sleep(10 * time.Millisecond)
	m.SubmitAck("call-1")
	waitFor(func() bool { return ackCount.Load() == 1 }, "first ack not processed")

	m.SubmitInvite(&command.InviteInfo{CallId: "call-1"})
	waitFor(func() bool { return byeCount.Load() == 1 }, "old server not stopped on duplicate invite")

	m.SubmitAck("call-1")
	waitFor(func() bool { return ackCount.Load() == 2 }, "second ack not processed")

	m.SubmitBye("call-1")
	waitFor(func() bool { return byeCount.Load() == 2 }, "bye not processed")

	// 3. Stop 后继续 Submit 不阻塞
	m.Stop()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 20; i++ {
			m.SubmitInvite(&command.InviteInfo{CallId: "after-stop"})
			m.SubmitAck("after-stop")
			m.SubmitBye("after-stop")
		}
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Submit* blocked after Manage.Stop()")
	}
}

type manageStubConverter struct {
	ackCount *atomic.Int32
	byeCount *atomic.Int32
}

func (m *manageStubConverter) OnAck(_ *command.InviteInfo) {
	m.ackCount.Add(1)
}

func (m *manageStubConverter) ConvertToGB28181(data []byte) ([][]byte, error) {
	return [][]byte{data}, nil
}

func (m *manageStubConverter) OnBye(_ string) {
	m.byeCount.Add(1)
}
