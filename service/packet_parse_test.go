package service

import (
	"encoding/hex"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/model"
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
				t.Errorf("unpack() error = %v, wantErr %v", err, tt.want.hasErr)
				return
			}
			if len(gotMsgs) != len(tt.want.msgs) {
				t.Errorf("unpack() gotMsgs len = %v, want %v", len(gotMsgs), len(tt.want.msgs))
				return
			}
			if p.frameReader.Pending() != tt.want.historyDataLen {
				t.Errorf("unpack() historyData len = %v, want %v", p.frameReader.Pending(), tt.want.historyDataLen)
				return
			}
			for k, v := range gotMsgs {
				str := fmt.Sprintf("%x", v.ExtensionFields.TerminalData)
				if !reflect.DeepEqual(str, tt.want.msgs[k]) {
					t.Errorf("unpack() gotMsgs = %s\n want %s", str, tt.want.msgs)
					return
				}
			}
		})
	}
}

func Test_packageParse_supplementarySubPackageMax255(t *testing.T) {
	header := *mustDecodeJTMessage(t, heartbeatPacket).Header
	prop := *header.Property
	header.Property = &prop
	header.ID = 0x0801
	header.SubPackageSum = 300

	p := newPackageParse()
	p.add(header.ID, &header)
	p.timeoutRecord[header.ID].updateTime = time.Now().Add(-6 * time.Second)

	msgs, ok := p.supplementarySubPackage()
	if !ok || len(msgs) != 1 {
		t.Fatalf("supplementarySubPackage ok=%v len=%d", ok, len(msgs))
	}
	var p8003 model.P0x8003
	if err := p8003.Parse(msgs[0].JTMessage); err != nil {
		t.Fatalf("parse P0x8003 err = %v", err)
	}
	if p8003.AgainPackageCount != math.MaxUint8 || len(p8003.AgainPackageList) != math.MaxUint8 {
		t.Fatalf("count=%d len=%d, want 255", p8003.AgainPackageCount, len(p8003.AgainPackageList))
	}
	if first, last := p8003.AgainPackageList[0], p8003.AgainPackageList[254]; first != 1 || last != 255 {
		t.Fatalf("list range [%d, %d], want [1, 255]", first, last)
	}
}

func Test_packageParse_supplementaryNotModifyOriginalHeader(t *testing.T) {
	header := *mustDecodeJTMessage(t, heartbeatPacket).Header
	prop := *header.Property
	prop.PacketFragmented = 1
	header.Property = &prop
	header.ID = 0x0801
	header.SubPackageSum = 3
	header.SubPackageNo = 1
	want := header
	wantProp := prop

	p := newPackageParse()
	p.add(header.ID, &header)
	p.timeoutRecord[header.ID].updateTime = time.Now().Add(-6 * time.Second)
	if _, ok := p.supplementarySubPackage(); !ok {
		t.Fatal("supplementarySubPackage not triggered")
	}
	if header.ReplyID != want.ReplyID || *header.Property != wantProp {
		t.Fatalf("original header modified: ReplyID=%x Property=%+v", header.ReplyID, *header.Property)
	}
}
