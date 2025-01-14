package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/hertz-contrib/cors"
	"github.com/hertz-contrib/pprof"
	"github.com/natefinch/lumberjack"
	"log/slog"
	"net/http"
	"os"
	"time"
	"web/internal/file"
	"web/internal/mq"
	"web/internal/shared"
	"web/service/command"
	"web/service/conf"
	"web/service/custom"
	"web/service/record"
	"web/service/router"
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
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	})
	slog.SetDefault(slog.New(handler))
	hlog.SetLevel(3)

	if conf.GetData().NatsConfig.Open {
		if err := mq.Init(conf.GetData().NatsConfig.Address); err != nil {
			panic(fmt.Sprintf("也可以选择关闭nats模式 启动失败: %v", err))
		}
	}

	if minio := conf.GetData().FileConfig.CameraConfig.MinioConfig; minio.Enable {
		if err := file.Init(minio.Endpoint, minio.AppKey, minio.AppSecret, minio.Bucket); err != nil {
			panic(err)
		}
	}

	dirs := []string{
		conf.GetData().FileConfig.CameraConfig.Dir,
	}
	for _, dir := range dirs {
		_ = os.MkdirAll(dir, os.ModePerm)
	}
	go record.Run()
}

func main() {
	h := server.New(
		server.WithALPN(true),
		server.WithHostPorts(conf.GetData().ServerConfig.Address),
		server.WithHandleMethodNotAllowed(true),
	)
	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           6 * time.Hour,
	}))
	pprof.Register(h)

	{
		config := conf.GetData().JTConfig
		goJt808 := service.New(
			service.WithHostPorts(config.Address),
			service.WithNetwork("tcp"),
			service.WithCustomTerminalEventer(func() service.TerminalEventer {
				// 自定义终端事件 终端进入 离开 读写报文事件
				return custom.NewTerminalEvent()
			}),
			// 自定义key key和终端一一对应 默认使用手机号
			service.WithKeyFunc(func(msg *service.Message) (string, bool) {
				if msg.Command == consts.T0100Register {
					var register command.Register
					_ = register.Parse(msg.JTMessage)
					fmt.Println("注册", register.String())
					return msg.JTMessage.Header.TerminalPhoneNo, true
				}
				if !conf.GetData().JTConfig.Verify { // 不校验的话 任意一个指令过来就加入
					return msg.JTMessage.Header.TerminalPhoneNo, true
				}
				return "", false
			}),
			service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
				verifyInfo := command.NewVerifyInfo()
				return map[consts.JT808CommandType]service.Handler{
					consts.T0100Register: command.NewRegister(verifyInfo),
					// 如果没有注册过的终端鉴权拒绝 让他触发一次注册报文
					consts.T0102RegisterAuth:         command.NewAuth(verifyInfo),
					consts.T0801MultimediaDataUpload: command.NewCamera(),
				}
			}),
		)
		go goJt808.Run()
		h.Use(func(c context.Context, ctx *app.RequestContext) {
			ctx.Set(config.ID, goJt808)
		})
	}

	router.Register(h)
	h.StaticFS("/", appFS())
	h.StaticFile("/index.html", "tstrtvs.html")
	h.Spin()
}

func appFS() *app.FS {
	_ = os.MkdirAll("./static/", os.ModePerm)
	return &app.FS{
		Root:        "./static/",
		PathRewrite: app.NewPathSlashesStripper(0),
		PathNotFound: func(_ context.Context, c *app.RequestContext) {
			c.JSON(http.StatusOK, shared.Response{
				Code: http.StatusNotFound,
				Msg:  "找不到路由",
				Data: string(c.Request.URI().Path()),
			})
		},
		CacheDuration:        5 * time.Second,
		IndexNames:           []string{"*index.html"},
		Compress:             true,
		CompressedFileSuffix: "hertz-jt808-web",
		AcceptByteRange:      true,
	}
}
