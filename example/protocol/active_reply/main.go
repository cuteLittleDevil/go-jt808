package main

import (
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"log/slog"
	"net"
	"os"
	"strings"
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
				consts.P8104QueryTerminalParams: &p8104{},
			}
		}),
	)
	go goJt808.Run()
	time.Sleep(time.Second) // 等待启动完成

	phone := "14419999999"
	go client(phone)
	time.Sleep(5 * time.Second) // 等待模拟器注册成功

	replyMsg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              phone,                           // 默认使用手机号作为唯一key 根据key找到对应终端的TCP链接
		Command:          consts.P8104QueryTerminalParams, // 下发的指令
		Body:             nil,                             // 下发的body数据 8104为空
		OverTimeDuration: 3 * time.Second,                 // 超时时间 设备这段时间没有回复则失败
	})
	extension := replyMsg.ExtensionFields
	fmt.Println("0x8104发送", time.Now().Format(time.RFC3339),
		fmt.Sprintf("seq[%d] %x", extension.PlatformSeq, extension.PlatformData))
	fmt.Println("0x8104应答", fmt.Sprintf("seq[%d] %x", extension.TerminalSeq, extension.TerminalData))

	var t0x0104 model.T0x0104
	if err := t0x0104.Parse(replyMsg.JTMessage); err != nil {
		panic(err)
	}
	fmt.Println(t0x0104.String())
}

func client(phone string) {
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, phone))
	var (
		register = t.CreateDefaultCommandData(consts.T0100Register)
		auth     = t.CreateDefaultCommandData(consts.T0102RegisterAuth)
	)
	conn, err := net.Dial("tcp", "0.0.0.0:8080")
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	go func() {
		data := make([]byte, 1023)
		for {
			if n, _ := conn.Read(data); n > 0 {
				msg := fmt.Sprintf("%x", data[:n])
				if strings.HasPrefix(msg, "7e8104") {
					msg0x0104 := "7E010443A20100000000014419999999000500035B00000001040000000A00000002040000003C00000003040000000200000004040000003C00000005040000000200000006040000003C000000070400000002000000100B31333031323334353637300000001105313233343500000012053132333435000000130E3132372E302E302E313A37303030000000140531323334350000001505313233343500000016053132333435000000170531323334350000001A093132372E302E302E310000001B04000004570000001C04000004580000001D093132372E302E302E310000002004000000000000002104000000000000002204000000000000002301300000002401300000002501300000002601300000002704000000000000002804000000000000002904000000000000002C04000003E80000002D04000003E80000002E04000003E80000002F04000003E800000030040000000A0000003102003C000000320416320A1E000000400B3133303132333435363731000000410B3133303132333435363732000000420B3133303132333435363733000000430B3133303132333435363734000000440B3133303132333435363735000000450400000001000000460400000000000000470400000000000000480B3133303132333435363738000000490B313330313233343536373900000050040000000000000051040000000000000052040000000000000053040000000000000054040000000000000055040000003C000000560400000014000000570400003840000000580400000708000000590400001C200000005A040000012C0000005B0200500000005C0200050000005D02000A0000005E02001E00000064040000000100000065040000000100000070040000000100000071040000006F000000720400000070000000730400000071000000740400000072000000751500030190320000002800030190320000002800050100000076130400000101000002020000030300000404000000000077160101000301F43200000028000301F43200000028000500000079032808010000007A04000000230000007B0232320000007C1405000000000000000000000000000000000000000000008004000000240000008102000B000000820200660000008308BEA9415830303031000000840101000000900102000000910101000000920101000000930400000001000000940100000000950400000001000001000400000064000001010213880000010204000000640000010302138800000110080000000000000101F77E"
					writeData, _ := hex.DecodeString(msg0x0104)
					time.Sleep(time.Second)
					_, _ = conn.Write(writeData)
				}
			}
		}
	}()

	_, _ = conn.Write(auth)
	time.Sleep(time.Second)
	_, _ = conn.Write(register)
	time.Sleep(time.Second)
	_, _ = conn.Write(auth)
	select {}

}

type p8104 struct {
	model.P0x8104
}

func (p p8104) OnReadExecutionEvent(_ *service.Message) {
}

func (p p8104) OnWriteExecutionEvent(message service.Message) {
	fmt.Println("0x8104发送", time.Now().Format(time.RFC3339),
		fmt.Sprintf("%x", message.ExtensionFields.PlatformData))
}
