package main

import (
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"github.com/jlaffaye/ftp"
	"net"
	"os"
	"strings"
	"time"
)

func client(phone string) {
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, phone))
	var (
		register = t.CreateDefaultCommandData(consts.T0100Register)
		auth     = t.CreateDefaultCommandData(consts.T0102RegisterAuth)
	)
	conn, err := net.Dial("tcp", "0.0.0.0:808")
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
				} else if strings.HasPrefix(msg, "7e9206") {
					sendData := t.CreateDefaultCommandData(consts.T1206FileUploadCompleteNotice)
					jtMsg := jt808.NewJTMessage()
					_ = jtMsg.Decode(sendData)
					var t0x1206 model.T0x1206
					_ = t0x1206.Parse(jtMsg)
					t0x1206.RespondSerialNumber = 4
					jtMsg.Header.ReplyID = uint16(t0x1206.Protocol())
					_, _ = conn.Write(jtMsg.Header.Encode(t0x1206.Encode()))
					go ftpClient(msg)
				}
			}
		}
	}()

	_, _ = conn.Write(auth)
	time.Sleep(100 * time.Millisecond)
	_, _ = conn.Write(register)
	time.Sleep(100 * time.Millisecond)
	_, _ = conn.Write(auth)
	select {}
}

func ftpClient(msg string) {
	jtMsg := jt808.NewJTMessage()
	data, _ := hex.DecodeString(msg)
	if err := jtMsg.Decode(data); err != nil {
		panic(err)
	}
	var p0x9206 model.P0x9206
	if err := p0x9206.Parse(jtMsg); err != nil {
		panic(err)
	}
	addr := fmt.Sprintf("%s:%d", p0x9206.FTPAddr, p0x9206.Port)
	fmt.Println("ftp连接地址", addr)
	c, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		panic(err)
	}

	err = c.Login(p0x9206.Username, p0x9206.Password)
	if err != nil {
		panic(err)
	}
	fmt.Println("ftp目录", baseFTPDir+p0x9206.FileUploadPath)
	if err := c.ChangeDir(p0x9206.FileUploadPath); err != nil {
		panic(err)
	}

	// 根据日期搜索文件 这里固定上传一个
	f, err := os.Open("../simulator/testdata/atop_cpu.png")
	if err != nil {
		panic(err)
	}
	err = c.Stor("atop_cpu.png", f)
	if err != nil {
		panic(err)
	}

	if err := c.Quit(); err != nil {
		panic(err)
	}
}
