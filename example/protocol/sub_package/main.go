package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"io"
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
				consts.T1205UploadAudioVideoResourceList: &t1205{},
				consts.P8003ReissueSubcontractingRequest: &p8003{},
			}
		}),
		//service.WithHasFilterSubPack(false),
	)
	go goJt808.Run()

	go func() {
		time.Sleep(1 * time.Second)
		phone := "17299841738"
		client(phone)
	}()
	time.Sleep(100 * time.Second)
}

func client(phone string) {
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, phone))
	var (
		register  = t.CreateDefaultCommandData(consts.T0100Register)
		auth      = t.CreateDefaultCommandData(consts.T0102RegisterAuth)
		heartBeat = t.CreateDefaultCommandData(consts.T0002HeartBeat)
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
			_, _ = conn.Read(data)
		}
	}()

	go func() {
		_, _ = conn.Write(auth)
		time.Sleep(time.Second)
		_, _ = conn.Write(register)
		time.Sleep(time.Second)
		_, _ = conn.Write(auth)
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			_, _ = conn.Write(heartBeat)
		}
	}()

	sendFunc := func(name string) {
		file, err := os.Open(name)
		if err != nil {
			panic(err)
		}
		reader := bufio.NewReader(file)
		for {
			line, _, err := reader.ReadLine()
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
			data, _ := hex.DecodeString(string(line))
			_, _ = conn.Write(data)
		}
		_ = file.Close()
	}
	// 1 发送完整的数据 -> 等待直接完成
	sendFunc("./sub_package/0x1205.txt")
	time.Sleep(10 * time.Second)

	// 2 发送不完整的数据 过一段时间在发送缺失补传 -> 等待过一段时间完成
	sendFunc("./sub_package/0x1205缺失.txt")
	time.Sleep(10 * time.Second)
	sendFunc("./sub_package/0x1205缺失补传.txt")
	time.Sleep(15 * time.Second)

	// 3 发送不完整的数据 -> 等待完不成后删除
	sendFunc("./sub_package/0x1205缺失.txt")
	select {}
}

type t1205 struct {
	model.T0x1205
}

func (t *t1205) OnReadExecutionEvent(message *service.Message) {
	fmt.Println("接收的数据 0x1205", len(message.Body))
	var tmp model.T0x1205
	_ = tmp.Parse(message.JTMessage)
	fmt.Println(tmp.String())
}

func (t *t1205) OnWriteExecutionEvent(_ service.Message) {}

type p8003 struct {
	model.P0x8003
}

func (p *p8003) OnReadExecutionEvent(_ *service.Message) {}

func (p *p8003) OnWriteExecutionEvent(message service.Message) {
	fmt.Println("补包请求", fmt.Sprintf("%x", message.ExtensionFields.PlatformData), time.Now().Format(time.DateTime))
}
