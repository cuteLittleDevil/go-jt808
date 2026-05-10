package main

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
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
	activeRespondHandlers := map[consts.JT808CommandType]func(activeMsg *service.ActiveMessage, terminalMsg *service.Message) bool{
		0x6661: func(activeMsg *service.ActiveMessage, terminalMsg *service.Message) bool {
			var tmp CustomTerminalReply
			if err := tmp.Parse(terminalMsg.JTMessage); err != nil {
				return false
			}
			fmt.Println("平台发的序列号", activeMsg.ExtensionFields.PlatformSeq)
			return tmp.RespondSerialNumber == activeMsg.ExtensionFields.PlatformSeq
		}, // 自定义指令,发0x6660 -> 收到0x6661
	}
	goJt808 := service.New(
		service.WithHostPorts("0.0.0.0:808"),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				consts.JT808CommandType(0x6660): &CustomTerminalRequest{},
				consts.JT808CommandType(0x6661): &CustomTerminalReply{},
			}
		}),
		service.WithCustomActiveRespondHandlerFunc(func() map[consts.JT808CommandType]func(*service.ActiveMessage, *service.Message) bool {
			return activeRespondHandlers
		}),
	)
	key := "1001"
	go goJt808.Run()
	time.Sleep(time.Second)
	go client(key, "127.0.0.1:808") // 模拟一个设备连接
	time.Sleep(time.Second)

	p9003 := model.P0x9003{}
	reply1003Msg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              key,
		Command:          p9003.Protocol(),
		Body:             p9003.Encode(),
		OverTimeDuration: 5 * time.Second,
	})
	var t0x1003 model.T0x1003
	err := t0x1003.Parse(reply1003Msg.JTMessage)
	fmt.Println(t0x1003.Protocol().String(), t0x1003.EnterAudioEncoding, err, reply1003Msg.ExtensionFields.Err)

	p9101 := model.P0x9101{
		BaseHandle: model.BaseHandle{},
		ChannelNo:  1,
	}
	reply9101Msg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              key,
		Command:          p9101.Protocol(),
		Body:             p9101.Encode(),
		OverTimeDuration: 5 * time.Second,
	})
	var t0x0001 model.T0x0001
	err = t0x0001.Parse(reply9101Msg.JTMessage)
	fmt.Println(t0x0001.Protocol().String(), t0x0001.Result, err)

	p8104 := model.P0x8104{}
	reply0104Msg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              key,
		Command:          p8104.Protocol(),
		Body:             p8104.Encode(),
		OverTimeDuration: 5 * time.Second,
	})
	var t0x0104 model.T0x0104
	err = t0x0104.Parse(reply0104Msg.JTMessage)
	fmt.Println(p8104.Protocol().String(), t0x0104.RespondSerialNumber, err, reply0104Msg.ExtensionFields.Err)

	p8201 := model.P0x8201{}
	reply8201Msg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              key,
		Command:          p8201.Protocol(),
		Body:             p8201.Encode(),
		OverTimeDuration: 5 * time.Second,
	})
	var t0x0201 model.T0x0201
	err = t0x0201.Parse(reply8201Msg.JTMessage)
	fmt.Println(p8201.Protocol().String(), t0x0201.RespondSerialNumber, err, reply8201Msg.ExtensionFields.Err)

	p9205 := model.P0x9205{
		BaseHandle: model.BaseHandle{},
		ChannelNo:  1,
		StartTime:  time.Now().Add(-24 * time.Hour).Format(time.DateTime),
		EndTime:    time.Now().Format(time.DateTime),
	}
	reply1205Msg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              key,
		Command:          p9205.Protocol(),
		Body:             p9205.Encode(),
		OverTimeDuration: 5 * time.Second,
	})
	var t0x1205 model.T0x1205
	err = t0x1205.Parse(reply1205Msg.JTMessage)
	fmt.Println(t0x1205.Protocol().String(), t0x1205.SerialNumber, err)

	p9206 := model.P0x9206{
		BaseHandle: model.BaseHandle{},
		ChannelNo:  1,
		StartTime:  time.Now().Add(-24 * time.Hour).Format(time.DateTime),
		EndTime:    time.Now().Format(time.DateTime),
	}
	reply1206Msg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              key,
		Command:          p9206.Protocol(),
		Body:             p9206.Encode(),
		OverTimeDuration: 5 * time.Second,
	})
	var t0x1206 model.T0x1206
	err = t0x1206.Parse(reply1206Msg.JTMessage)
	fmt.Println(t0x1206.Protocol().String(), t0x1206.RespondSerialNumber, err, reply1206Msg.ExtensionFields.Err)

	p8801 := model.P0x8801{
		ChannelID:                1,
		ShootCommand:             2,
		PhotoIntervalOrVideoTime: 3,
		SaveFlag:                 1,
		Resolution:               4,
		VideoQuality:             5,
		Intensity:                255,
		Contrast:                 127,
		Saturation:               127,
		Chroma:                   255,
	}
	reply8801Msg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              key,
		Command:          p8801.Protocol(),
		Body:             p8801.Encode(),
		OverTimeDuration: 5 * time.Second,
	})
	var t0x0805 model.T0x0805
	err = t0x0805.Parse(reply8801Msg.JTMessage)
	fmt.Println(t0x0805.Protocol().String(), t0x0805.RespondSerialNumber, err, reply8801Msg.ExtensionFields.Err)

	reply6661Msg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              key,
		Command:          consts.JT808CommandType(0x6660),
		OverTimeDuration: 5 * time.Second,
	})
	var customReply CustomTerminalReply
	err = customReply.Parse(reply6661Msg.JTMessage)
	fmt.Println("自定义指令 0x6660 -> 0x6661", customReply.RespondSerialNumber, err, reply6661Msg.ExtensionFields.Err)
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
	for {
		if n, _ := conn.Read(data); n > 0 {
			jtMsg := jt808.NewJTMessage()
			_ = jtMsg.Decode(data[:n])
			seq := jtMsg.Header.SerialNumber
			jtMsg.Header.PlatformSerialNumber = seq

			switch consts.JT808CommandType(jtMsg.Header.ID) {
			case consts.P9003QueryTerminalAudioVideoProperties:
				tmp := model.T0x1003{
					EnterAudioEncoding: 100,
				}
				jtMsg.Header.ReplyID = uint16(tmp.Protocol())
				_, _ = conn.Write(jtMsg.Header.Encode(tmp.Encode()))

			case consts.P9101RealTimeAudioVideoRequest:
				tmp := model.T0x0001{
					SerialNumber: seq,
					ID:           jtMsg.Header.ID,
					Result:       1,
				}
				jtMsg.Header.ReplyID = uint16(tmp.Protocol())
				_, _ = conn.Write(jtMsg.Header.Encode(tmp.Encode()))

			case consts.P8104QueryTerminalParams:
				t0x0104 := make([]byte, 3)
				binary.BigEndian.PutUint16(t0x0104[0:2], seq)
				jtMsg.Header.ReplyID = uint16(consts.T0104QueryParameter)
				_, _ = conn.Write(jtMsg.Header.Encode(t0x0104))

			case consts.P8201QueryLocation:
				tmp := model.T0x0201{
					RespondSerialNumber: seq,
					T0x0200LocationItem: model.T0x0200LocationItem{
						AlarmSign:  1024,
						StatusSign: 2048,
						Latitude:   116307629,
						Longitude:  40058359,
						Altitude:   312,
						Speed:      3,
						Direction:  99,
						DateTime:   "2024-10-01 23:59:59",
					},
				}
				jtMsg.Header.ReplyID = uint16(tmp.Protocol())
				_, _ = conn.Write(jtMsg.Header.Encode(tmp.Encode()))

			case consts.P9205QueryResourceList:
				tmp := model.T0x1205{
					SerialNumber:            seq,
					AudioVideoResourceTotal: 1,
					AudioVideoResourceList: []model.T0x1205AudioVideoResource{
						{
							ChannelNo:              1,
							StartTime:              "2024-11-02 00:00:00",
							EndTime:                "2024-11-02 00:01:02",
							AlarmFlag:              1024,
							AudioVideoResourceType: 1,
							StreamType:             1,
							MemoryType:             1,
							FileSizeByte:           11,
						},
					},
				}
				jtMsg.Header.ReplyID = uint16(tmp.Protocol())
				_, _ = conn.Write(jtMsg.Header.Encode(tmp.Encode()))

			case consts.P9206FileUploadInstructions:
				tmp := model.T0x1206{
					RespondSerialNumber: seq,
					Result:              1,
				}
				jtMsg.Header.ReplyID = uint16(tmp.Protocol())
				_, _ = conn.Write(jtMsg.Header.Encode(tmp.Encode()))

			case consts.P8801CameraShootImmediateCommand:
				tmp := model.T0x0001{
					SerialNumber: seq,
					ID:           jtMsg.Header.ID,
					Result:       1,
				}
				jtMsg.Header.ReplyID = uint16(tmp.Protocol())
				_, _ = conn.Write(jtMsg.Header.Encode(tmp.Encode()))
				time.Sleep(1 * time.Second)

				{
					tmp := model.T0x0805{
						RespondSerialNumber: seq,
						Result:              1,
					}
					jtMsg.Header.ReplyID = uint16(tmp.Protocol())
					_, _ = conn.Write(jtMsg.Header.Encode(tmp.Encode()))
				}

			case consts.JT808CommandType(0x6660):
				tmp := CustomTerminalReply{
					RespondSerialNumber: seq,
				}
				jtMsg.Header.ReplyID = uint16(tmp.Protocol())
				_, _ = conn.Write(jtMsg.Header.Encode(tmp.Encode()))
			}
		}
	}
}
