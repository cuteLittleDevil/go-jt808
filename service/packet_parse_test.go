package service

import (
	"encoding/hex"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

func Test_packageParse_unpack(t *testing.T) {
	type want struct {
		msgs           []string
		hasErr         bool
		historyDataLen int
	}
	tests := []struct {
		name string
		args string
		want want
	}{
		{
			name: "单个-完整包",
			args: "7e0100002d0144199999990001000b0065373034343358485830303030320000000000000000000000006964303030303301d4c1413138383838927e",
			want: want{
				msgs: []string{
					"7e0100002d0144199999990001000b0065373034343358485830303030320000000000000000000000006964303030303301d4c1413138383838927e",
				},
				hasErr:         false,
				historyDataLen: 0,
			},
		},
		{
			name: "单个-不完整包",
			args: "7e0100002d0144199999990001000b00653",
			want: want{
				msgs:           nil,
				hasErr:         false,
				historyDataLen: 17,
			},
		},
		{
			name: "单个-错误的包",
			args: "7e000200000123456789017fff7e",
			want: want{
				msgs:           nil,
				hasErr:         true,
				historyDataLen: 0,
			},
		},
		{
			name: "多个-完整包",
			args: "7e000200000123456789017fff0a7e7e0100002d0144199999990001000b0065373034343358485830303030320000000000000000000000006964303030303301d4c1413138383838927e",
			want: want{
				msgs: []string{
					"7e000200000123456789017fff0a7e",
					"7e0100002d0144199999990001000b0065373034343358485830303030320000000000000000000000006964303030303301d4c1413138383838927e",
				},
				hasErr:         false,
				historyDataLen: 0,
			},
		},
		{
			name: "多个-不完整包",
			args: "7e000200000144199999990007c07e7e800100000144199999990002467e7e0100002d0144199999990001000b00653",
			want: want{
				msgs: []string{
					"7e000200000144199999990007c07e",
					"7e800100000144199999990002467e",
				},
				hasErr:         false,
				historyDataLen: 17,
			},
		},
		{
			name: "多个-部分错误的包",
			args: "7e000200000144199999990007c07e7e800100000144199999997e7e000200000123456789017fff0a7e",
			want: want{
				msgs: []string{
					"7e000200000144199999990007c07e",
				},
				hasErr:         true,
				historyDataLen: 15,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPackageParse()
			argData, _ := hex.DecodeString(tt.args)
			gotMsgs, err := p.unpack(argData)
			if (err != nil) != tt.want.hasErr {
				t.Fatalf("unpack() error = %v, wantErr %v", err, tt.want.hasErr)
			}
			if len(gotMsgs) != len(tt.want.msgs) {
				t.Fatalf("unpack() gotMsgs len = %v, want %v", len(gotMsgs), len(tt.want.msgs))
			}
			if p.frameReader.Pending() != tt.want.historyDataLen {
				t.Fatalf("unpack() historyData len = %v, want %v", p.frameReader.Pending(), tt.want.historyDataLen)
			}
			for k, v := range gotMsgs {
				str := fmt.Sprintf("%x", v.ExtensionFields.TerminalData)
				if !reflect.DeepEqual(str, tt.want.msgs[k]) {
					t.Fatalf("unpack() gotMsgs = %s\n want %s", str, tt.want.msgs)
				}
			}
		})
	}
}

func Test_packageParse_subcontractLifecycle(t *testing.T) {
	base := mustDecodeJTMessage(t, heartbeatPacket)
	header := base.Header
	header.ReplyID = uint16(consts.T0801MultimediaDataUpload)
	header.PlatformSerialNumber = 10

	body := make([]byte, 1500)
	for i := range body {
		body[i] = byte(i)
	}
	packets := header.EncodePackets(body)

	p := newPackageParse()

	// 1. 首包未到达时先收到后续分包，忽略且不触发重组
	msgs, err := p.parse(packets[1])
	if err != nil || len(msgs) != 1 || msgs[0].ExtensionFields.SubcontractComplete {
		t.Fatalf("orphan subpacket msgs=%d err=%v", len(msgs), err)
	}

	// 2. 首包到达，随后再次收到新的首包覆盖旧的未完成分包记录
	if _, err := p.parse(packets[0]); err != nil {
		t.Fatalf("first packet err = %v", err)
	}
	firstMsgs, err := p.parse(packets[0])
	if err != nil || len(firstMsgs) != 1 {
		t.Fatalf("overwrite first packet err = %v", err)
	}
	origHeader := *firstMsgs[0].JTMessage.Header
	origProp := *firstMsgs[0].JTMessage.Header.Property

	// 3. 超过5s未收齐时，收到新报文触发 0x8003 补传第2包，且不污染已分发首包的 Header
	const subID = uint16(consts.T0801MultimediaDataUpload)
	p.records[subID].updateTime = time.Now().Add(-6 * time.Second)
	hbMsgs, err := p.parse(heartbeatPacket)
	if err != nil || len(hbMsgs) != 2 || hbMsgs[1].Command != consts.P8003ReissueSubcontractingRequest {
		t.Fatalf("expected heartbeat + P8003, got len=%d err=%v", len(hbMsgs), err)
	}
	if firstMsgs[0].JTMessage.Header.ReplyID != origHeader.ReplyID || *firstMsgs[0].JTMessage.Header.Property != origProp {
		t.Fatal("supplementarySubPackage mutated original message header")
	}
	if hbMsgs[1].Header.Property.BodyDayaLen != uint16(len(hbMsgs[1].Body)) ||
		strings.Contains(hbMsgs[1].Header.Property.String(), "是否分包:[true]") ||
		len(hbMsgs[1].ExtensionFields.TerminalData) == 0 {
		t.Fatalf("unexpected P8003 decoded header/data: %s", hbMsgs[1].Header.String())
	}

	// 4. 补齐第2包后完成重组
	doneMsgs, err := p.parse(packets[1])
	if err != nil || len(doneMsgs) != 2 || !doneMsgs[1].ExtensionFields.SubcontractComplete {
		t.Fatalf("expected completed subcontract message, got len=%d err=%v", len(doneMsgs), err)
	}
	if !reflect.DeepEqual(doneMsgs[1].Body, body) {
		t.Fatalf("reassembled body len=%d, want %d", len(doneMsgs[1].Body), len(body))
	}

	// 5. 缺失超过255包时，单次 0x8003 截断到 255 个
	bigHeader := *header
	prop := *header.Property
	bigHeader.Property = &prop
	bigHeader.ID = subID
	bigHeader.SubPackageSum = 300
	p.add(subID, &bigHeader)
	p.records[subID].updateTime = time.Now().Add(-6 * time.Second)
	reissueMsgs, ok := p.supplementarySubPackage()
	if !ok || len(reissueMsgs) != 1 {
		t.Fatalf("supplementarySubPackage ok=%v len=%d", ok, len(reissueMsgs))
	}
	var p8003 model.P0x8003
	if err := p8003.Parse(reissueMsgs[0].JTMessage); err != nil {
		t.Fatalf("parse P0x8003 err = %v", err)
	}
	if p8003.AgainPackageCount != math.MaxUint8 || len(p8003.AgainPackageList) != math.MaxUint8 {
		t.Fatalf("count=%d len=%d, want 255", p8003.AgainPackageCount, len(p8003.AgainPackageList))
	}

	// 6. 超过60s未完成的分包自动超时清理，以及 clear 清理残余记录
	p.records[subID].createTime = time.Now().Add(-61 * time.Second)
	if _, err := p.parse(heartbeatPacket); err != nil {
		t.Fatalf("parse for timeout cleanup err = %v", err)
	}
	if len(p.records) != 0 {
		t.Fatalf("expected records empty after 60s expiry, got %d", len(p.records))
	}
	p.add(subID, &bigHeader)
	p.clear()
	if len(p.records) != 0 {
		t.Fatal("clear did not empty records map")
	}
}
