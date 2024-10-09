package terminal

import (
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"strings"
)

type Terminal struct {
	header          *jt808.Header
	protocolHandles map[consts.JT808CommandType]Handler
}

func New(opts ...Option) *Terminal {
	msg := "7e000100050123456789017fff007b01c803bd7e"
	jtMsg := jt808.NewJTMessage()
	data, _ := hex.DecodeString(msg)
	_ = jtMsg.Decode(data)
	header := jtMsg.Header
	options := NewOptions(opts)
	if options.Header != nil {
		header = options.Header
	}
	return &Terminal{
		header:          header,
		protocolHandles: defaultProtocolHandles(),
	}
}

func (t *Terminal) CreateDefaultCommandData(commandType consts.JT808CommandType) []byte {
	if v, ok := t.protocolHandles[commandType]; ok {
		body := v.Encode()
		t.header.ReplyID = uint16(commandType)
		t.header.PlatformSerialNumber++
		return t.header.Encode(body)
	}
	slog.Warn("not found command",
		slog.String("command", commandType.String()))
	return nil
}

func (t *Terminal) ExpectedReply(seq uint16, msg string) []byte {
	jtMsg := jt808.NewJTMessage()
	data, _ := hex.DecodeString(msg)
	if err := jtMsg.Decode(data); err != nil {
		slog.Warn("decode fail",
			slog.String("msg", msg),
			slog.Any("err", err))
		return nil
	}
	header := jtMsg.Header
	var body []byte
	commandType := consts.JT808CommandType(header.ID)
	if v, ok := t.protocolHandles[commandType]; ok {
		header.ReplyID = uint16(v.ReplyProtocol())
		header.PlatformSerialNumber = seq
		body = v.Encode()
	}
	return header.Encode(body)
}

func (t *Terminal) ProtocolDetails(msg string) string {
	jtMsg := jt808.NewJTMessage()
	data, _ := hex.DecodeString(msg)
	if err := jtMsg.Decode(data); err != nil {
		slog.Warn("decode fail",
			slog.String("msg", msg),
			slog.Any("err", err))
		return ""
	}
	commandType := consts.JT808CommandType(jtMsg.Header.ID)
	if v, ok := t.protocolHandles[commandType]; ok {
		_ = v.Parse(jtMsg)
		return strings.Join([]string{
			"[7e]开始: 126",
			jtMsg.Header.String(),
			v.String(),
			fmt.Sprintf("[%02x] 校验码:[%d]", jtMsg.VerifyCode, jtMsg.VerifyCode),
			"[7e]结束: 126",
		}, "\n")
	}
	slog.Warn("not parse msg",
		slog.String("msg", msg))
	return ""
}
