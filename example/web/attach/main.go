package main

import (
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cuteLittleDevil/go-jt808/attachment"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"gopkg.in/natefinch/lumberjack.v2"
	"log/slog"
	"os"
	"web/attach/conf"
	"web/attach/custom"
	"web/internal/file"
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

	if minio := conf.GetData().AttachConfig.MinioConfig; minio.Enable {
		if err := file.Init(minio.Endpoint, minio.AppKey, minio.AppSecret, minio.Bucket); err != nil {
			panic(err)
		}
	}

	dirs := []string{
		conf.GetData().AttachConfig.Dir,
	}
	for _, dir := range dirs {
		_ = os.MkdirAll(dir, os.ModePerm)
	}
}

func main() {
	attach := attachment.New(
		attachment.WithNetwork("tcp"),
		attachment.WithHostPorts(conf.GetData().Addr),
		attachment.WithActiveSafetyType(consts.ActiveSafetyJS), // 默认苏标 支持黑标 广东标 湖南标 四川标
		attachment.WithFileEventerFunc(func() attachment.FileEventer {
			// 自定义文件处理 开始 结束 当前进度 补传 完成等事件
			return custom.NewFileEvent()
		}),
	)
	attach.Run()
}
