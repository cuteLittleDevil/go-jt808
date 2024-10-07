package model

import (
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"testing"
)

func TestReplyProtocol(t *testing.T) {
	type Handler interface {
		Protocol() uint16
		ReplyProtocol() uint16
	}
	tests := []struct {
		name              string
		args              Handler
		wantProtocol      uint16
		wantReplyProtocol consts.PlatformReplyRequest
	}{
		{
			name:              "T0x0001 终端-通用应答",
			args:              &T0x0001{},
			wantProtocol:      uint16(consts.T0001GeneralRespond),
			wantReplyProtocol: consts.P8001GeneralRespond,
		},
		{
			name:              "P0x8001 平台-通用应答",
			args:              &P0x8001{},
			wantProtocol:      uint16(consts.P8001GeneralRespond),
			wantReplyProtocol: 0,
		},
		{
			name:              "P0x8100 终端-注册消息应答",
			args:              &P0x8100{},
			wantProtocol:      uint16(consts.P8100RegisterRespond),
			wantReplyProtocol: 0,
		},
		{
			name:              "T0x0002 终端-心跳",
			args:              &T0x0002{},
			wantProtocol:      uint16(consts.T0002HeartBeat),
			wantReplyProtocol: consts.P8001GeneralRespond,
		},
		{
			name:              "T0x0102 终端-鉴权",
			args:              &T0x0102{},
			wantProtocol:      uint16(consts.T0102RegisterAuth),
			wantReplyProtocol: consts.P8001GeneralRespond,
		},
		{
			name:              "T0x0100 终端-注册",
			args:              &T0x0100{},
			wantProtocol:      uint16(consts.T0100Register),
			wantReplyProtocol: consts.P8100RegisterRespond,
		},
		{
			name:              "T0x0200 终端-位置上报",
			args:              &T0x0200{},
			wantProtocol:      uint16(consts.T0200LocationReport),
			wantReplyProtocol: consts.P8001GeneralRespond,
		},
		{
			name:              "T0x0704 终端-位置批量上传",
			args:              &T0x0704{},
			wantProtocol:      uint16(consts.T0704LocationBatchUpload),
			wantReplyProtocol: consts.P8001GeneralRespond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.Protocol() != tt.wantProtocol {
				t.Errorf("Protocol() = %v, want %v", tt.args.Protocol(), tt.wantProtocol)
			}
			if tt.args.ReplyProtocol() != uint16(tt.wantReplyProtocol) {
				t.Errorf("ReplyProtocol() = %v, want %v", tt.args.ReplyProtocol(), uint16(tt.wantReplyProtocol))
			}
		})
	}
}
