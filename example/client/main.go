package main

import (
	"encoding/json"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"log"
	"log/slog"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"
)

type (
	Config struct {
		HTTPAddr string       `yaml:"httpAddr"`
		Server   ServerConfig `yaml:"server"`
		Client   ClientConfig `yaml:"client"`
	}

	ServerConfig struct {
		Addr string `yaml:"addr"` // 服务端地址，例如 101.35.2.3:808
	}

	ClientConfig struct {
		IP                  string    `yaml:"ip"`                  // 客户端IP
		IntervalMicrosecond int       `yaml:"intervalMicrosecond"` // 多久生成一个新客户端
		Sum                 int       `yaml:"sum"`                 // 测试的最大客户端数量
		Sim                 int       `yaml:"sim"`                 // 初始的sim卡号
		Version             int       `yaml:"version"`             // 客户端版本（2013 或 2019）
		Commands            []Command `yaml:"commands"`            // 按周期循环发送的指令列表
	}

	Command struct {
		Name           int  `yaml:"name"`           // 指令名称（如 0x0200、0x0002）
		Enable         bool `yaml:"enable"`         // 是否启用该指令
		IntervalSecond int  `yaml:"intervalSecond"` // 发送间隔，单位：秒
		Sum            int  `yaml:"sum"`            // 最多发送次数，0 表示不限制
	}
)

type (
	// key=sim卡号
	sessionOperationFunc func(record map[string]*Record)

	Record struct {
		Sim string `json:"sim"`
		Sum int    `json:"sum"`
		// key是指令名称 如终端-位置上报
		Commands   map[string]int `json:"commands"`
		CreateTime time.Time      `json:"createTime"`
		UpdateTime time.Time      `json:"updateTime"`
	}
)

var (
	GlobalConfig = &Config{}
	Manage       = &Manager{
		operationFuncChan: make(chan sessionOperationFunc, 100),
	}
)

func init() {
	v := viper.New()
	v.SetConfigFile("./config.yaml")
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(&GlobalConfig); err != nil {
		panic(err)
	}
	writeSyncer := &lumberjack.Logger{
		Filename:   "./app.log",
		MaxSize:    1,    // 单位是MB，日志文件最大为1MB
		MaxBackups: 3,    // 最多保留3个旧文件
		MaxAge:     28,   // 最大保存天数为28天
		Compress:   true, // 是否压缩旧文件
	}
	handler := slog.NewTextHandler(writeSyncer, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	})
	slog.SetDefault(slog.New(handler))

	b, _ := json.MarshalIndent(GlobalConfig, "", "  ")
	fmt.Println(string(b))
	go Manage.Run()
}

func main() {
	go httpApi(GlobalConfig.HTTPAddr)
	var wg sync.WaitGroup
	for i := 0; i < GlobalConfig.Client.Sum; i++ {
		wg.Add(1)
		sim := fmt.Sprintf("%d", GlobalConfig.Client.Sim+i)
		go func(sim string) {
			defer wg.Done()
			// 实际目前永远不会结束
			client(sim)
		}(sim)
		time.Sleep(time.Duration(GlobalConfig.Client.IntervalMicrosecond) * time.Microsecond)
	}
	wg.Wait()
}

func client(phone string) {
	localAddr, err := net.ResolveTCPAddr("tcp", GlobalConfig.Client.IP+":0") // 端口设置为0以让系统分配
	if err != nil {
		slog.Warn("invalid local address",
			slog.String("server", GlobalConfig.Server.Addr),
			slog.Any("error", err))
		return
	}

	remoteAddr, err := net.ResolveTCPAddr("tcp", GlobalConfig.Server.Addr)
	if err != nil {
		slog.Warn("invalid server address",
			slog.String("server", GlobalConfig.Server.Addr),
			slog.Any("client", localAddr),
			slog.Any("error", err))
		return
	}

	conn, err := net.DialTCP("tcp", localAddr, remoteAddr)
	if err != nil {
		slog.Warn("conn fail",
			slog.String("server", GlobalConfig.Server.Addr),
			slog.Any("client", localAddr),
			slog.Any("error", err))
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	version := consts.JT808Protocol2013
	if GlobalConfig.Client.Version == 2019 {
		version = consts.JT808Protocol2019
	}
	t := terminal.New(terminal.WithHeader(version, phone))
	t.CreateCustomMessageFunc = func(commandType consts.JT808CommandType) (terminal.Handler, bool) {
		if commandType == consts.T0200LocationReport {
			return &model.T0x0200{
				T0x0200LocationItem: model.T0x0200LocationItem{
					AlarmSign:  1024,
					StatusSign: 2048,
					Latitude:   116307629,
					Longitude:  40058359,
					Altitude:   312,
					Speed:      3,
					Direction:  99,
					DateTime:   time.Now().Format(time.DateTime),
				},
			}, true
		}
		return nil, false
	}
	register := t.CreateDefaultCommandData(consts.T0100Register)
	_, _ = conn.Write(register)
	data := make([]byte, 1023)
	if n, err := conn.Read(data); err == nil {
		var jtMsg = jt808.NewJTMessage()
		if err := jtMsg.Decode(data[:n]); err == nil {
			var p08100 model.P0x8100
			_ = p08100.Parse(jtMsg)
			t0x0102 := model.T0x0102{
				BaseHandle:      model.BaseHandle{},
				AuthCodeLen:     uint8(len(p08100.AuthCode)),
				AuthCode:        p08100.AuthCode,
				TerminalIMEI:    "123456789012345",
				SoftwareVersion: "3.7.15",
				Version:         version,
			}
			auth := t.CreateCommandData(consts.T0102RegisterAuth, t0x0102.Encode())
			_, _ = conn.Write(auth)
		}
	}

	Manage.Add(phone)
	writeChan := make(chan []byte, 10)
	for _, v := range GlobalConfig.Client.Commands {
		if v.Enable {
			go func(command Command) {
				ticker := time.NewTicker(time.Duration(command.IntervalSecond) * time.Second)
				defer ticker.Stop()
				count := 0
				for range ticker.C {
					count++
					sendData := t.CreateDefaultCommandData(consts.JT808CommandType(command.Name))
					if len(sendData) == 0 {
						slog.Warn("invalid command",
							slog.Any("command", command.Name))
					} else {
						writeChan <- sendData
					}
					if command.Sum != 0 && count == command.Sum {
						return
					}
					tmp := fmt.Sprintf("%04x:%s", command.Name, consts.JT808CommandType(command.Name))
					Manage.Update(phone, tmp)
				}
			}(v)
		}
	}
	for sendData := range writeChan {
		_, _ = conn.Write(sendData)
	}
}

type Manager struct {
	operationFuncChan chan sessionOperationFunc
}

func (m *Manager) Run() {
	records := make(map[string]*Record)
	for op := range m.operationFuncChan {
		op(records)
	}
}

func (m *Manager) Add(sim string) {
	ch := make(chan struct{})
	m.operationFuncChan <- func(records map[string]*Record) {
		defer close(ch)
		records[sim] = &Record{
			Sim:        sim,
			Commands:   map[string]int{},
			CreateTime: time.Now(),
		}
	}
	<-ch
}

func (m *Manager) Update(sim string, command string) {
	ch := make(chan struct{})
	m.operationFuncChan <- func(records map[string]*Record) {
		defer close(ch)
		if v, ok := records[sim]; ok {
			v.Sum++
			v.Commands[command]++
			v.UpdateTime = time.Now()
			records[sim] = v
		}
	}
	<-ch
}

func (m *Manager) ShowAll(sim string) any {
	reply := make([]Record, 0, 100)
	ch := make(chan int, 1)
	m.operationFuncChan <- func(records map[string]*Record) {
		defer close(ch)
		allSum := 0
		if sim == "" {
			for _, v := range records {
				reply = append(reply, *v)
				allSum += v.Sum
			}
		} else if v, ok := records[sim]; ok {
			reply = append(reply, *v)
			allSum = v.Sum
		}
		ch <- allSum
	}
	allSum := <-ch
	type Result struct {
		AllSum  int      `json:"allSum"`
		Records []Record `json:"records"`
	}
	sort.Slice(reply, func(i, j int) bool {
		return reply[i].CreateTime.After(reply[j].CreateTime)
	})
	return Result{
		AllSum:  allSum,
		Records: reply,
	}
}

func httpApi(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/all", func(w http.ResponseWriter, r *http.Request) {
		sim := r.URL.Query().Get("sim")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(Manage.ShowAll(sim)); err != nil {
			http.Error(w, "encode json failed", http.StatusInternalServerError)
			return
		}
	})
	api := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	log.Fatal(api.ListenAndServe())
}
