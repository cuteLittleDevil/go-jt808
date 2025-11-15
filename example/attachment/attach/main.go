package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cuteLittleDevil/go-jt808/attachment"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/go-resty/resty/v2"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"log/slog"
)

func init() {
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	b, _ := json.MarshalIndent(viper.AllSettings(), "", "  ")
	fmt.Println(string(b))

	writeSyncer := &lumberjack.Logger{
		Filename:   "./app.log",
		MaxSize:    1,    // 单位是MB，日志文件最大为1MB
		MaxBackups: 3,    // 最多保留3个旧文件
		MaxAge:     28,   // 最大保存天数为28天
		Compress:   true, // 是否压缩旧文件
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(writeSyncer, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})))
}

func main() {
	attach := attachment.New(
		attachment.WithNetwork("tcp"),
		attachment.WithActiveSafetyType(consts.ActiveSafetyType(viper.GetInt("attach.type"))),
		attachment.WithHostPorts(viper.GetString("attach.addr")),
		attachment.WithFileEventerFunc(func() attachment.FileEventer {
			return &FileUploadEventHandler{
				eventURL:  viper.GetString("attach.onEventURL"),
				dirPrefix: viper.GetString("attach.dirPrefix"),
			}
		}),
	)
	attach.Run()
}

type FileUploadEventHandler struct {
	eventURL  string
	dirPrefix string
}

func (h *FileUploadEventHandler) OnEvent(progress *attachment.PackageProgress) {
	phoneNo := ""
	if msg := progress.ExtensionFields.RecentTerminalMessage; msg != nil && msg.Header != nil {
		phoneNo = msg.Header.TerminalPhoneNo
	}

	event := map[string]any{
		"phone":  phoneNo,
		"status": progress.ProgressStage,
		"remark": progress.ProgressStage.String(),
	}
	// 只有成功完成才保存文件并上报文件信息
	if progress.ProgressStage == attachment.ProgressStageSuccessQuit {
		saveDir := filepath.Join(h.dirPrefix, phoneNo)
		if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
			slog.Error("mkdir failed",
				slog.String("dir", saveDir),
				slog.Any("err", err))
		}
		type FileSaveResult struct {
			Name string `json:"name"`
			Path string `json:"path"`
			Size uint32 `json:"size"`
			Err  string `json:"err,omitempty"` // 只有出错时才输出
		}

		fileResults := make([]FileSaveResult, 0, len(progress.Record))
		for fileName, pack := range progress.Record {
			savePath := filepath.Join(saveDir, fileName)
			result := FileSaveResult{
				Name: fileName,
				Path: savePath,
				Size: pack.FileSize,
			}
			if err := os.WriteFile(savePath, pack.StreamBody, 0o644); err != nil {
				result.Err = err.Error()
				slog.Error("save file failed",
					slog.String("path", savePath),
					slog.Any("err", err))
			}
			fileResults = append(fileResults, result)
		}

		event["files"] = fileResults
	}

	if h.eventURL != "" {
		go h.reportEvent(event)
	}
}

func (h *FileUploadEventHandler) reportEvent(event map[string]any) {
	client := resty.New()
	client.SetDebug(false)
	client.SetTimeout(5 * time.Second)
	_, err := client.R().
		SetBody(event).
		ForceContentType("application/json; charset=utf-8").
		Post(h.eventURL)
	if err != nil {
		slog.Warn("report file event failed",
			slog.String("url", h.eventURL),
			slog.Any("event", event),
			slog.Any("error", err))
	}
}
