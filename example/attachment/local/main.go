package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/attachment"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"io/fs"
	"math/rand/v2"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	address          string
	phone            string
	dir              string
	alarmID          string
	activeSafetyType consts.ActiveSafetyType
)

func main() {
	var asType int
	flag.StringVar(&address, "address", "0.0.0.0:17017", "主动安全服务地址")
	flag.StringVar(&phone, "phone", "1001", "测试的手机号")
	flag.StringVar(&dir, "dir", "../jt1078/data/", "要上传的文件目录")
	flag.StringVar(&alarmID, "alarmID", "2024-11-22_10_00_00_", "报警编号 上传的文件名称包含这个报警编号")
	// 6-暂不支持北京标
	flag.IntVar(&asType, "activeSafetyType", 1, "主动安全告警 1-苏标 2-黑标 3-广东标 4-湖南标 5-四川标 6-北京标")
	flag.Parse()
	activeSafetyType = consts.ActiveSafetyType(asType)
	//activeSafetyType = consts.ActiveSafetyBJ
	attach := attachment.New(
		attachment.WithNetwork("tcp"),
		attachment.WithHostPorts(address),
		attachment.WithActiveSafetyType(activeSafetyType),
		attachment.WithFileEventerFunc(func() attachment.FileEventer {
			f, _ := os.OpenFile("file.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
			return &meFileEvent{file: f}
		}),
	)
	go client(address)
	attach.Run()
}

type JT808Dataer interface {
	Protocol() consts.JT808CommandType
	Encode() []byte
}

func client(address string) {
	time.Sleep(time.Second) // 等待文件服务完全启动
	t := terminal.New(
		terminal.WithHeader(consts.JT808Protocol2013, phone),
	)
	t0x1211s := uploadFiles()
	fmt.Println("传输的文件数量", len(t0x1211s))
	t0x1210 := uploadNotice(t0x1211s)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	sendJT808Data(conn, t, &t0x1210)
	time.Sleep(5 * time.Second)
	for _, v := range t0x1211s {
		sendJT808Data(conn, t, &v)
		// 发送数据
		group := 10
		if datas := splitStreamData(v.FileName, group); len(datas) > 0 {
			num := rand.IntN(group)
			var (
				missOffset = 0
				missNum    = 0
			)
			offset := 0
			for i := 0; i < group; i++ {
				if i == num {
					missOffset = offset
					missNum = i
				} else {
					sendStreamData(conn, v.FileName, offset, datas[i])
					time.Sleep(time.Second) // 慢慢发送数据 可以打印进度
				}
				offset += len(datas[i])
			}
			// 发送完成
			t0x1212 := model.T0x1212{
				T0x1211: v,
			}
			sendJT808Data(conn, t, &t0x1212)
			if num >= 0 {
				// 发送补传的数据
				sendStreamData(conn, v.FileName, missOffset, datas[missNum])
				// 再次发送完成
				sendJT808Data(conn, t, &t0x1212)
			}
		}
	}
}

func splitStreamData(name string, group int) [][]byte {
	data, err := os.ReadFile(filepath.Join(dir, strings.ReplaceAll(name, alarmID, "")))
	if err != nil {
		panic(err)
	}
	if group <= 0 || group > len(data) {
		group = len(data)
	}

	result := make([][]byte, group)
	size, extra := len(data)/group, len(data)%group

	for i := range result {
		start := i*size + min(i, extra)
		end := start + size
		if i < extra {
			end++
		}
		result[i] = data[start:end]
	}
	return result
}

func sendStreamData(conn net.Conn, name string, offset int, data []byte) {
	_, _ = conn.Write([]byte{0x30, 0x31, 0x63, 0x64})
	if activeSafetyType == consts.ActiveSafetyHLJ {
		_, _ = conn.Write([]byte{byte(len(name))})
		_, _ = conn.Write([]byte(name))
	} else {
		_, _ = conn.Write(utils.String2FillingBytes(name, 50))
	}
	_, _ = conn.Write(binary.BigEndian.AppendUint32([]byte{}, uint32(offset)))
	_, _ = conn.Write(binary.BigEndian.AppendUint32([]byte{}, uint32(len(data))))
	_, _ = conn.Write(data)
}

func sendJT808Data(conn net.Conn, t *terminal.Terminal, jt808Data JT808Dataer) {
	data := t.CreateCommandData(jt808Data.Protocol(), jt808Data.Encode())
	_, _ = conn.Write(data)
	fmt.Println(fmt.Sprintf("%x", data))
	var buf [1023]byte
	_, _ = conn.Read(buf[:])
}

func uploadNotice(ts []model.T0x1211) model.T0x1210 {
	t0x1210 := model.T0x1210{
		TerminalID: "1234cd.",
		P9208AlarmSign: model.P9208AlarmSign{
			TerminalID:       "1234cd.",
			Time:             time.Now().Format(time.DateTime),
			SerialNumber:     1,
			AttachNumber:     byte(len(ts)),
			ActiveSafetyType: activeSafetyType,
		},
		AlarmID:              alarmID,
		InfoType:             0,
		AttachCount:          byte(len(ts)),
		T0x1210AlarmItemList: make([]model.T0x1210AlarmItem, 0, len(ts)),
	}
	for _, t0x1211 := range ts {
		t0x1210.T0x1210AlarmItemList = append(t0x1210.T0x1210AlarmItemList, model.T0x1210AlarmItem{
			FileNameLen: t0x1211.FileNameLen,
			FileName:    t0x1211.FileName,
			FileSize:    t0x1211.FileSize,
		})
	}
	return t0x1210
}

func uploadFiles() []model.T0x1211 {
	var list []model.T0x1211
	_ = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Size() == 0 {
			return nil
		}
		name := alarmID + info.Name()
		list = append(list, model.T0x1211{
			FileNameLen: byte(len(name)),
			FileName:    name,
			FileType:    0x04, // 全部用其他类型 方便点
			FileSize:    uint32(info.Size()),
		})
		return nil
	})
	return list
}
