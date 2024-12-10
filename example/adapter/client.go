package main

import (
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

func client(phone string, address string) {
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, phone))
	var (
		register  = t.CreateDefaultCommandData(consts.T0100Register)
		auth      = t.CreateDefaultCommandData(consts.T0102RegisterAuth)
		heartBeat = t.CreateDefaultCommandData(consts.T0002HeartBeat)
	)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	f, _ := os.OpenFile("client.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	go func() {
		data := make([]byte, 1023)
		for {
			if n, _ := conn.Read(data); n > 0 {
				msg := fmt.Sprintf("%x", data[:n])
				_, _ = f.WriteString(msg + "\n")
				if strings.HasPrefix(msg, "7e9205") {
					sendData := t.CreateDefaultCommandData(consts.T1205UploadAudioVideoResourceList)
					jtMsg := jt808.NewJTMessage()
					_ = jtMsg.Decode(sendData)
					var t0x1205 model.T0x1205
					_ = t0x1205.Parse(jtMsg)
					t0x1205.SerialNumber = 3
					jtMsg.Header.ReplyID = uint16(t0x1205.Protocol())
					jt1205Data := jtMsg.Header.Encode(t0x1205.Encode())
					_, _ = conn.Write(jt1205Data)
					_, _ = f.WriteString(fmt.Sprintf("%x\n", jt1205Data))
				}
				_ = f.Sync()
			}
		}
	}()
	_, _ = conn.Write(auth)
	_, _ = f.WriteString(fmt.Sprintf("%x\n", auth))
	time.Sleep(time.Second)
	_, _ = conn.Write(register)
	_, _ = f.WriteString(fmt.Sprintf("%x\n", register))
	time.Sleep(time.Second)
	_, _ = conn.Write(auth)
	_, _ = f.WriteString(fmt.Sprintf("%x\n", auth))
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		_, _ = conn.Write(heartBeat)
		_, _ = f.WriteString(fmt.Sprintf("%x\n", heartBeat))
		_ = f.Sync()
	}
}
