package main

import (
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"log/slog"
	"net"
	"os"
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
		singleLimitTotal         int
		logLevel                 int
	)
	flag.IntVar(&maxPortRange, "max", 10000, "本地可用端口数量")
	flag.StringVar(&localIP, "ip", "192.168.1.10", "本地ip")
	flag.IntVar(&phoneStart, "start", 1, "手机号开始的数字")
	flag.IntVar(&version, "version", 2013, "协议版本 默认2013")
	flag.StringVar(&address, "addr", "127.0.0.1:8080", "服务端ip")
	flag.IntVar(&heartBeatCycleSecond, "hc", 20, "心跳周期 默认20秒 小于等于0则不发送")
	flag.IntVar(&locationCycleSecond, "lc", 5, "定位上传周期 默认5秒 小于等于0则不发送")
	flag.IntVar(&batchLocationCycleSecond, "blc", 60, "定位批量上传周期 默认60秒 小于等于0则不发送")
	flag.IntVar(&singleLimitTotal, "limit", 0, "每一个模拟终端最大发送有效包数量[0x0200 0x0704] 小于等于0无限")
	flag.IntVar(&logLevel, "lv", -4, "slog 的等级 -4=debug 0=info 4=warn 8=err")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.Level(logLevel),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				formattedTime := a.Value.Time().Format(time.DateTime)
				return slog.String(slog.TimeKey, formattedTime)
			}
			return a
		},
	}))
	slog.SetDefault(logger)

	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	sum := int32(0)
	count := int32(1)
	for range ticker.C {
		go func(phoneStart int) {
			defer func() {
				atomic.AddInt32(&sum, -1)
			}()
			slog.Debug("start",
				slog.String("ip", localIP),
				slog.String("addr", address),
				slog.Int("conn sum", int(sum)),
				slog.Int("phone", phoneStart))
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
			total := 0
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
						total++
					}
					if batchLocationCycleSecond > 0 && num%batchLocationCycleSecond == 0 {
						data = append(data, batchLocation)
						total++
					}
					// 合在一起 更好模拟数据粘包的场景
					var sendData []byte
					for _, v := range data {
						sendData = append(sendData, v...)
					}
					if len(sendData) > 0 {
						if _, err := conn.Write(sendData); err != nil {
							slog.Warn("write",
								slog.Int("phone", phoneStart),
								slog.Any("err", err))
							once.Do(func() {
								close(stopChan)
							})
							return
						}
					}
					if singleLimitTotal > 0 && total == singleLimitTotal {
						slog.Debug("complete",
							slog.Int("phone", phoneStart),
							slog.Int("total", total))
						// 等待一下 让数据肯定发送完
						time.Sleep(3 * time.Second)
						once.Do(func() {
							close(stopChan)
						})
						return
					}
				}
			}
		}(phoneStart)
		phoneStart++
		atomic.AddInt32(&sum, 1)
		count++
		if count > int32(maxPortRange) {
			cTicker := time.NewTicker(10 * time.Second)
			for range cTicker.C {
				slog.Info("detection",
					slog.String("ip", localIP),
					slog.Int("max", maxPortRange),
					slog.Int("active sum", int(sum)))
			}
		}
	}
}
