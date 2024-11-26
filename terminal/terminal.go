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
	TerminalPhoneNo string
	header          *jt808.Header
	protocolHandles map[consts.JT808CommandType]Handler
}

func New(opts ...Option) *Terminal {
	var header *jt808.Header
	options := newOptions(opts)
	if options.Header != nil {
		header = options.Header
	}
	if header == nil {
		msg := "7e000200001234567820130001387e"
		jtMsg := jt808.NewJTMessage()
		data, _ := hex.DecodeString(msg)
		_ = jtMsg.Decode(data)
		header = jtMsg.Header
	}
	protocolHandles := defaultProtocolHandles(header.ProtocolVersion)
	if options.CustomProtocolHandleFunc != nil {
		for commandType, handler := range options.CustomProtocolHandleFunc() {
			protocolHandles[commandType] = handler
		}
	}
	return &Terminal{
		TerminalPhoneNo: header.TerminalPhoneNo,
		header:          header,
		protocolHandles: protocolHandles,
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

func (t *Terminal) CreateCommandData(commandType consts.JT808CommandType, body []byte) []byte {
	t.header.ReplyID = uint16(commandType)
	t.header.PlatformSerialNumber++
	return t.header.Encode(body)
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
		body, _ = v.ReplyBody(jtMsg)
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
		if err := v.Parse(jtMsg); err != nil {
			slog.Warn("parse fail",
				slog.String("msg", msg))
			return ""
		}
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
