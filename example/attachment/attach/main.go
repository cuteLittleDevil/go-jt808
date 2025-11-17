package main

import (
	"encoding/json"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/go-resty/resty/v2"
	"os"
	"path/filepath"
	"time"

	"github.com/cuteLittleDevil/go-jt808/attachment"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
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
		Filename:   viper.GetString("attach.log.name"),
		MaxSize:    viper.GetInt("attach.log.maxSize"),    // 单位是MB，日志文件最大为xMB
		MaxBackups: viper.GetInt("attach.log.maxBackups"), // 最多保留x个旧文件
		MaxAge:     viper.GetInt("attach.log.maxAge"),     // 最大保存天数为x天
		Compress:   viper.GetBool("attach.log.compress"),  // 是否压缩旧文件
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(writeSyncer, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.Level(viper.GetInt("attach.log.level")),
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "source" {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					// 只保留文件名部分
					a.Value = slog.AnyValue(filepath.Base(source.File))
				}
			}
			return a
		},
	})))
}

func main() {
	afType := consts.ActiveSafetyType(viper.GetInt("attach.type"))
	attach := attachment.New(
		attachment.WithNetwork("tcp"),
		attachment.WithActiveSafetyType(afType),
		attachment.WithHostPorts(viper.GetString("attach.addr")),
		attachment.WithFileEventerFunc(func() attachment.FileEventer {
			return &FileUploadEventHandler{
				eventURL:         viper.GetString("attach.onEventURL"),
				dirPrefix:        viper.GetString("attach.dirPrefix"),
				ActiveSafetyType: afType,
			}
		}),
	)
	attach.Run()
}

type FileUploadEventHandler struct {
	eventURL  string
	dirPrefix string
	alarmID   string
	consts.ActiveSafetyType
}

func (h *FileUploadEventHandler) OnEvent(progress *attachment.PackageProgress) {
	phoneNo := progress.ExtensionFields.TerminalPhoneNo

	event := map[string]any{
		"phone":   phoneNo,
		"status":  int(progress.ProgressStage),
		"remark":  progress.ProgressStage.String(),
		"alarmID": h.alarmID,
	}

	if debug := progress.ExtensionFields.Debug; debug.NetErr != nil {
		event["addr"] = debug.RemoteAddr
		event["netErr"] = debug.NetErr
		event["historyData"] = fmt.Sprintf("%x", debug.HistoryData)
	}

	if err := progress.ExtensionFields.Err; err != nil {
		slog.Error("on event",
			slog.String("phone", phoneNo),
			slog.Any("error", err))
		event["error"] = err.Error()
	}

	switch progress.ProgressStage {
	case attachment.ProgressStageInit:
		h.alarmID = h.getAlarmID(progress.ExtensionFields.RecentTerminalMessage)
		event["alarmID"] = h.alarmID
	case attachment.ProgressStageSuccessQuit:
		saveDir := filepath.Join(h.dirPrefix, phoneNo)
		event["files"] = h.handleFiles(saveDir, progress.Record)
	default:
	}

	slog.Debug("on event",
		slog.Any("event", event))
	if h.eventURL != "" {
		go h.reportEvent(event)
	}
}

func (h *FileUploadEventHandler) getAlarmID(msg *jt808.JTMessage) string {
	t0x1210 := model.T0x1210{
		P9208AlarmSign: model.P9208AlarmSign{
			ActiveSafetyType: h.ActiveSafetyType,
		},
	}
	if err := t0x1210.Parse(msg); err != nil {
		slog.Error("parse t0x1210 failed",
			slog.Any("error", err))
		return ""
	}
	return t0x1210.AlarmID
}

func (h *FileUploadEventHandler) handleFiles(saveDir string, record map[string]*attachment.Package) any {
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

	fileResults := make([]FileSaveResult, 0, len(record))
	for fileName, pack := range record {
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
	return fileResults
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
