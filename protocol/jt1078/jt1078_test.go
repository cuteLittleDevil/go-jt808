package jt1078

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"testing"
)

func TestPacketDecode(t *testing.T) {
	type want struct {
		msg string
		err error
	}
	tests := []struct {
		name string
		args string
		want want
	}{
		{
			name: "原子包-视频I祯",
			args: "3031636481e20000295696659617010000000000000000000000000000020000",
			want: want{
				msg: "",
				err: nil,
			},
		},
		{
			name: "错误数据",
			args: "2031636481e20000295696659617010000000000000000000000000000020000",
			want: want{
				msg: "2031636481e20000295696659617010000000000000000000000000000020000",
				err: protocol.ErrUnqualifiedData,
			},
		},
		{
			name: "head不足 最低标准长度16",
			args: "3031636481e200",
			want: want{
				msg: "3031636481e200",
				err: protocol.ErrHeaderLength2Short,
			},
		},
		{
			name: "head不足 不确定的数据",
			args: "3031636481e2000029569665961701000000000000000000000000000",
			want: want{
				msg: "3031636481e2000029569665961701000000000000000000000000000",
				err: protocol.ErrHeaderLength2Short,
			},
		},
		{
			name: "body不足",
			args: "3031636481e200002956966596170100000000000000000000000000000200",
			want: want{
				msg: "00",
				err: protocol.ErrBodyLength2Short,
			},
		},
		{
			name: "body盈余",
			args: "3031636481e200002956966596170100000000000000000000000000000200003031636481e20000295696659617010000000000000000000000000000020000",
			want: want{
				msg: "3031636481e20000295696659617010000000000000000000000000000020000",
				err: nil,
			},
		},
		{
			name: "视频I帧-分包处理时的第一个包",
			args: "3031636481e20000295696659617010100000000000000000000000000020000",
			want: want{
				msg: "",
				err: nil,
			},
		},
		{
			name: "视频I帧-分包处理时的中间包",
			args: "3031636481e20000295696659617010300000000000000000000000000020000",
			want: want{
				msg: "",
				err: nil,
			},
		},
		{
			name: "视频I帧-分包处理时的最后一个包",
			args: "3031636481e20000295696659617010200000000000000000000000000020000",
			want: want{
				msg: "",
				err: nil,
			},
		},
		{
			name: "透传数据",
			args: "3031636481e200002956966596170240000101",
			want: want{
				msg: "",
				err: nil,
			},
		},
		{
			name: "G711A",
			args: "3031636481060000295696659617010000000000000000000000000000020000",
			want: want{
				msg: "",
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := hex.DecodeString(tt.args)
			p := NewPacket()
			remainData, err := p.Decode(data)
			if !errors.Is(err, tt.want.err) {
				t.Errorf("Encode() = %s\n want %s", err, tt.want.err)
				return
			}
			wantData, _ := hex.DecodeString(tt.want.msg)
			if bytes.Compare(remainData, wantData) != 0 {
				t.Errorf("Encode() = %x\n want %x", remainData, wantData)
				return
			}
			_ = p.String()
		})
	}

}
