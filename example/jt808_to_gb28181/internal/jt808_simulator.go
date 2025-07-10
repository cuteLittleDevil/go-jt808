package internal

import (
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"jt808_to_gb2818108/conf"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"
)

func Client(phone string, address string) {
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, phone))
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}
	defer func() {
		time.Sleep(3 * time.Second)
		_ = conn.Close()
	}()

	var (
		register = t.CreateDefaultCommandData(consts.T0100Register)
		auth     = t.CreateDefaultCommandData(consts.T0102RegisterAuth)
		location = t.CreateDefaultCommandData(consts.T0200LocationReport)
	)

	go func() {
		data := make([]byte, 1023)
		for {
			if n, _ := conn.Read(data); n > 0 {
				msg := fmt.Sprintf("%x", data[:n])
				if strings.HasPrefix(msg, "7e9101") {
					jtMsg := jt808.NewJTMessage()
					_ = jtMsg.Decode(data[:n])

					var play model.P0x9101
					_ = play.Parse(jtMsg)
					fmt.Println("模拟器收到的9101")
					fmt.Println(play.String())

					{
						t0x0001 := model.T0x0001{
							BaseHandle:   model.BaseHandle{},
							SerialNumber: jtMsg.Header.SerialNumber,
							ID:           jtMsg.Header.ID,
							Result:       0,
						}
						jtMsg.Header.ReplyID = uint16(t0x0001.Protocol())
						_, _ = conn.Write(jtMsg.Header.Encode(t0x0001.Encode()))
						fmt.Println(t0x0001.String())
					}

					go func() {
						jt1078Conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", play.ServerIPAddr, play.TcpPort))
						if err != nil {
							slog.Error("jt1078Conn",
								slog.String("sim", phone),
								slog.Any("err", err))
							return
						}
						time.Sleep(time.Second)
						dataPath := conf.GetData().Simulator.FilePath
						tmp, err := os.ReadFile(dataPath)
						if err != nil {
							slog.Error("jt1078Conn",
								slog.String("sim", phone),
								slog.String("dataPath", dataPath),
								slog.Any("err", err))
						}
						sendData, _ := hex.DecodeString(string(tmp))
						const groupSum = 1023
						count := 10
						for count > 0 {
							start := 0
							end := 0
							for i := 0; i < len(sendData)/groupSum; i++ {
								start = i * groupSum
								end = start + groupSum
								_, _ = jt1078Conn.Write(sendData[start:end])
								time.Sleep(20 * time.Millisecond)
							}
							_, _ = jt1078Conn.Write(sendData[end:])
							count--
						}
						fmt.Println("本次9101发送完成")
					}()
				}
			}
		}
	}()

	_, _ = conn.Write(register)
	time.Sleep(time.Second)
	_, _ = conn.Write(auth)

	ticker := time.NewTicker(5 * time.Second)
	sum := 3
	for range ticker.C {
		_, _ = conn.Write(location)
		sum--
		if sum == 0 {
			break
		}
	}
	time.Sleep(time.Second)
	time.Sleep(time.Hour)
	second := conf.GetData().Simulator.LeaveSecond
	if second < 0 {
		fmt.Println(fmt.Sprintf("[%s] 模拟设备停止发送经纬度信息 不退出", phone))
		select {}
	} else {
		fmt.Println(fmt.Sprintf("[%s] 模拟设备停止发送经纬度信息 在过[%d]秒模拟设备退出", phone, second))
		time.Sleep(time.Duration(second) * time.Second)
	}
}
