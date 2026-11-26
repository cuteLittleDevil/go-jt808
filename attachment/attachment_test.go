package attachment

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
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

func TestHeiBiaoStreamDataHandle_ParseDataSlice(t *testing.T) {
	name := "hlj.jpg"
	payload := []byte("hello-hlj")
	buf := make([]byte, 0, 4+1+len(name)+4+4+len(payload))
	buf = append(buf, 0x30, 0x31, 0x63, 0x64)
	buf = append(buf, byte(len(name)))
	buf = append(buf, []byte(name)...)
	buf = binary.BigEndian.AppendUint32(buf, 128)
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(payload)))
	buf = append(buf, payload...)

	h := newHeiBiaoStreamDataHandle()
	if !h.HasStreamData(buf) || !h.HasMinHeadLen(buf) {
		t.Fatal("expected valid heiBiao stream header")
	}
	headLen, bodyLen := h.Parse(buf)
	if headLen != 4+1+len(name)+4+4 || bodyLen != len(payload) {
		t.Fatalf("headLen=%d bodyLen=%d, want %d %d", headLen, bodyLen, 4+1+len(name)+4+4, len(payload))
	}
	if !bytes.Equal(h.Data, payload) {
		t.Fatalf("h.Data = %x, want %x", h.Data, payload)
	}
}

func TestStandardJT808DataHandle_SupplementarySetsCurrentPackageAndClearsStale(t *testing.T) {
	handle := newStandardJT808DataHandle(consts.ActiveSafetyJS)
	progress := &PackageProgress{
		ProgressStage: ProgressStageInit,
		Record: map[string]*Package{
			"a.jpg": {
				FileName:         "a.jpg",
				FileSize:         100,
				OffsetDataRecord: map[int][]byte{},
				OffsetRecord:     map[int]int{},
			},
		},
		handle: handle,
	}

	make1212 := func(fileName string, fileSize uint32) *jt808.JTMessage {
		body := (&model.T0x1211{
			FileNameLen: byte(len(fileName)),
			FileName:    fileName,
			FileType:    0,
			FileSize:    fileSize,
		}).Encode()
		return &jt808.JTMessage{
			Header: &jt808.Header{
				ID:              uint16(consts.T1212FileUploadComplete),
				TerminalPhoneNo: "12345678901",
			},
			Body: body,
		}
	}

	// 未收到任何流数据时收到 0x1212，进入补传阶段，CurrentPackage 必须非 nil 且 fileEvent 不 panic
	if err := handle.Parse(make1212("a.jpg", 100)); err != nil {
		t.Fatalf("Parse 0x1212 failed: %v", err)
	}
	handle.OnPackageProgressEvent(progress)
	if progress.ProgressStage != ProgressStageSupplementary {
		t.Fatalf("stage = %v, want ProgressStageSupplementary", progress.ProgressStage)
	}
	if progress.ExtensionFields.CurrentPackage != progress.Record["a.jpg"] {
		t.Fatalf("CurrentPackage = %v, want a.jpg package", progress.ExtensionFields.CurrentPackage)
	}
	if len(handle.T0x1212.P0x9212RetransmitPacketList) != 1 {
		t.Fatalf("retransmit list len = %d, want 1", len(handle.T0x1212.P0x9212RetransmitPacketList))
	}
	newFileEvent().OnEvent(progress)

	// 后续收到不存在文件名的 0x1212，不能残留上一次的补传列表
	if err := handle.Parse(make1212("unknown.jpg", 50)); err != nil {
		t.Fatalf("Parse unknown 0x1212 failed: %v", err)
	}
	handle.OnPackageProgressEvent(progress)
	if len(handle.T0x1212.P0x9212RetransmitPacketList) != 0 {
		t.Fatalf("retransmit list len = %d, want 0", len(handle.T0x1212.P0x9212RetransmitPacketList))
	}
}

func TestConnection_IncompleteDisconnectIsFailQuit(t *testing.T) {
	rawHex, err := os.ReadFile("./testdata/su_biao/7e1210.log")
	if err != nil {
		t.Fatalf("read 7e1210.log: %v", err)
	}
	pkt1210, err := hex.DecodeString(string(rawHex))
	if err != nil {
		t.Fatalf("decode 7e1210.log: %v", err)
	}

	serverConn, clientConn := net.Pipe()
	rec := &stageRecorder{}
	c := newConnection(serverConn, consts.ActiveSafetyJS, nil, rec, time.Second)

	done := make(chan struct{})
	go func() {
		defer close(done)
		c.run()
	}()

	// 只发 0x1210，读完平台应答后直接断开，不传任何文件流数据
	if _, err := clientConn.Write(pkt1210); err != nil {
		t.Fatalf("write 0x1210: %v", err)
	}
	buf := make([]byte, 256)
	if _, err := clientConn.Read(buf); err != nil {
		t.Fatalf("read 0x1210 reply: %v", err)
	}
	_ = clientConn.Close()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connection did not exit after client close")
	}
	if len(rec.stages) == 0 || rec.stages[len(rec.stages)-1] != ProgressStageFailQuit {
		t.Fatalf("stages = %v, want last stage ProgressStageFailQuit", rec.stages)
	}
}

func TestPackageProgress_Non7eGarbageReturnsError(t *testing.T) {
	progress := &PackageProgress{
		ProgressStage: ProgressStageInit,
		Record:        map[string]*Package{},
		historyData:   []byte("0123456789abcdef"),
		handle:        newStandardJT808DataHandle(consts.ActiveSafetyJS),
	}
	var gotErr error
	for err := range progress.iter() {
		gotErr = err
	}
	if !errors.Is(gotErr, ErrDataInconsistency) {
		t.Fatalf("iter err = %v, want ErrDataInconsistency", gotErr)
	}
}

type errReplyDataHandler struct {
	DataHandler
	replyErr error
}

func (e *errReplyDataHandler) ReplyData() ([]byte, error) {
	return nil, e.replyErr
}

func TestConnection_ReplyErrorStopsAndPreservesErr(t *testing.T) {
	rawHex, err := os.ReadFile("./testdata/su_biao/7e1210.log")
	if err != nil {
		t.Fatalf("read 7e1210.log: %v", err)
	}
	pkt1210, err := hex.DecodeString(string(rawHex))
	if err != nil {
		t.Fatalf("decode 7e1210.log: %v", err)
	}

	serverConn, clientConn := net.Pipe()
	defer func() {
		_ = serverConn.Close()
		_ = clientConn.Close()
	}()

	wantErr := errors.New("reply boom")
	c := newConnection(serverConn, consts.ActiveSafetyJS, func() DataHandler {
		return &errReplyDataHandler{
			DataHandler: newStandardJT808DataHandle(consts.ActiveSafetyJS),
			replyErr:    wantErr,
		}
	}, &stageRecorder{}, time.Second)

	progress := &PackageProgress{
		ProgressStage: ProgressStageInit,
		Record:        map[string]*Package{},
		handle:        c.dataHandleFunc(),
	}
	if ok := c.handleReadData(progress, pkt1210); ok {
		t.Fatal("handleReadData expected false on ReplyData error")
	}
	if !errors.Is(progress.ExtensionFields.Err, wantErr) {
		t.Fatalf("Err = %v, want %v", progress.ExtensionFields.Err, wantErr)
	}
}

func TestOptions_NilFileEventerAndResetReadTimeout(t *testing.T) {
	opts := newOptions([]Option{
		WithFileEventerFunc(nil),
		WithReadTimeout(5 * time.Second),
		WithReadTimeout(0),
	})
	if opts.FileEventerFunc == nil || opts.FileEventerFunc() == nil {
		t.Fatal("expected non-nil default FileEventerFunc when WithFileEventerFunc(nil) is passed")
	}
	if opts.ReadTimeout != 0 {
		t.Fatalf("ReadTimeout = %v, want 0", opts.ReadTimeout)
	}
}

func buildTerminalJT808Frame(phone string, seq uint16, cmd consts.JT808CommandType, body []byte) []byte {
	bcdPhone, _ := hex.DecodeString(phone)
	raw := make([]byte, 0, 12+len(body)+1)
	raw = binary.BigEndian.AppendUint16(raw, uint16(cmd))
	raw = binary.BigEndian.AppendUint16(raw, uint16(len(body)))
	raw = append(raw, bcdPhone...)
	raw = binary.BigEndian.AppendUint16(raw, seq)
	raw = append(raw, body...)
	raw = append(raw, utils.CreateVerifyCode(raw))

	escaped := make([]byte, 0, len(raw)+4)
	escaped = append(escaped, 0x7e)
	for _, b := range raw {
		switch b {
		case 0x7e:
			escaped = append(escaped, 0x7d, 0x02)
		case 0x7d:
			escaped = append(escaped, 0x7d, 0x01)
		default:
			escaped = append(escaped, b)
		}
	}
	escaped = append(escaped, 0x7e)
	return escaped
}

func buildStreamPacket(asType consts.ActiveSafetyType, name string, offset uint32, payload []byte) []byte {
	if asType == consts.ActiveSafetyHLJ {
		buf := make([]byte, 0, 4+1+len(name)+4+4+len(payload))
		buf = append(buf, 0x30, 0x31, 0x63, 0x64)
		buf = append(buf, byte(len(name)))
		buf = append(buf, []byte(name)...)
		buf = binary.BigEndian.AppendUint32(buf, offset)
		buf = binary.BigEndian.AppendUint32(buf, uint32(len(payload)))
		buf = append(buf, payload...)
		return buf
	}
	buf := make([]byte, 0, 62+len(payload))
	buf = append(buf, 0x30, 0x31, 0x63, 0x64)
	buf = append(buf, utils.String2FillingBytes(name, 50)...)
	buf = binary.BigEndian.AppendUint32(buf, offset)
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(payload)))
	buf = append(buf, payload...)
	return buf
}

func readJT808Reply(t *testing.T, conn net.Conn) *jt808.JTMessage {
	t.Helper()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read reply failed: %v", err)
	}
	msg := jt808.NewJTMessage()
	if err := msg.Decode(buf[:n]); err != nil {
		t.Fatalf("decode reply failed: %v (raw=%x)", err, buf[:n])
	}
	return msg
}

func TestEndToEnd_UploadAndRetransmit_SuBiaoAndHeiBiao(t *testing.T) {
	tests := []struct {
		name   string
		asType consts.ActiveSafetyType
	}{
		{name: "苏标多文件多段补传", asType: consts.ActiveSafetyJS},
		{name: "黑标多文件多段补传", asType: consts.ActiveSafetyHLJ},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type fileSpec struct {
				name      string
				chunks    [][]byte
				missRound [][]int // 每轮故意漏发/等待补传的分片下标
			}
			files := []fileSpec{
				{
					// 缺开头、中间、末尾，分两轮补传完成
					name: "alarm_01.jpg",
					chunks: [][]byte{
						[]byte("chunk-0000"),
						[]byte("chunk-1111"),
						[]byte("chunk-2222"),
						[]byte("chunk-3333"),
						[]byte("chunk-4444"),
					},
					missRound: [][]int{{0, 2, 4}, {2, 4}},
				},
				{
					// 正常顺序传输，无缺包
					name: "alarm_02.txt",
					chunks: [][]byte{
						[]byte("hello "),
						[]byte("world!"),
					},
					missRound: nil,
				},
			}

			phone := "000000001001"
			var items []model.T0x1210AlarmItem
			expected := make(map[string][]byte, len(files))
			for _, f := range files {
				var full []byte
				for _, ch := range f.chunks {
					full = append(full, ch...)
				}
				expected[f.name] = full
				items = append(items, model.T0x1210AlarmItem{
					FileNameLen: byte(len(f.name)),
					FileName:    f.name,
					FileSize:    uint32(len(full)),
				})
			}

			var (
				lastStage  ProgressStage
				savedFiles = map[string][]byte{}
			)
			ev := &funcFileEvent{onEvent: func(p *PackageProgress) {
				lastStage = p.ProgressStage
				if p.ProgressStage == ProgressStageSuccessQuit {
					for name, pack := range p.Record {
						savedFiles[name] = bytes.Clone(pack.StreamBody)
					}
				}
			}}

			serverConn, clientConn := net.Pipe()
			c := newConnection(serverConn, tt.asType, nil, ev, 2*time.Second)
			done := make(chan struct{})
			go func() {
				defer close(done)
				c.run()
			}()

			seq := uint16(1)
			// 1. 发送 0x1210 报警附件信息
			t1210 := &model.T0x1210{
				TerminalID: "1234cd.",
				P9208AlarmSign: model.P9208AlarmSign{
					TerminalID:       "1234cd.",
					Time:             "2025-02-19 20:41:00",
					SerialNumber:     1,
					AttachNumber:     byte(len(items)),
					ActiveSafetyType: tt.asType,
				},
				AlarmID:              "2025-02-19_alarm_01",
				InfoType:             0,
				AttachCount:          byte(len(items)),
				T0x1210AlarmItemList: items,
			}
			if _, err := clientConn.Write(buildTerminalJT808Frame(phone, seq, consts.T1210AlarmAttachInfoMessage, t1210.Encode())); err != nil {
				t.Fatalf("write 0x1210: %v", err)
			}
			seq++
			reply := readJT808Reply(t, clientConn)
			if consts.JT808CommandType(reply.Header.ID) != consts.P8001GeneralRespond {
				t.Fatalf("0x1210 reply id = %x, want 0x8001", reply.Header.ID)
			}

			// 2. 逐文件执行 0x1211 -> 流数据（含丢包/拆包/粘包）-> 0x1212 -> 0x9212 校验 -> 补传
			for _, f := range files {
				t1211 := &model.T0x1211{
					FileNameLen: byte(len(f.name)),
					FileName:    f.name,
					FileType:    0x04,
					FileSize:    uint32(len(expected[f.name])),
				}
				if _, err := clientConn.Write(buildTerminalJT808Frame(phone, seq, consts.T1211FileInfoUpload, t1211.Encode())); err != nil {
					t.Fatalf("write 0x1211: %v", err)
				}
				seq++
				reply = readJT808Reply(t, clientConn)
				if consts.JT808CommandType(reply.Header.ID) != consts.P8001GeneralRespond {
					t.Fatalf("0x1211 reply id = %x, want 0x8001", reply.Header.ID)
				}

				// 计算每个分片的偏移
				offsets := make([]uint32, len(f.chunks))
				var off uint32
				for i, ch := range f.chunks {
					offsets[i] = off
					off += uint32(len(ch))
				}

				skipSet := map[int]bool{}
				if len(f.missRound) > 0 {
					for _, idx := range f.missRound[0] {
						skipSet[idx] = true
					}
				}
				// 发送首轮分片（对第一片做半包拆分，其余正常发送）
				for i, ch := range f.chunks {
					if skipSet[i] {
						continue
					}
					pkt := buildStreamPacket(tt.asType, f.name, offsets[i], ch)
					if i == 1 {
						// 模拟 TCP 半包拆包（先发 4 字节帧头，再发一半包头，再发剩余）
						_, _ = clientConn.Write(pkt[:4])
						_, _ = clientConn.Write(pkt[4:10])
						_, _ = clientConn.Write(pkt[10:])
					} else {
						_, _ = clientConn.Write(pkt)
					}
				}

				t1212 := &model.T0x1212{T0x1211: *t1211}
				send1212AndRead9212 := func() model.P0x9212 {
					if _, err := clientConn.Write(buildTerminalJT808Frame(phone, seq, consts.T1212FileUploadComplete, t1212.Encode())); err != nil {
						t.Fatalf("write 0x1212: %v", err)
					}
					seq++
					r := readJT808Reply(t, clientConn)
					if consts.JT808CommandType(r.Header.ID) != consts.P9212FileUploadCompleteRespond {
						t.Fatalf("0x1212 reply id = %x, want 0x9212", r.Header.ID)
					}
					var p9212 model.P0x9212
					if err := p9212.Parse(r); err != nil {
						t.Fatalf("parse 0x9212 failed: %v", err)
					}
					return p9212
				}

				if len(f.missRound) > 0 {
					// 第 1 次 0x1212：应返回 3 段缺包 [0, 2, 4]
					p9212 := send1212AndRead9212()
					if p9212.UploadResult != 1 || int(p9212.RetransmitPacketNumber) != len(f.missRound[0]) {
						t.Fatalf("round 1 p9212 = %+v, want result=1 count=%d", p9212, len(f.missRound[0]))
					}
					for k, missIdx := range f.missRound[0] {
						got := p9212.P0x9212RetransmitPacketList[k]
						if got.DataOffset != offsets[missIdx] || got.DataLength != uint32(len(f.chunks[missIdx])) {
							t.Fatalf("round 1 miss[%d] = %+v, want offset=%d len=%d",
								k, got, offsets[missIdx], len(f.chunks[missIdx]))
						}
					}

					// 仅补传第 0 片（其余 2、4 仍缺失）
					_, _ = clientConn.Write(buildStreamPacket(tt.asType, f.name, offsets[0], f.chunks[0]))

					// 第 2 次 0x1212：应只剩 2 段缺包 [2, 4]
					p9212 = send1212AndRead9212()
					if p9212.UploadResult != 1 || int(p9212.RetransmitPacketNumber) != len(f.missRound[1]) {
						t.Fatalf("round 2 p9212 = %+v, want result=1 count=%d", p9212, len(f.missRound[1]))
					}

					// 粘包一次性补传第 2、4 片，并重复再发一次第 4 片
					var sticky []byte
					sticky = append(sticky, buildStreamPacket(tt.asType, f.name, offsets[2], f.chunks[2])...)
					sticky = append(sticky, buildStreamPacket(tt.asType, f.name, offsets[4], f.chunks[4])...)
					sticky = append(sticky, buildStreamPacket(tt.asType, f.name, offsets[4], f.chunks[4])...)
					_, _ = clientConn.Write(sticky)
				}

				// 最终 0x1212：应返回 UploadResult=0，无需补传
				p9212 := send1212AndRead9212()
				if p9212.UploadResult != 0 || p9212.RetransmitPacketNumber != 0 {
					t.Fatalf("final p9212 = %+v, want result=0 count=0", p9212)
				}
			}

			_ = clientConn.Close()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("connection did not exit")
			}

			if lastStage != ProgressStageSuccessQuit {
				t.Fatalf("lastStage = %v, want ProgressStageSuccessQuit", lastStage)
			}
			for name, want := range expected {
				if !bytes.Equal(savedFiles[name], want) {
					t.Fatalf("file %s = %q, want %q", name, savedFiles[name], want)
				}
			}
		})
	}
}

type funcFileEvent struct {
	onEvent func(pack *PackageProgress)
}

func (f *funcFileEvent) OnEvent(pack *PackageProgress) {
	if f.onEvent != nil {
		f.onEvent(pack)
	}
}

type badStreamDataHandler struct {
	baseStreamDataHandle
	headLen int
	bodyLen int
}

func (b *badStreamDataHandler) HasStreamData(_ []byte) bool { return true }
func (b *badStreamDataHandler) HasMinHeadLen(_ []byte) bool { return true }
func (b *badStreamDataHandler) Parse(_ []byte) (int, int)   { return b.headLen, b.bodyLen }

type customStreamOwnerHandler struct {
	DataHandler
	stream StreamDataHandler
}

func (c *customStreamOwnerHandler) CreateStreamDataHandler() StreamDataHandler {
	return c.stream
}

type errReplyBodyJT808Handler struct {
	replyErr error
}

func (e *errReplyBodyJT808Handler) Parse(_ *jt808.JTMessage) error { return nil }
func (e *errReplyBodyJT808Handler) ReplyBody(_ *jt808.JTMessage) ([]byte, error) {
	return nil, e.replyErr
}
func (e *errReplyBodyJT808Handler) ReplyProtocol() consts.JT808CommandType {
	return consts.P8001GeneralRespond
}

func TestEdgeCases_ParserAndHandlers(t *testing.T) {
	t.Run("recordStreamChunk same offset larger chunk", func(t *testing.T) {
		pack := &Package{
			FileName:         "grow.bin",
			FileSize:         10,
			OffsetDataRecord: map[int][]byte{},
			OffsetRecord:     map[int]int{},
		}
		recordStreamChunk(pack, 0, 4, 4, []byte("abcd"))
		recordStreamChunk(pack, 0, 10, 10, []byte("abcdefghij"))
		if pack.CurrentSize != 10 || string(pack.OffsetDataRecord[0]) != "abcdefghij" {
			t.Fatalf("pack = %+v, want CurrentSize=10 data=abcdefghij", pack)
		}
	})

	t.Run("stream header too short and negative lengths", func(t *testing.T) {
		p := &PackageProgress{
			Record:      map[string]*Package{},
			historyData: []byte{0x30, 0x31, 0x63, 0x64, 0x01}, // < 62 bytes
			handle:      newStandardJT808DataHandle(consts.ActiveSafetyJS),
		}
		if err := p.stageStreamData(); !errors.Is(err, ErrInsufficientDataLen) {
			t.Fatalf("err = %v, want ErrInsufficientDataLen", err)
		}

		// 流头完整（62字节），但流体未收齐
		partial := buildSuBiaoStreamPacket("partial.jpg", 0, []byte("0123456789"))
		p.historyData = partial[:len(partial)-3]
		if err := p.stageStreamData(); !errors.Is(err, ErrInsufficientDataLen) {
			t.Fatalf("partial body err = %v, want ErrInsufficientDataLen", err)
		}

		// 黑标长度不足 5 字节
		h := newHeiBiaoStreamDataHandle()
		if h.HasMinHeadLen([]byte{0x30, 0x31, 0x63, 0x64}) {
			t.Fatal("expected HasMinHeadLen false for 4 bytes")
		}

		// headLen < 0 / bodyLen < 0
		pNeg := &PackageProgress{
			Record:      map[string]*Package{},
			historyData: []byte{0x30, 0x31, 0x63, 0x64},
			handle: &customStreamOwnerHandler{
				DataHandler: newStandardJT808DataHandle(consts.ActiveSafetyJS),
				stream:      &badStreamDataHandler{headLen: -1, bodyLen: 10},
			},
		}
		if err := pNeg.stageStreamData(); !errors.Is(err, ErrDataInconsistency) {
			t.Fatalf("err = %v, want ErrDataInconsistency", err)
		}
	})

	t.Run("parseJT808Message missing end 0x7e and bad decode", func(t *testing.T) {
		p := &PackageProgress{
			Record:      map[string]*Package{},
			historyData: append([]byte{0x7e}, bytes.Repeat([]byte{0x01}, 15)...),
			handle:      newStandardJT808DataHandle(consts.ActiveSafetyJS),
		}
		if _, err := p.parseJT808Message(); !errors.Is(err, ErrInsufficientDataLen) {
			t.Fatalf("err = %v, want ErrInsufficientDataLen", err)
		}

		// 合法起止 0x7e 但校验码错误
		p.historyData = []byte{0x7e, 0x12, 0x10, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x10, 0x01, 0x00, 0x01, 0xff, 0x7e}
		if _, err := p.parseJT808Message(); err == nil {
			t.Fatal("expected decode error for invalid checksum")
		}
	})

	t.Run("BaseJT808DataHandler unknown command, ReplyBody err, CreateStreamDataHandler", func(t *testing.T) {
		wantErr := errors.New("reply body fail")
		base := &BaseJT808DataHandler[*errReplyBodyJT808Handler, *errReplyBodyJT808Handler, *errReplyBodyJT808Handler]{
			T0x1210: &errReplyBodyJT808Handler{replyErr: wantErr},
			T0x1211: &errReplyBodyJT808Handler{replyErr: wantErr},
			T0x1212: &errReplyBodyJT808Handler{replyErr: wantErr},
		}
		if base.CreateStreamDataHandler() == nil {
			t.Fatal("expected non-nil stream handler")
		}
		unknownMsg := &jt808.JTMessage{
			Header: &jt808.Header{ID: uint16(consts.T0200LocationReport)},
		}
		if err := base.Parse(unknownMsg); !errors.Is(err, ErrUnknownCommand) {
			t.Fatalf("Parse unknown err = %v, want ErrUnknownCommand", err)
		}
		if _, err := base.ReplyData(); !errors.Is(err, ErrUnknownCommand) {
			t.Fatalf("ReplyData unknown err = %v, want ErrUnknownCommand", err)
		}

		// 合法命令但 ReplyBody 返回错误
		validMsg := &jt808.JTMessage{
			Header: &jt808.Header{
				ID:              uint16(consts.T1210AlarmAttachInfoMessage),
				TerminalPhoneNo: "12345678901",
				Property:        &jt808.BodyProperty{},
			},
		}
		if err := base.Parse(validMsg); err != nil {
			t.Fatalf("Parse valid err = %v", err)
		}
		if _, err := base.ReplyData(); !errors.Is(err, wantErr) {
			t.Fatalf("ReplyData err = %v, want %v", err, wantErr)
		}

		// stageJT808Data 遇到未知指令返回错误
		frame := buildTerminalJT808Frame("000000001001", 1, consts.T0200LocationReport, []byte{0x00})
		p := &PackageProgress{
			Record:      map[string]*Package{},
			historyData: frame,
			handle:      newStandardJT808DataHandle(consts.ActiveSafetyJS),
		}
		if err := p.stageJT808Data(); !errors.Is(err, ErrUnknownCommand) {
			t.Fatalf("stageJT808Data err = %v, want ErrUnknownCommand", err)
		}
	})

	t.Run("connection write error, bad data in run, nil fileEventer fallback", func(t *testing.T) {
		serverConn, clientConn := net.Pipe()
		// 测试 nil fileEventer 回退不 panic
		c := newConnection(serverConn, consts.ActiveSafetyJS, nil, nil, time.Second)
		if c.fileEventer == nil {
			t.Fatal("expected default fileEventer")
		}

		// 关闭 clientConn 让 c.conn.Write 失败
		_ = clientConn.Close()
		rawHex, _ := os.ReadFile("./testdata/su_biao/7e1210.log")
		pkt1210, _ := hex.DecodeString(string(rawHex))
		progress := &PackageProgress{
			ProgressStage: ProgressStageInit,
			Record:        map[string]*Package{},
			handle:        c.dataHandleFunc(),
		}
		if ok := c.handleReadData(progress, pkt1210); ok {
			t.Fatal("expected handleReadData false when Write fails")
		}
		if progress.ExtensionFields.Err == nil {
			t.Fatal("expected non-nil write error")
		}
		_ = serverConn.Close()

		// 测试 run() 中收到非 0x7e 脏数据触发 handleReadData 返回 false 并进入 UnexpectedExit
		s2, c2 := net.Pipe()
		rec := &stageRecorder{}
		conn2 := newConnection(s2, consts.ActiveSafetyJS, nil, rec, time.Second)
		done := make(chan struct{})
		go func() {
			defer close(done)
			conn2.run()
		}()
		_, _ = c2.Write([]byte("invalid-garbage-bytes"))
		<-done
		_ = c2.Close()
		if len(rec.stages) == 0 || rec.stages[len(rec.stages)-1] != ProgressStageUnexpectedExit {
			t.Fatalf("stages = %v, want UnexpectedExit", rec.stages)
		}
	})

	t.Run("fileEvent FailQuit, UnexpectedExit, empty saveRoot and ProgressStage.String", func(t *testing.T) {
		tmp := t.TempDir()
		t.Chdir(tmp)

		ev := &fileEvent{saveRoot: ""}
		ev.OnEvent(&PackageProgress{
			ProgressStage: ProgressStageFailQuit,
			ExtensionFields: ExtensionFields{
				Err: ErrDataInconsistency,
			},
		})
		ev.OnEvent(&PackageProgress{
			ProgressStage: ProgressStageUnexpectedExit,
		})
		ev.OnEvent(&PackageProgress{
			ProgressStage: ProgressStageSuccessQuit,
			Record: map[string]*Package{
				"empty_root.txt": {FileName: "empty_root.txt", StreamBody: []byte("ok")},
			},
			ExtensionFields: ExtensionFields{
				TerminalPhoneNo:  "10086",
				ActiveSafetyType: consts.ActiveSafetyJS,
			},
		})
		if got, err := os.ReadFile(filepath.Join(tmp, "10086", "empty_root.txt")); err != nil || string(got) != "ok" {
			t.Fatalf("empty saveRoot file = %q err=%v, want ok", got, err)
		}

		// 覆盖 ProgressStage.String() 所有分支及默认分支
		for _, s := range []ProgressStage{
			ProgressStageInit, ProgressStageStart, ProgressStageStreamData,
			ProgressStageSupplementary, ProgressStageStreamDataComplete,
			ProgressStageComplete, ProgressStageSuccessQuit,
			ProgressStageFailQuit, ProgressStageUnexpectedExit,
		} {
			if s.String() == "" {
				t.Fatalf("stage %d returned empty string", s)
			}
		}
		if ProgressStage(255).String() != "" {
			t.Fatal("unknown stage should return empty string")
		}
	})
}

type mockListener struct {
	acceptCalls int
	conn        net.Conn
}

func (m *mockListener) Accept() (net.Conn, error) {
	m.acceptCalls++
	switch m.acceptCalls {
	case 1:
		return nil, errors.New("temporary accept error")
	case 2:
		return m.conn, nil
	default:
		return nil, net.ErrClosed
	}
}

func (m *mockListener) Close() error   { return nil }
func (m *mockListener) Addr() net.Addr { return &net.TCPAddr{} }

func TestGoJT808_NewAndRun(t *testing.T) {
	// 1. 覆盖所有 Option 与 listen 失败分支
	gFail := New(
		WithNetwork("invalid-net"),
		WithHostPorts("127.0.0.1:0"),
		WithActiveSafetyType(consts.ActiveSafetyHLJ),
		WithFileEventerFunc(func() FileEventer { return &stageRecorder{} }),
		WithDataHandleFunc(func() DataHandler { return newStandardJT808DataHandle(consts.ActiveSafetyHLJ) }),
		WithReadTimeout(100*time.Millisecond),
	)
	gFail.Run()

	// 2. 覆盖 Run 的 Accept 临时错误、正常连接处理、以及 net.ErrClosed 退出
	serverConn, clientConn := net.Pipe()
	connDone := make(chan struct{})

	origListen := netListen
	t.Cleanup(func() { netListen = origListen })
	netListen = func(_, _ string) (net.Listener, error) {
		return &mockListener{conn: serverConn}, nil
	}

	g := New(
		WithNetwork("tcp"),
		WithHostPorts("127.0.0.1:0"),
		WithFileEventerFunc(func() FileEventer {
			return &funcFileEvent{onEvent: func(p *PackageProgress) {
				if p.ProgressStage == ProgressStageUnexpectedExit {
					close(connDone)
				}
			}}
		}),
	)
	runDone := make(chan struct{})
	go func() {
		defer close(runDone)
		g.Run()
	}()

	_ = clientConn.Close()
	select {
	case <-connDone:
	case <-time.After(time.Second):
		t.Fatal("accepted connection did not finish")
	}
	select {
	case <-runDone:
	case <-time.After(time.Second):
		t.Fatal("Run did not exit on net.ErrClosed")
	}
}
