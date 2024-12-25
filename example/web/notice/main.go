package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"github.com/hertz-contrib/cors"
	"github.com/hertz-contrib/websocket"
	"github.com/natefinch/lumberjack"
	"log/slog"
	"net/http"
	"time"
	"web/internal/mq"
	"web/internal/shared"
)

func init() {
	writeSyncer := &lumberjack.Logger{
		Filename:   "./notice.log",
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
	var (
		address  string
		natsAddr string
	)
	flag.StringVar(&address, "address", "0.0.0.0:18003", "server address")
	flag.StringVar(&natsAddr, "nats", "127.0.0.1:4222", "nats address")
	flag.Parse()

	if err := mq.Init(natsAddr); err != nil {
		slog.Error("nats init fail",
			slog.String("nats", natsAddr),
			slog.String("err", err.Error()))
		return
	}
	h := server.New(
		server.WithALPN(true),
		server.WithHostPorts(address),
		server.WithHandleMethodNotAllowed(true),
	)
	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           6 * time.Hour,
	}))
	h.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		fmt.Println(c.Request.URI().String(), string(c.Request.Body()))
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusNotFound,
			Msg:  "未找到的路由",
		})
	})
	group := h.Group("/api/v1/notice")
	group.GET("/", ws)
	group.POST("/parse", parse)
	h.Spin()
}

func ws(_ context.Context, c *app.RequestContext) {
	var upGrader = websocket.HertzUpgrader{}
	sim := c.DefaultQuery("sim", "")
	// 检查这个sim卡号有没有在线 -> 目前懒得弄 可以HTTP查询service服务
	err := upGrader.Upgrade(c, func(hc *websocket.Conn) {
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			_ = hc.Close()
		}()
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if _, _, err := hc.ReadMessage(); err != nil {
						cancel()
						return
					}
				}
			}
		}()
		ch, err := mq.Default().SubNotice(ctx, fmt.Sprintf("%s.*.%s.*", shared.WriteSubjectPrefix, sim))
		if err != nil {
			slog.Warn("sub notice fail",
				slog.String("sim", sim),
				slog.String("err", err.Error()))
			return
		}

		for v := range ch {
			var data shared.EventData
			if err := data.Parse(v); err == nil {
				ex := data.ExtensionFields
				source, target := ex.TerminalCommand, ex.PlatformCommand
				if ex.ActiveSend {
					source, target = target, source
				}
				if ex.SubcontractComplete {
					// 分包的情况 TerminalData 是body 重新组装回来
					data.JTMessage.Header.ReplyID = uint16(ex.TerminalCommand)
					ex.TerminalData = data.JTMessage.Header.Encode(ex.TerminalData)
				}
				notice := shared.Notice{
					Command:      ex.TerminalCommand,
					TerminalData: fmt.Sprintf("%x", ex.TerminalData),
					PlatformData: fmt.Sprintf("%x", ex.PlatformData),
					Remark:       fmt.Sprintf("%s -> %s", source, target),
				}
				b, _ := json.MarshalIndent(notice, "", "\t")
				if err := hc.WriteMessage(websocket.TextMessage, b); err != nil {
					return
				}
			}
		}
	})
	if err != nil {
		slog.Warn("websocket init fail",
			slog.String("err", err.Error()))
		return
	}
}

func parse(_ context.Context, c *app.RequestContext) {
	var notice shared.Notice
	if err := c.BindAndValidate(&notice); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	t := terminal.New()
	c.String(http.StatusOK, "[终端]%s \n\n[平台]%s",
		t.ProtocolDetails(notice.TerminalData),
		t.ProtocolDetails(notice.PlatformData))
}
