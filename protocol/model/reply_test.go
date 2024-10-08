package model

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"reflect"
	"testing"
)

func TestReply(t *testing.T) {
	type Handler interface {
		HasReply() bool
		ReplyBody(*jt808.JTMessage) ([]byte, error)
		ReplyProtocol() consts.JT808CommandType
	}
	type args struct {
		Handler
		msg2011 string
		msg2013 string
		msg2019 string
	}
	type want struct {
		result2011 string
		result2013 string
		result2019 string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			// 测试的数据使用terminal.go中的CreateTerminalPackage生成
			// 终端和平台的流水号都使用0
			name: "T0x0002 终端-心跳",
			args: args{
				Handler: &T0x0002{},
				msg2013: "7e000200000123456789017fff0a7e",
				msg2019: "7e000240000100000000017299841738ffff027e",
			},
			want: want{
				result2013: "7e8001000501234567890100007fff0002008e7e",
				result2019: "7e8001400501000000000172998417380000ffff000200867e",
			},
		},
		{
			name: "T0x0001 终端-通用应答",
			args: args{
				Handler: &T0x0001{},
				msg2013: "7e000100050123456789017fff007b01c803bd7e",
			},
			want: want{
				result2013: "7e8001000501234567890100007fff0001008d7e",
			},
		},
		{
			name: "T0x0102 终端-鉴权",
			args: args{
				Handler: &T0x0102{},
				msg2013: "7e0102000b01234567890100003137323939383431373338b57e",
				msg2019: "7e0102402f010000000001729984173800000b3137323939383431373338313233343536373839303132333435332e372e31350000000000000000000000000000227e",
			},
			want: want{
				result2013: "7e80010005012345678901000000000102010e7e",
				result2019: "7e80014005010000000001729984173800000000010200877e",
			},
		},
		{
			name: "T0x0100 终端-注册",
			args: args{
				Handler: &T0x0100{},
				msg2011: "7e010000200123456789010000001f007363640000007777772e3830382e3736353433323101b2e24131323334a17e",
				msg2013: "7e0100002c0123456789010000001f007363640000007777772e3830382e636f6d0000000000000000003736353433323101b2e24131323334cc7e",
				msg2019: "7e0100405301000000000172998417380000001f007363640000000000000000007777772e3830382e636f6d0000000000000000000000000000000000000037363534333231000000000000000000000000000000000000000000000001b2e241313233343b7e",
			},
			want: want{
				result2011: "7e8100000e01234567890100000000003132333435363738393031377e",
				result2013: "7e8100000e01234567890100000000003132333435363738393031377e",
				result2019: "7e8100400e010000000001729984173800000000003137323939383431373338ba7e",
			},
		},
		{
			name: "T0x0200 终端-位置信息",
			args: args{
				Handler: &T0x0200{},
				msg2013: "7e0200007c0123456789017fff000004000000080006eeb6ad02633df701380003006320070719235901040000000b02020016030200210402002c051e3737370000000000000000000000000000000000000000000000000000001105420000004212064d0000004d4d1307000000580058582504000000632a02000a2b040000001430011e3101286b7e",
				msg2019: "7e0200407c0100000000017299841738ffff000004000000080006eeb6ad02633df701380003006320070719235901040000000b02020016030200210402002c051e3737370000000000000000000000000000000000000000000000000000001105420000004212064d0000004d4d1307000000580058582504000000632a02000a2b040000001430011e310128637e",
			},
			want: want{
				result2013: "7e8001000501234567890100007fff0200008e7e",
				result2019: "7e8001400501000000000172998417380000ffff020000867e",
			},
		},
		{
			name: "T0x0704 终端-位置信息批量上传",
			args: args{
				Handler: &T0x0704{},
				msg2013: "7e0704005d0123456789017fff000301001c000004000000080006eeb6ad02633df7013800030063200707192359001c000004000000080006eeb6ad02633df7013800030063200707192359001c000004000000080006eeb6ad02633df7013800030063200707192359067e",
				msg2019: "7e0704405d0100000000017299841738ffff000301001c000004000000080006eeb6ad02633df7013800030063200707192359001c000004000000080006eeb6ad02633df7013800030063200707192359001c000004000000080006eeb6ad02633df70138000300632007071923590e7e",
			},
			want: want{
				result2013: "7e8001000501234567890100007fff0704008f7e",
				result2019: "7e8001400501000000000172998417380000ffff070400877e",
			},
		},
	}
	checkReplyInfo := func(t *testing.T, msg string, handler Handler, expectedResult string) {
		if msg == "" {
			return
		}
		data, _ := hex.DecodeString(msg)
		jtMsg := jt808.NewJTMessage()
		if err := jtMsg.Decode(data); err != nil {
			t.Errorf("Parse() error = %v", err)
			return
		}
		jtMsg.Header.ReplyID = uint16(handler.ReplyProtocol())
		if ok := handler.HasReply(); !ok {
			return
		}
		body, _ := handler.ReplyBody(jtMsg)
		got := jtMsg.Header.Encode(body)
		if !reflect.DeepEqual(fmt.Sprintf("%x", got), expectedResult) {
			t.Errorf("ReplyInfo() got = [%x]\n want = [%s]", got, expectedResult)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkReplyInfo(t, tt.args.msg2011, tt.args.Handler, tt.want.result2011)
			checkReplyInfo(t, tt.args.msg2013, tt.args.Handler, tt.want.result2013)
			checkReplyInfo(t, tt.args.msg2019, tt.args.Handler, tt.want.result2019)
		})
	}
}

// 为了覆盖率100%增加的测试 ------------------------------------
func TestT0x0102Reply(t *testing.T) {
	msg := "7e0102402f010000000001729984173800000b3137323939383431373338313233343536373839303132333435332e372e31350000000000000000000000000000227e"
	data, _ := hex.DecodeString(msg)
	jtMsg := jt808.NewJTMessage()
	_ = jtMsg.Decode(data)
	handler := &T0x0102{}
	// 强制错误情况
	jtMsg.Body = nil
	if _, err := handler.ReplyBody(jtMsg); !errors.Is(err, protocol.ErrBodyLengthInconsistency) {
		t.Errorf("T0x0102 ReplyBody() err[%v]", err)
		return
	}
}

func TestT0x0002Encode(t *testing.T) {
	handler := &T0x0002{}
	got := handler.Encode()
	if got != nil {
		t.Errorf("T0x002 Encode() got = [%x]", got)
	}
}

func TestT0x0200Encode(t *testing.T) {
	handler := &T0x0200{
		T0x0200LocationItem: T0x0200LocationItem{
			AlarmSign:  1024,
			StatusSign: 2048,
			Latitude:   119552894,
			Longitude:  40058359,
			Altitude:   312,
			Speed:      3,
			Direction:  99,
			DateTime:   "2024-10-01 23:59:59",
		},
	}
	body := handler.Encode()
	got := fmt.Sprintf("%x", body)
	want := "000004000000080007203b7e02633df7013800030063241001235959"
	if got != want {
		t.Errorf("T0x0200 Encode() got = %s\n want = %s", got, want)
	}
}
