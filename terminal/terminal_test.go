package terminal

import (
	"bytes"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"io"
	"os"
	"testing"
	"time"
)

func TestTerminal_CreateDefaultCommandData(t *testing.T) {
	tests := []struct {
		name string
		args consts.JT808CommandType
	}{
		{
			name: "T0X0002 终端-心跳",
			args: consts.T0002HeartBeat,
		},
		{
			name: "T0X0100 终端-注册",
			args: consts.T0100Register,
		},
		{
			name: "T0X0102 终端-鉴权",
			args: consts.T0102RegisterAuth,
		},
		{
			name: "T0x0200 终端-位置上报",
			args: consts.T0200LocationReport,
		},
		{
			name: "T0x0704 终端-位置批量上传",
			args: consts.T0704LocationBatchUpload,
		},
		{
			name: "P0x8001 平台-通用应答",
			args: consts.P8001GeneralRespond,
		},
		{
			name: "P0x8100 平台-注册应答",
			args: consts.P8100RegisterRespond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			infos := map[consts.ProtocolVersionType]string{
				consts.JT808Protocol2011: "123456782011",
				consts.JT808Protocol2013: "123456782013",
				consts.JT808Protocol2019: "12345678901234562019",
			}
			for versionType, phone := range infos {
				t.Run(versionType.String(), func(t *testing.T) {
					tmp := New(WithHeader(versionType, phone))
					sendData := tmp.CreateDefaultCommandData(tt.args)
					msg := fmt.Sprintf("%x", sendData)
					details := tmp.ProtocolDetails(msg)
					replyData := tmp.ExpectedReply(1, msg)
					replyDetails := tmp.ProtocolDetails(fmt.Sprintf("%x", replyData))
					got := fmt.Sprintf("%x\n%s\n-----------\n%x\n%s",
						sendData, details, replyData, replyDetails)
					txt := fmt.Sprintf("./testdata/[%04x]-%s-%s.txt",
						uint16(tt.args), versionType.String(), tt.args)
					f, err := os.Open(txt)
					if err != nil {
						_ = os.WriteFile(txt, []byte(got), os.ModePerm)
					}
					if data, _ := io.ReadAll(f); string(data) != got {
						t.Errorf("CreateDefaultCommandData()=%s\n want %s", got, string(data))
					}
				})
			}
		})

	}
}

func TestTerminal_CreateCustomMessageFunc(t *testing.T) {
	t1 := New(WithHeader(consts.JT808Protocol2013, "1"))
	body1 := t1.CreateDefaultCommandData(consts.T0200LocationReport)

	t2 := New(WithHeader(consts.JT808Protocol2013, "1"))
	t2.CreateCustomMessageFunc = func(commandType consts.JT808CommandType) (Handler, bool) {
		if commandType == consts.T0200LocationReport {
			return &model.T0x0200{
				T0x0200LocationItem: model.T0x0200LocationItem{
					AlarmSign:  1024,
					StatusSign: 2048,
					Latitude:   116307629,
					Longitude:  40058359,
					Altitude:   312,
					Speed:      3,
					Direction:  99,
					DateTime:   time.Now().Format(time.DateTime),
				},
			}, true
		}
		return nil, false
	}
	body2 := t2.CreateDefaultCommandData(consts.T0200LocationReport)

	if bytes.Compare(body1, body2) == 0 {
		t.Errorf("CreateCustomMessageFunc body1=[%x]\n body2=[%x]\n", body1, body2)
	}
}

func TestTerminal_AllDefaultProtocols(t *testing.T) {
	const wantProtocolCount = 41
	versions := []struct {
		version consts.ProtocolVersionType
		phone   string
	}{
		{version: consts.JT808Protocol2013, phone: "123456782013"},
		{version: consts.JT808Protocol2019, phone: "12345678901234562019"},
	}
	for _, v := range versions {
		t.Run(v.version.String(), func(t *testing.T) {
			term := New(WithHeader(v.version, v.phone))
			if got := len(term.protocolHandles); got != wantProtocolCount {
				t.Fatalf("protocolHandles count = %d, want %d", got, wantProtocolCount)
			}
			for cmd := range term.protocolHandles {
				sendData := term.CreateDefaultCommandData(cmd)
				if len(sendData) == 0 {
					t.Errorf("cmd %s: CreateDefaultCommandData returned empty", cmd)
					continue
				}
				msgHex := fmt.Sprintf("%x", sendData)
				if details := term.ProtocolDetails(msgHex); details == "" {
					t.Errorf("cmd %s: ProtocolDetails returned empty", cmd)
				}
				replyData := term.ExpectedReply(1, msgHex)
				jtMsg := jt808.NewJTMessage()
				if err := jtMsg.Decode(replyData); err != nil {
					t.Errorf("cmd %s: ExpectedReply decode err = %v", cmd, err)
				}
			}
		})
	}
}

func TestTerminal_CreateCommandPackets(t *testing.T) {
	term := New(WithHeader(consts.JT808Protocol2013, "1001"))

	// Single packet
	single := term.CreateDefaultCommandPackets(consts.T0002HeartBeat)
	if len(single) != 1 {
		t.Fatalf("single packets len = %d, want 1", len(single))
	}
	jtSingle := jt808.NewJTMessage()
	if err := jtSingle.Decode(single[0]); err != nil {
		t.Fatalf("decode single packet err = %v", err)
	}
	if jtSingle.Header.SerialNumber != 1 {
		t.Errorf("single SerialNumber = %d, want 1", jtSingle.Header.SerialNumber)
	}

	// Multi-packet body (2500 bytes -> 3 packets: 1000 + 1000 + 500)
	payload := bytes.Repeat([]byte{0xab}, 2500)
	packets := term.CreateCommandPackets(consts.T0801MultimediaDataUpload, payload)
	if len(packets) != 3 {
		t.Fatalf("multi packets len = %d, want 3", len(packets))
	}
	var reassembled []byte
	for i, pkt := range packets {
		jtMsg := jt808.NewJTMessage()
		if err := jtMsg.Decode(pkt); err != nil {
			t.Fatalf("packet %d decode err = %v", i, err)
		}
		if jtMsg.Header.SubPackageSum != 3 || jtMsg.Header.SubPackageNo != uint16(i+1) {
			t.Errorf("packet %d subpackage = %d/%d, want %d/3",
				i, jtMsg.Header.SubPackageNo, jtMsg.Header.SubPackageSum, i+1)
		}
		wantSeq := uint16(2 + i)
		if jtMsg.Header.SerialNumber != wantSeq {
			t.Errorf("packet %d SerialNumber = %d, want %d", i, jtMsg.Header.SerialNumber, wantSeq)
		}
		reassembled = append(reassembled, jtMsg.Body...)
	}
	if !bytes.Equal(reassembled, payload) {
		t.Errorf("reassembled payload len = %d, want %d", len(reassembled), len(payload))
	}

	// Next packet should continue sequence number from 5
	next := term.CreateDefaultCommandPackets(consts.T0002HeartBeat)
	jtNext := jt808.NewJTMessage()
	if err := jtNext.Decode(next[0]); err != nil {
		t.Fatalf("decode next packet err = %v", err)
	}
	if jtNext.Header.SerialNumber != 5 {
		t.Errorf("next SerialNumber = %d, want 5", jtNext.Header.SerialNumber)
	}

	// Unknown command returns nil
	if got := term.CreateDefaultCommandPackets(consts.JT808CommandType(0xffff)); got != nil {
		t.Errorf("unknown command packets = %v, want nil", got)
	}
}
