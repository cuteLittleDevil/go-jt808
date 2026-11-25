package attachment

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

func buildSuBiaoStreamPacket(name string, offset uint32, payload []byte) []byte {
	buf := make([]byte, 62+len(payload))
	copy(buf[0:4], []byte{0x30, 0x31, 0x63, 0x64})
	copy(buf[4:54], []byte(name))
	binary.BigEndian.PutUint32(buf[54:58], offset)
	binary.BigEndian.PutUint32(buf[58:62], uint32(len(payload)))
	copy(buf[62:], payload)
	return buf
}

func TestSafeSavePath_RejectsTraversalAndSavesSafely(t *testing.T) {
	root := t.TempDir()

	badCases := []struct {
		phone string
		name  string
	}{
		{phone: "12345678901", name: "../../etc/passwd"},
		{phone: "12345678901", name: `..\evil.jpg`},
		{phone: "12345678901", name: "sub/dir.jpg"},
		{phone: "12345678901", name: ".."},
		{phone: "12345678901", name: ""},
		{phone: "../escape", name: "a.jpg"},
	}
	for _, tc := range badCases {
		if _, _, err := safeSavePath(root, tc.phone, tc.name); err == nil {
			t.Errorf("safeSavePath(%q, %q) expected error, got nil", tc.phone, tc.name)
		}
	}

	ev := &fileEvent{saveRoot: root}
	progress := &PackageProgress{
		ProgressStage: ProgressStageSuccessQuit,
		Record: map[string]*Package{
			"ok.jpg":        {FileName: "ok.jpg", StreamBody: []byte("hello")},
			"../../pwn.jpg": {FileName: "../../pwn.jpg", StreamBody: []byte("bad")},
		},
		ExtensionFields: ExtensionFields{
			TerminalPhoneNo:  "12345678901",
			ActiveSafetyType: consts.ActiveSafetyJS,
		},
	}
	ev.OnEvent(progress)

	savedPath := filepath.Join(root, "12345678901", "ok.jpg")
	got, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("expected valid file saved at %s: %v", savedPath, err)
	}
	if string(got) != "hello" {
		t.Fatalf("saved content = %q, want hello", got)
	}
	if _, err := os.Stat(filepath.Join(root, "pwn.jpg")); !os.IsNotExist(err) {
		t.Fatal("path traversal file should not exist")
	}
}

func TestPackageProgress_DuplicateRetransmitCompletesCleanly(t *testing.T) {
	const fileName = "00_64_6401_00_0001.jpg"
	part1 := []byte("0123456789")
	part2 := []byte("abcdefghij")
	full := append(append([]byte(nil), part1...), part2...)

	progress := &PackageProgress{
		ProgressStage: ProgressStageStart,
		Record: map[string]*Package{
			fileName: {
				FileName:         fileName,
				FileSize:         uint32(len(full)),
				OffsetDataRecord: map[int][]byte{},
				OffsetRecord:     map[int]int{},
			},
		},
		handle: newStandardJT808DataHandle(consts.ActiveSafetyJS),
	}

	// 故意连续发送两次 part1（模拟重传），再发 part2，再重发一次 part2
	packets := [][]byte{
		buildSuBiaoStreamPacket(fileName, 0, part1),
		buildSuBiaoStreamPacket(fileName, 0, part1),
		buildSuBiaoStreamPacket(fileName, uint32(len(part1)), part2),
		buildSuBiaoStreamPacket(fileName, uint32(len(part1)), part2),
	}
	for _, pkt := range packets {
		progress.historyData = append(progress.historyData, pkt...)
		for err := range progress.iter() {
			if err != nil {
				t.Fatalf("unexpected iter err = %v", err)
			}
		}
	}

	pack := progress.Record[fileName]
	if pack.CurrentSize != uint32(len(full)) {
		t.Fatalf("CurrentSize = %d, want %d", pack.CurrentSize, len(full))
	}
	if !bytes.Equal(pack.StreamBody, full) {
		t.Fatalf("StreamBody = %q, want %q", pack.StreamBody, full)
	}
	if len(pack.StatisticalMissSegments()) != 0 {
		t.Fatalf("expected 0 miss segments, got %+v", pack.StatisticalMissSegments())
	}
}

func TestPackage_StatisticalMissSegmentsOverlapping(t *testing.T) {
	pack := &Package{
		FileSize:    200,
		CurrentSize: 120,
		OffsetRecord: map[int]int{
			0:  100, // [0, 100)
			20: 30,  // [20, 50) 被 [0, 100) 完全覆盖，不应把 currentOffset 倒退回 50
			150: 50, // [150, 200)
		},
	}
	miss := pack.StatisticalMissSegments()
	if len(miss) != 1 || miss[0].DataOffset != 100 || miss[0].DataLength != 50 {
		t.Fatalf("miss = %+v, want [{DataOffset:100 DataLength:50}]", miss)
	}
}

func TestPackageProgress_IterEarlyBreakNoPanic(t *testing.T) {
	const fileName = "early.jpg"
	progress := &PackageProgress{
		ProgressStage: ProgressStageStart,
		Record: map[string]*Package{
			fileName: {
				FileName:         fileName,
				FileSize:         20,
				OffsetDataRecord: map[int][]byte{},
				OffsetRecord:     map[int]int{},
			},
		},
		handle: newStandardJT808DataHandle(consts.ActiveSafetyJS),
	}
	progress.historyData = append(progress.historyData, buildSuBiaoStreamPacket(fileName, 0, []byte("0123456789"))...)
	progress.historyData = append(progress.historyData, buildSuBiaoStreamPacket(fileName, 10, []byte("abcdefghij"))...)

	count := 0
	for err := range progress.iter() {
		if err != nil {
			t.Fatalf("unexpected err = %v", err)
		}
		count++
		break // 提前退出，不得触发 Go 1.23 yield panic
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

type stageRecorder struct {
	stages []ProgressStage
}

func (s *stageRecorder) OnEvent(pack *PackageProgress) {
	s.stages = append(s.stages, pack.ProgressStage)
}

func TestConnection_ReadTimeoutExitsCleanly(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer func() { _ = clientConn.Close() }()

	rec := &stageRecorder{}
	c := newConnection(serverConn, consts.ActiveSafetyJS, nil, rec, 50*time.Millisecond)

	done := make(chan struct{})
	go func() {
		defer close(done)
		c.run()
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connection did not exit on read timeout")
	}
	if len(rec.stages) == 0 || rec.stages[len(rec.stages)-1] != ProgressStageUnexpectedExit {
		t.Fatalf("stages = %v, want last stage ProgressStageUnexpectedExit", rec.stages)
	}
}
