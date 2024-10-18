package main

import (
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"net"
	"os"
	"time"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(logger)
}

func main() {
	goJt808 := service.New(
		service.WithHostPorts("0.0.0.0:8080"),
		service.WithNetwork("tcp"),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				consts.T0200LocationReport: &Location{},
			}
		}),
	)
	go goJt808.Run()

	time.Sleep(time.Second)
	conn, _ := net.Dial("tcp", "127.0.0.1:8080")
	msg := "7E0200007B0123456789017FFF000004000000080006EEB6AD02633DF701380003006320070719235901040000000B02020016030200210402002C051E37373700000000000000000000000000000000000000000000000000000011010012064D0000004D4D1307000000580058582504000000632A02000A2B040000001430011E3101283301207A7E"
	data, _ := hex.DecodeString(msg)
	_, _ = conn.Write(data)

	ticker := time.NewTicker(3 * time.Second)
	for range ticker.C {
		_, _ = conn.Write(data)
	}
}

type Location struct {
	model.T0x0200
	customMile  int
	customValue uint8
}

func (l *Location) Parse(jtMsg *jt808.JTMessage) error {
	l.T0x0200AdditionDetails.CustomAdditionContentFunc = func(id uint8, content []byte) (model.AdditionContent, bool) {
		if id == uint8(consts.A0x01Mile) {
			l.customMile = 100
		}
		if id == 0x33 {
			value := content[0]
			l.customValue = value
			return model.AdditionContent{
				Data:        content,
				CustomValue: value,
			}, true
		}
		return model.AdditionContent{}, false
	}
	return l.T0x0200.Parse(jtMsg)
}

func (l *Location) OnReadExecutionEvent(message *service.Message) {
	var tmp Location
	_ = tmp.Parse(message.JTMessage)
	fmt.Println(tmp.T0x0200AdditionDetails.String())
	if v, ok := tmp.Additions[consts.A0x01Mile]; ok {
		fmt.Println(fmt.Sprintf("里程[%d] 自定义辅助里程[%d]", v.Content.Mile, tmp.customMile))
	}
	id := consts.JT808LocationAdditionType(0x33)
	if v, ok := tmp.Additions[id]; ok {
		fmt.Println("自定义未知信息扩展", v.Content.CustomValue, tmp.customValue)
	}
}

func (l *Location) OnWriteExecutionEvent(_ service.Message) {}
