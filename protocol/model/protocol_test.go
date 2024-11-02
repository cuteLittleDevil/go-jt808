package model

import (
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"testing"
)

func TestReplyProtocol(t *testing.T) {
	type Handler interface {
		Protocol() consts.JT808CommandType
		ReplyProtocol() consts.JT808CommandType
	}
	tests := []struct {
		name              string
		args              Handler
		wantProtocol      consts.JT808CommandType
		wantReplyProtocol consts.JT808CommandType
	}{
		{
			name:              "T0x0001 终端-通用应答",
			args:              &T0x0001{},
			wantProtocol:      consts.T0001GeneralRespond,
			wantReplyProtocol: consts.P8001GeneralRespond,
		},
		{
			name:              "P0x8001 平台-通用应答",
			args:              &P0x8001{},
			wantProtocol:      consts.P8001GeneralRespond,
			wantReplyProtocol: 0,
		},
		{
			name:              "P0x8100 平台-注册消息应答",
			args:              &P0x8100{},
			wantProtocol:      consts.P8100RegisterRespond,
			wantReplyProtocol: 0,
		},
		{
			name:              "T0x0002 终端-心跳",
			args:              &T0x0002{},
			wantProtocol:      consts.T0002HeartBeat,
			wantReplyProtocol: consts.P8001GeneralRespond,
		},
		{
			name:              "T0x0102 终端-鉴权",
			args:              &T0x0102{},
			wantProtocol:      consts.T0102RegisterAuth,
			wantReplyProtocol: consts.P8001GeneralRespond,
		},
		{
			name:              "T0x0100 终端-注册",
			args:              &T0x0100{},
			wantProtocol:      consts.T0100Register,
			wantReplyProtocol: consts.P8100RegisterRespond,
		},
		{
			name:              "T0x0200 终端-位置上报",
			args:              &T0x0200{},
			wantProtocol:      consts.T0200LocationReport,
			wantReplyProtocol: consts.P8001GeneralRespond,
		},
		{
			name:              "T0x0704 终端-位置批量上传",
			args:              &T0x0704{},
			wantProtocol:      consts.T0704LocationBatchUpload,
			wantReplyProtocol: consts.P8001GeneralRespond,
		},
		{
			name:              "P0x8104 平台-查询终端参数",
			args:              &P0x8104{},
			wantProtocol:      consts.P8104QueryTerminalParams,
			wantReplyProtocol: consts.T0104QueryParameter,
		},
		{
			name:              "T0x0104 终端-查询参数回复",
			args:              &T0x0104{},
			wantProtocol:      consts.T0104QueryParameter,
			wantReplyProtocol: 0,
		},
		{
			name:              "P0x9003 平台-查询终端音视频属性",
			args:              &P0x9003{},
			wantProtocol:      consts.P9003QueryTerminalAudioVideoProperties,
			wantReplyProtocol: consts.T1003UploadAudioVideoAttr,
		},
		{
			name:              "T0x1003 终端-查询参数回复",
			args:              &T0x1003{},
			wantProtocol:      consts.T1003UploadAudioVideoAttr,
			wantReplyProtocol: 0,
		},
		{
			name:              "P0x9101 平台-实时音视频传输请求",
			args:              &P0x9101{},
			wantProtocol:      consts.P9101RealTimeAudioVideoRequest,
			wantReplyProtocol: consts.T0001GeneralRespond,
		},
		{
			name:              "T0x1005 终端-上传乘客流量",
			args:              &T0x1005{},
			wantProtocol:      consts.T1005UploadPassengerFlow,
			wantReplyProtocol: 0,
		},
		{
			name:              "P0x9102 平台-音视频实时传输控制",
			args:              &P0x9102{},
			wantProtocol:      consts.P9102AudioVideoControl,
			wantReplyProtocol: consts.T0001GeneralRespond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.Protocol() != tt.wantProtocol {
				t.Errorf("Protocol() = %v, want %v", tt.args.Protocol(), tt.wantProtocol)
			}
			if tt.args.ReplyProtocol() != tt.wantReplyProtocol {
				t.Errorf("ReplyProtocol() = %v, want %v", tt.args.ReplyProtocol(), tt.wantReplyProtocol)
			}
		})
	}
}
