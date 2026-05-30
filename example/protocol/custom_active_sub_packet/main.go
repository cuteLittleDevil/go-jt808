package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"log/slog"
	"net"
	"os"
	"sync"
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

const reqPacketLen = 3333

func main() {
	goJt808 := service.New(
		service.WithHostPorts("0.0.0.0:808"),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				// 自定义的，需要先加一下处理，表示这些指令支持了
				custom3333:      &CustomTerminalRequest{},
				custom3333Reply: &CustomTerminalReply{},
			}
		}),
		service.WithCustomActiveRespondHandlerFunc(func() map[consts.JT808CommandType]func(*service.ActiveMessage, *service.Message) bool {
			// 自定义的，需要自己去实现主动下发和回复的映射关系。不设置则默认超时
			// 每次都新建一个map，一个协程一个map
			return map[consts.JT808CommandType]func(activeMsg *service.ActiveMessage, terminalMsg *service.Message) bool{
				custom3333Reply: func(activeMsg *service.ActiveMessage, terminalMsg *service.Message) bool {
					var tmp CustomTerminalReply
					if err := tmp.Parse(terminalMsg.JTMessage); err != nil {
						return false
					}
					fmt.Println("平台发的序列号", activeMsg.ExtensionFields.PlatformSeq)
					return tmp.RespondSerialNumber == activeMsg.ExtensionFields.PlatformSeq
				}, // 自定义指令,发0x6660 -> 收到0x6661
			}
		}),
	)
	key := "1001"
	go goJt808.Run()
	time.Sleep(time.Second)
	go client(key, "127.0.0.1:808") // 模拟一个设备连接
	time.Sleep(time.Second)

	for i := 0; i < 2; i++ {
		req := CustomTerminalRequest{}
		reply6661Msg := goJt808.SendActiveMessage(&service.ActiveMessage{
			Key:              key,
			Command:          req.Protocol(),
			OverTimeDuration: 5 * time.Second,
			Body:             req.ToEncode(reqPacketLen),
		})
		if reply6661Msg.ExtensionFields.Err == nil {
			var customReply CustomTerminalReply
			err := customReply.Parse(reply6661Msg.JTMessage)
			fmt.Println("自定义指令 0x3333 -> 0x3334", customReply.RespondSerialNumber, err)
			fmt.Println(reply6661Msg.ExtensionFields.PlatformSeq, len(reply6661Msg.ExtensionFields.PlatformData))
		}
		time.Sleep(10 * time.Second)
	}
}

func client(phone string, address string) {
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, phone),
		terminal.WithCustomProtocolHandleFunc(func() map[consts.JT808CommandType]terminal.Handler {
			return map[consts.JT808CommandType]terminal.Handler{
				consts.T0805CameraShootImmediately: &model.T0x0805{},
				consts.T0801MultimediaDataUpload:   &model.T0x0801{},
				consts.P9208AlarmAttachUpload:      &model.P0x9208{},
			}
		}))
	register := t.CreateDefaultCommandData(consts.T0100Register)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	_, _ = conn.Write(register)

	data := make([]byte, 1023)
	historyData := make([]byte, 0, 2000)
	receive := make([]byte, 0, 3000)
	var once sync.Once
	for {
		if n, _ := conn.Read(data); n > 0 {
			historyData = append(historyData, data[:n]...)
		}

		const sign = 0x7e
		for {
			end := -1
			if len(historyData) > 2 && historyData[0] == sign {
				for i := 1; i < len(historyData); i++ {
					if historyData[i] == sign {
						end = i + 1
						break
					}
				}
			}
			if end == -1 {
				break
			}
			originalData := historyData[:end]
			jtMsg := jt808.NewJTMessage()
			if err := jtMsg.Decode(originalData); err == nil {
				fmt.Println("服务器->客户端的流水号", jtMsg.Header.SerialNumber)
				if jtMsg.Header.ID == uint16(custom3333) {
					receive = append(receive, jtMsg.Body...)
					if len(receive) == reqPacketLen {
						fmt.Println("全部收到")
						receive = receive[0:0]
						tmp := CustomTerminalReply{
							RespondSerialNumber: jtMsg.Header.SerialNumber,
						}
						jtMsg.Header.ReplyID = uint16(tmp.Protocol())
						_, _ = conn.Write(jtMsg.Header.Encode(tmp.Encode()))

						once.Do(func() {
							go func() {
								ticker := time.NewTicker(1 * time.Second)
								for range ticker.C {
									_, _ = conn.Write(t.CreateDefaultCommandData(consts.T0002HeartBeat))
								}
							}()
						})
					}
				}
				historyData = historyData[end:]
			} else {
				break
			}
		}
	}
}
