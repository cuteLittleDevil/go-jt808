package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/adapter"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"jt808_to_gb2818108/conf"
	"jt808_to_gb2818108/internal"
	"log/slog"
	"path/filepath"
	"time"
)

func init() {
	if err := conf.InitConfig("./config.yaml"); err != nil {
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
		AddSource: true,
		Level:     slog.LevelInfo,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "source" {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					// 只保留文件名部分
					a.Value = slog.AnyValue(filepath.Base(source.File))
				}
			}
			return a
		},
	})
	slog.SetDefault(slog.New(handler))
}

func main() {
	if config := conf.GetData().Adapter; config.Enable {
		terminals := make([]adapter.Terminal, 0, 2)
		terminals = append(terminals, adapter.Terminal{
			Mode:       adapter.Leader, // 服务和设备之间读写全部正常
			TargetAddr: config.Leader,  // 原项目的jt808服务
		})
		for _, follower := range config.Followers {
			terminals = append(terminals, adapter.Terminal{
				Mode:          adapter.Follower, // 服务读正常 写默认拒绝（只下发指定命令）
				TargetAddr:    follower.Address, // go-jt808项目的jt808服务
				AllowCommands: follower.AllowCommands,
			})
		}
		if config.RetrySecond < 5 {
			config.RetrySecond = 5
		}
		second := time.Duration(config.RetrySecond) * time.Second
		adapterGroup := adapter.New(
			adapter.WithHostPorts(config.Address),
			//adapter.WithAllowCommand( // 全局都允许的向设备写的命令
			//	consts.P9101RealTimeAudioVideoRequest,
			//),
			adapter.WithTimeoutRetry(second), // 模拟连接断开后 多久重试一次
			adapter.WithTerminals(terminals...),
		)
		go adapterGroup.Run()
		time.Sleep(time.Second)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, "+
			"Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	{
		goJt808 := service.New(
			service.WithHostPorts(conf.GetData().JT808.Address),
			service.WithCustomTerminalEventer(func() service.TerminalEventer {
				return internal.NewAdapterTerminal(conf.GetData().JT808.HasDetails)
			}),
		)
		go goJt808.Run()
		r.Use(func(c *gin.Context) {
			c.Set("jt808", goJt808)
		})
	}

	{
		group := r.Group("/api/v1/jt808/")
		group.POST("/9003", internal.P9003)
		group.POST("/9101", internal.P9101)
		group.POST("/9102", internal.P9102)
		group.POST("/9201", internal.P9201)
		group.POST("/9202", internal.P9202)
		group.POST("/9205", internal.P9205)
		group.POST("/9206", internal.P9206)
		group.POST("/9208", internal.P9208)
	}

	{
		group := r.Group("/api/v1/jt808/gb28181/")
		group.POST("/device", internal.Device)
	}

	if simulator := conf.GetData().Simulator; simulator.Enable {
		time.Sleep(1 * time.Second)
		fmt.Println("使用模拟链接")
		go internal.Client(simulator.Sim, conf.GetData().Simulator.Address) // 模拟一个设备连接
	}

	_ = r.Run(conf.GetData().JT808.ApiAddress)
}
