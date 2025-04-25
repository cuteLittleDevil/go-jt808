package main

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/natefinch/lumberjack"
	"github.com/patrickmn/go-cache"
	_ "github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"jt1078/help"
	"log/slog"
	"time"
)

type Config struct {
	Service struct {
		Addr    string `yaml:"addr"`
		WebAddr string `yaml:"webAddr"`
	} `yaml:"service"`
	Zlm struct {
		Port               uint16 `yaml:"port"`
		Secret             string `yaml:"secret"`
		PlayURLFormat      string `yaml:"playURLFormat"`
		HttpsPlayURLFormat string `yaml:"httpsPlayURLFormat"`
		Intercom           struct {
			Url             string `yaml:"url"`
			Vhost           string `yaml:"vhost"`
			App             string `yaml:"app"`
			IntervalSecond  int    `yaml:"intervalSecond"`
			OvertimeSecond  int    `yaml:"overtimeSecond"`
			GetMediaInfoURL string `yaml:"getMediaInfoURL"`
		}
		OnStreamNotFound struct {
			IP string `yaml:"ip"`
		} `yaml:"onStreamNotFound"`
	} `yaml:"zlm"`
}

var GlobalConfig Config

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
	hlog.SetLevel(3)
}

func main() {
	goJt808 := service.New(
		service.WithHostPorts(GlobalConfig.Service.Addr),
		service.WithNetwork("tcp"),
		service.WithCustomTerminalEventer(func() service.TerminalEventer {
			return &help.LogTerminal{}
		}),
	)
	go goJt808.Run()

	h := server.Default(
		server.WithHostPorts(GlobalConfig.Service.WebAddr),
	)
	h.Use(func(_ context.Context, c *app.RequestContext) {
		c.Set("jt808", goJt808)
	})
	idCache := cache.New(1*time.Minute, 5*time.Minute)
	h.Use(func(_ context.Context, c *app.RequestContext) {
		c.Set("cache", idCache)
	})
	apiV1 := h.Group("/api/v1/")
	apiV1.POST("/9101", p9101)
	apiV1.POST("/on_stream_not_found", onStreamNotFound)
	apiV1.POST("/on_publish", onPublish)
	apiV1.POST("/start_send_rtp_talk", startSendRtpTalk)
	h.Spin()
}
