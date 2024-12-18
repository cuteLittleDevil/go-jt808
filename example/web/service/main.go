package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cuteLittleDevil/go-jt808/attachment"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/hertz-contrib/pprof"
	"log/slog"
	"net/http"
	"os"
	"time"
	"web/internal/mq"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(logger)
	hlog.SetLevel(3)

	if conf.GetData().NatsConfig.Open {
		if err := mq.Init(conf.GetData().NatsConfig.Address); err != nil {
			panic(err)
		}
	}

	dirs := []string{
		conf.GetData().FileConfig.Dir,
		conf.GetData().JTConfig.CameraDir,
	}
	for _, dir := range dirs {
		_ = os.MkdirAll(dir, os.ModePerm)
	}

	go record.Run()

	{
		config := conf.GetData().FileConfig
		attach := attachment.New(
			attachment.WithNetwork("tcp"),
			attachment.WithHostPorts(config.Address),
			attachment.WithActiveSafetyType(consts.ActiveSafetyJS), // 默认苏标 支持黑标 广东标 湖南标 四川标
			attachment.WithFileEventerFunc(func() attachment.FileEventer {
				// 自定义文件处理 开始 结束 当前进度 补传 完成等事件
				return custom.NewFileEvent(config.Dir, config.LogFile)
			}),
		)
		go attach.Run()
	}
}

func main() {
	h := server.New(
		server.WithALPN(true),
		server.WithHostPorts(conf.GetData().ServerConfig.Address),
		server.WithHandleMethodNotAllowed(true),
	)
	pprof.Register(h)

	{
		config := conf.GetData().JTConfig
		goJt808 := service.New(
			service.WithHostPorts(config.Address),
			service.WithNetwork("tcp"),
			service.WithCustomTerminalEventer(func() service.TerminalEventer {
				// 自定义终端事件 终端进入 离开 读写报文事件
				return custom.NewTerminalEvent(config.ID)
			}),
			// 自定义key key和终端一一对应 默认使用手机号
			service.WithKeyFunc(func(msg *service.Message) (string, bool) {
				if msg.Command == consts.T0100Register {
					var register command.Register
					_ = register.Parse(msg.JTMessage)
					fmt.Println(register.String())
					return msg.JTMessage.Header.TerminalPhoneNo, true
				}
				return "", false
			}),
			service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
				authInfo := command.AuthInfo{Code: ""}
				return map[consts.JT808CommandType]service.Handler{
					consts.T0100Register: &command.Register{AuthInfo: &authInfo},
					// 如果没有注册过的终端鉴权拒绝 让他触发一次注册报文
					consts.T0102RegisterAuth:         &command.Auth{AuthInfo: &authInfo},
					consts.T0801MultimediaDataUpload: &command.Camera{Dir: config.CameraDir},
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
	h.Spin()
}

func appFS() *app.FS {
	_ = os.MkdirAll("./static/", os.ModePerm)
	return &app.FS{
		Root:        "./static/",
		PathRewrite: app.NewPathSlashesStripper(0),
		PathNotFound: func(_ context.Context, c *app.RequestContext) {
			type Response struct {
				Code int `json:"code"`
			}
			c.JSON(http.StatusOK, Response{
				Code: http.StatusNotFound,
			})
		},
		CacheDuration:        5 * time.Second,
		IndexNames:           []string{"*index.html"},
		Compress:             true,
		CompressedFileSuffix: "hertz",
		AcceptByteRange:      true,
	}
}
