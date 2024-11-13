package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
)

var (
	src   string
	dst   string
	reply bool
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
	flag.StringVar(&src, "src", "0.0.0.0:8081", "接收的地址 默认8081端口")
	flag.StringVar(&dst, "dst", "127.0.0.1:8082", "变成客户端后 需要连接的服务器地址 默认127.0.0.1:8082")
	flag.BoolVar(&reply, "reply", true, "是否主动回复")
	flag.Parse()

	fmt.Println(fmt.Sprintf("本地[%s] 远端808服务端[%s] 是否回复[%v]", src, dst, reply))

	in, err := net.Listen("tcp", src)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := in.Accept()
		if err != nil {
			panic(err)
		}
		go run(conn)
	}
}

func run(conn net.Conn) {
	srcWriteChan := make(chan []byte, 100)
	var once sync.Once
	stopChan := make(chan struct{})
	defer func() {
		once.Do(func() {
			close(stopChan)
		})
	}()
	go func() {
		for {
			select {
			case <-stopChan:
				return
			case data := <-srcWriteChan:
				if reply {
					if n, err := conn.Write(data); err != nil {
						slog.Warn("write",
							slog.String("data", fmt.Sprintf("%x", data[:n])),
							slog.Any("err", err))
					}
				}
			}
		}
	}()

	curData := make([]byte, 10230)
	record := map[string]*connection{}
	pack := newPackageParse()
	for {
		select {
		case <-stopChan:
			return
		default:
		}
		if n, err := conn.Read(curData); err != nil {
			slog.Warn("conn",
				slog.Any("err", err))
			return
		} else if n > 0 {
			msgs, err := pack.unpack(curData[:n])
			if err != nil {
				slog.Warn("unpack",
					slog.Any("err", err))
				continue
			}
			for _, msg := range msgs {
				sim := msg.Header.TerminalPhoneNo
				v, ok := record[sim]
				// 1 不存在就新建一个客户端
				// 2 意外退出了 也新建一个
				if !ok || (ok && v.quit) {
					client := newConnection(sim, dst, srcWriteChan, stopChan)
					go client.run()
					record[sim] = client
				}
				record[sim].localWriteChan <- msg.OriginalData
			}
		}
	}
}
