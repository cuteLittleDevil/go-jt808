package terminal

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"io"
	"os"
	"testing"
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			infos := map[consts.ProtocolVersionType]string{
				consts.JT808Protocol2011: "123456789098",
				consts.JT808Protocol2013: "123456789098",
				consts.JT808Protocol2019: "12345678901234567890",
			}
			for versionType, phone := range infos {
				t.Run(versionType.String(), func(t *testing.T) {
					tmp := New(WithHeader(versionType, phone))
					sendData := tmp.CreateDefaultCommandData(tt.args)
					msg := fmt.Sprintf("%x", sendData)
					details := tmp.ProtocolDetails(msg)
					got := fmt.Sprintf("%x\n%s", sendData, details)
					txt := fmt.Sprintf("./testdata/[%04x]-%s-%s.txt",
						uint16(tt.args), versionType.String(), tt.args)
					f, err := os.Open(txt)
					if err != nil {
						_ = os.WriteFile(txt, []byte(got), os.ModePerm)
					}
					if data, _ := io.ReadAll(f); string(data) != got {
						t.Errorf("CreateDefaultCommandData() = %s\n want %s", got, string(data))
					}
				})
			}
		})

	}
}
