package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/hertz-contrib/cors"
	"jt1078/help"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var goJt808 *service.GoJT808

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(logger)

	goJt808 = service.New(
		service.WithHostPorts("0.0.0.0:8081"),
		service.WithNetwork("tcp"),
		service.WithCustomTerminalEventer(func() service.TerminalEventer {
			return &help.LogTerminal{}
		}),
	)
	go goJt808.Run()
}

func main() {
	h := server.Default(
		server.WithHostPorts("0.0.0.0:9090"),
	)
	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	h.GET("/jt/startRealtimePlay", jt9101)
	h.GET("/zlm/startRealtimePlay", jt9101)
	h.GET("/jt/startBackPlay", jt9201)
	h.GET("/zlm/startBackPlay", jt9201)
	h.GET("/9101", jt9101)
	h.GET("/9201", jt9201)
	h.NoRoute(func(_ context.Context, c *app.RequestContext) {
		fmt.Println(string(c.Request.Method()), string(c.Request.Path()), string(c.Request.QueryString()))
		c.JSON(http.StatusNotFound, "")
	})
	h.Spin()
	// http://222.244.144.181:7777/video/1001-1-0-0.live.mp4
}

func jt9101(_ context.Context, c *app.RequestContext) {
	type Req9101 struct {
		IP         string `json:"ip,omitempty" query:"ip"`
		Port       int    `form:"port" json:"port,omitempty" query:"port"`
		ClientId   string `form:"clientId" json:"clientId,omitempty" query:"clientId"`
		ChannelNo  int    `form:"channelNo" json:"channelNo,omitempty" query:"channelNo"`
		MediaType  int    `form:"mediaType" json:"mediaType,omitempty" query:"mediaType"`
		StreamType int    `form:"streamType" json:"streamType,omitempty" query:"streamType"`
	}
	var req Req9101
	if err := c.BindAndValidate(&req); err != nil {
		slog.Warn("req fail",
			slog.Any("err", err))
		c.JSON(http.StatusOK, -1)
		return
	}
	//  222.244.144.181:7776
	if req.IP == "" {
		req.IP = "222.244.144.181"
	}
	if req.Port == 0 {
		req.Port = 7776
	}
	p9101 := &model.P0x9101{
		ServerIPLen:  byte(len(req.IP)),
		ServerIPAddr: req.IP,
		TcpPort:      uint16(req.Port),
		UdpPort:      0,
		ChannelNo:    byte(req.ChannelNo),
		DataType:     byte(req.MediaType),
		StreamType:   byte(req.StreamType),
	}
	replyMsg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              req.ClientId,
		Command:          p9101.Protocol(),
		Body:             p9101.Encode(),
		OverTimeDuration: 3 * time.Second,
	})
	extension := replyMsg.ExtensionFields
	fmt.Println(fmt.Sprintf("主动发送的9101 发送[%x] 应答[%x]",
		extension.TerminalSeq, extension.PlatformSeq))
	c.JSON(http.StatusOK, 1)
	return
}

func jt9201(_ context.Context, c *app.RequestContext) {
	type Req9201 struct {
		IP            string `json:"ip,omitempty" query:"ip"`
		Port          int    `form:"port" json:"port,omitempty" query:"port"`
		ClientId      string `form:"clientId" json:"clientId,omitempty" query:"clientId"`
		ChannelNo     int    `form:"channelNo" json:"channelNo,omitempty" query:"channelNo"`
		MediaType     int    `form:"mediaType" json:"mediaType,omitempty" query:"mediaType"`
		StreamType    int    `form:"streamType" json:"streamType,omitempty" query:"streamType"`
		PlayBackMode  int    `form:"playbackMode" json:"playbackMode,omitempty" query:"playbackMode"`
		PlayBackSpeed int    `form:"playbackSpeed" json:"playbackSpeed,omitempty" query:"playbackSpeed"`
		StartTime     string `form:"startTime" json:"startTime,omitempty" query:"startTime"`
		EndTime       string `form:"endTime" json:"endTime,omitempty" query:"endTime"`
	}
	var req Req9201
	if err := c.BindAndValidate(&req); err != nil {
		slog.Warn("req fail",
			slog.Any("err", err))
		c.JSON(http.StatusOK, -1)
		return
	}
	//  222.244.144.181:7776
	if req.IP == "" {
		req.IP = "222.244.144.181"
	}
	if req.Port == 0 {
		req.Port = 7776
	}
	p9201 := &model.P0x9201{
		ServerIPLen:  byte(len(req.IP)),
		ServerIPAddr: req.IP,
		TcpPort:      uint16(req.Port),
		UdpPort:      0,
		ChannelNo:    byte(req.ChannelNo),
		MediaType:    byte(req.MediaType),
		StreamType:   byte(req.StreamType),
		PlaybackWay:  byte(req.PlayBackMode),
		PlaySpeed:    byte(req.PlayBackSpeed),
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	}
	replyMsg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              req.ClientId,
		Command:          p9201.Protocol(),
		Body:             p9201.Encode(),
		OverTimeDuration: 3 * time.Second,
	})
	extension := replyMsg.ExtensionFields
	fmt.Println(fmt.Sprintf("主动发送的9102 发送[%x] 应答[%x]",
		extension.TerminalSeq, extension.PlatformSeq))
	c.JSON(http.StatusOK, 1)
	return
}
