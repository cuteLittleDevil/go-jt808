package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	var (
		localIP                  string
		address                  string
		maxPortRange             int
		heartBeatCycleSecond     int
		locationCycleSecond      int
		batchLocationCycleSecond int
	)
	flag.IntVar(&maxPortRange, "port", 10000, "本地可用端口数量")
	flag.StringVar(&localIP, "ip", "192.168.1.10", "本地ip")
	flag.StringVar(&address, "addr", "192.168.1.135:8080", "服务端ip")
	flag.IntVar(&heartBeatCycleSecond, "hc", 20, "心跳周期 默认20秒 小于等于0则不发送")
	flag.IntVar(&locationCycleSecond, "lc", 5, "定位上传周期 默认5秒 小于等于0则不发送")
	flag.IntVar(&batchLocationCycleSecond, "blc", 60, "定位批量上传周期 默认60秒 小于等于0则不发送")
	flag.Parse()

	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	sum := int32(0)
	for range ticker.C {
		go func() {
			defer func() {
				atomic.AddInt32(&sum, -1)
			}()
			localAddr, err := net.ResolveTCPAddr("tcp", localIP+":0") // 端口设置为0以让系统分配
			if err != nil {
				log.Fatal(err)
			}

			remoteAddr, err := net.ResolveTCPAddr("tcp", address)
			if err != nil {
				log.Fatal(err)
			}

			conn, err := net.DialTCP("tcp", localAddr, remoteAddr)
			if err != nil {
				return
			}
			defer func() {
				_ = conn.Close()
			}()
			var (
				register, _      = hex.DecodeString("7e0100002d0144199999990001000b0065373034343358485830303030320000000000000000000000006964303030303301d4c1413138383838927e")
				auth, _          = hex.DecodeString("7e0102000b01441999999900023134343139393939393939f67e")
				heartBeat, _     = hex.DecodeString("7e000200000144199999990015d27e")
				location, _      = hex.DecodeString("7e0200007c0123456789017fff000004000000080006eeb6ad02633df701380003006320070719235901040000000b02020016030200210402002c051e3737370000000000000000000000000000000000000000000000000000001105420000004212064d0000004d4d1307000000580058582504000000632a02000a2b040000001430011e3101286b7e")
				batchLocation, _ = hex.DecodeString("7e0702004e0123456789017fff022007211830000004d5c5c8fd313233343536373839303132330000000000000028d6d0bbaac8cbc3f1b9b2bacdb9fab5c0c2b7d4cbcae4b4d3d2b5c8cbd4b1b4d3d2b5d7cab8f1d6a420190630307e")
			)
			_, _ = conn.Write(register)
			time.Sleep(time.Second)
			_, _ = conn.Write(auth)

			stopChan := make(chan struct{})
			var once sync.Once
			go func() {
				data := make([]byte, 1023)
				for {
					if _, err := conn.Read(data); err != nil {
						once.Do(func() {
							close(stopChan)
						})
						return
					}
				}
			}()

			cTicker := time.NewTicker(1 * time.Second)
			defer cTicker.Stop()
			num := 0
			for {
				select {
				case <-stopChan:
					return
				case <-cTicker.C:
					num++
					var data [][]byte
					if heartBeatCycleSecond > 0 && num%heartBeatCycleSecond == 0 {
						data = append(data, heartBeat)
					}
					if locationCycleSecond > 0 && num%locationCycleSecond == 0 {
						data = append(data, location)
					}
					if batchLocationCycleSecond > 0 && num%batchLocationCycleSecond == 0 {
						data = append(data, batchLocation)
					}
					var sendData []byte
					for _, v := range data {
						sendData = append(sendData, v...)
					}
					if len(sendData) > 0 {
						if _, err := conn.Write(sendData); err != nil {
							once.Do(func() {
								close(stopChan)
							})
							return
						}
					}
				}
			}
		}()

		atomic.AddInt32(&sum, 1)
		fmt.Println(fmt.Sprintf("本地IP[%s] 可用端口数量[%d] 服务端IP[%s] 活跃连接[%d]",
			localIP, sum, address, sum))
		if sum >= int32(maxPortRange) {
			cTicker := time.NewTicker(1 * time.Second)
			for range cTicker.C {
				fmt.Println(fmt.Sprintf("本地IP[%s] 可用端口数量[%d] 服务端IP[%s] 活跃连接[%d]",
					localIP, sum, address, sum))
			}
		}
	}
}
