package main

import (
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"net"
	"os"
	"strings"
	"time"
)

func client(port, videoPort int, phone string, dataPath string) {
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, phone))
	var (
		register  = t.CreateDefaultCommandData(consts.T0100Register)
		auth      = t.CreateDefaultCommandData(consts.T0102RegisterAuth)
		heartBeat = t.CreateDefaultCommandData(consts.T0002HeartBeat)
	)
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
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
				if strings.HasPrefix(msg, "7e9101") {
					jtMsg := jt808.NewJTMessage()
					_ = jtMsg.Decode(data[:n])
					t0x0001 := model.T0x0001{
						BaseHandle:   model.BaseHandle{},
						SerialNumber: jtMsg.Header.SerialNumber,
						ID:           jtMsg.Header.ID,
						Result:       0,
					}
					jtMsg.Header.ReplyID = uint16(t0x0001.Protocol())
					_, _ = conn.Write(jtMsg.Header.Encode(t0x0001.Encode()))
					go func() {
						jt1078Conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", videoPort))
						if err != nil {
							panic(err)
						}
						time.Sleep(time.Second)
						tmp, err := os.ReadFile(dataPath)
						if err != nil {
							panic(err)
						}
						sendData, _ := hex.DecodeString(string(tmp))
						const groupSum = 1023
						for {
							start := 0
							end := 0
							for i := 0; i < len(sendData)/groupSum; i++ {
								start = i * groupSum
								end = start + groupSum
								_, _ = jt1078Conn.Write(sendData[start:end])
								time.Sleep(20 * time.Millisecond)
							}
							_, _ = jt1078Conn.Write(sendData[end:])
						}
					}()
				}
			}
		}
	}()

	_, _ = conn.Write(auth)
	_, _ = conn.Write(register)
	_, _ = conn.Write(auth)

	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		_, _ = conn.Write(heartBeat)
	}
}
