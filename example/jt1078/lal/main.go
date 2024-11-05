package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
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

var (
	_ip    = "127.0.0.1"
	_phone = "295696659617"
)

func main() {
	flag.StringVar(&_ip, "ip", "127.0.0.1", "Lal1078 ip地址")
	flag.StringVar(&_phone, "phone", "295696659617", "手机号")
	flag.Parse()

	goJt1078 := newLal1078("0.0.0.0:1078", "./conf/lalserver.conf.json")
	go goJt1078.run()

	goJt808 := service.New(
		service.WithHostPorts("0.0.0.0:808"),
		service.WithNetwork("tcp"),
	)
	go goJt808.Run()

	go func() {
		//quit := make(chan os.Signal)
		//signal.Notify(quit, syscall.SIGHUP) // kill -1
		//<-quit
		time.Sleep(3 * time.Second)
		p9101 := model.P0x9101{
			ServerIPLen:  byte(len(_ip)),
			ServerIPAddr: _ip,
			TcpPort:      1078,
			UdpPort:      0,
			ChannelNo:    1,
			DataType:     1,
			StreamType:   1,
		}
		body := p9101.Encode()
		fmt.Println(fmt.Sprintf(time.Now().Format(time.DateTime), "发送 9101指令 key=[%s]", _phone))
		activeMsg := service.NewActiveMessage(_phone, p9101.Protocol(), body, 3*time.Second)
		msg := goJt808.SendActiveMessage(activeMsg)
		var t0x0001 model.T0x0001
		if msg.WriteErr != nil {
			panic(msg.WriteErr)
		}
		if err := t0x0001.Parse(msg.JTMessage); err != nil {
			panic(err)
		}
		fmt.Println(time.Now().Format(time.DateTime), "终端的回复", t0x0001.String())
	}()

	time.Sleep(time.Second)
	client(_phone)
}

func client(phone string) {
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, phone))
	var (
		register  = t.CreateDefaultCommandData(consts.T0100Register)
		auth      = t.CreateDefaultCommandData(consts.T0102RegisterAuth)
		heartBeat = t.CreateDefaultCommandData(consts.T0002HeartBeat)
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
						jt1078Conn, err := net.Dial("tcp", fmt.Sprintf("%s:1078", _ip))
						if err != nil {
							panic(err)
						}
						time.Sleep(time.Second)
						tmp, err := os.ReadFile("../data/data.txt")
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
