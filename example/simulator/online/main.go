package main

import (
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	var (
		localIP                  string
		address                  string
		phoneStart               int
		version                  int
		maxPortRange             int
		heartBeatCycleSecond     int
		locationCycleSecond      int
		batchLocationCycleSecond int
	)
	flag.IntVar(&maxPortRange, "port", 10000, "本地可用端口数量")
	flag.StringVar(&localIP, "ip", "192.168.1.10", "本地ip")
	flag.IntVar(&phoneStart, "start", 1, "手机号开始的数字")
	flag.IntVar(&version, "version", 2013, "协议版本 默认2013")
	flag.StringVar(&address, "addr", "127.0.0.1:8080", "服务端ip")
	flag.IntVar(&heartBeatCycleSecond, "hc", 20, "心跳周期 默认20秒 小于等于0则不发送")
	flag.IntVar(&locationCycleSecond, "lc", 5, "定位上传周期 默认5秒 小于等于0则不发送")
	flag.IntVar(&batchLocationCycleSecond, "blc", 60, "定位批量上传周期 默认60秒 小于等于0则不发送")
	flag.Parse()

	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	sum := int32(0)
	count := int32(1)
	for range ticker.C {
		go func(phoneStart int) {
			defer func() {
				atomic.AddInt32(&sum, -1)
			}()
			fmt.Println(fmt.Sprintf("本地IP[%s] 可用端口数量[%d] 服务端IP[%s] 活跃连接[%d] 创建手机号[%d]",
				localIP, maxPortRange, address, sum, phoneStart))
			localAddr, err := net.ResolveTCPAddr("tcp", localIP+":0") // 端口设置为0以让系统分配
			if err != nil {
				return
			}

			remoteAddr, err := net.ResolveTCPAddr("tcp", address)
			if err != nil {
				return
			}

			conn, err := net.DialTCP("tcp", localAddr, remoteAddr)
			if err != nil {
				return
			}
			defer func() {
				_ = conn.Close()
			}()
			protocolVersion := consts.JT808Protocol2013
			switch version {
			case 2011:
				protocolVersion = consts.JT808Protocol2011
			case 2019:
				protocolVersion = consts.JT808Protocol2019
			}
			t := terminal.New(terminal.WithHeader(protocolVersion, fmt.Sprintf("%d", phoneStart)))
			var (
				register      = t.CreateDefaultCommandData(consts.T0100Register)
				auth          = t.CreateDefaultCommandData(consts.T0102RegisterAuth)
				heartBeat     = t.CreateDefaultCommandData(consts.T0002HeartBeat)
				location      = t.CreateDefaultCommandData(consts.T0200LocationReport)
				batchLocation = t.CreateDefaultCommandData(consts.T0704LocationBatchUpload)
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
		}(phoneStart)
		phoneStart++
		atomic.AddInt32(&sum, 1)
		count++
		if count > int32(maxPortRange) {
			cTicker := time.NewTicker(1 * time.Second)
			for range cTicker.C {
				fmt.Println(fmt.Sprintf("本地IP[%s] 可用端口数量[%d] 活跃连接[%d]", localIP, maxPortRange, sum))
			}
		}
	}
}
