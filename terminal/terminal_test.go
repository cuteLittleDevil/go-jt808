package terminal

import (
	"bytes"
	"fmt"
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
