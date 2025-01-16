package command

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"log/slog"
	"os"
	"time"
	"web/internal/file"
	"web/internal/mq"
	"web/internal/shared"
	"web/service/conf"
	"web/service/record"
)

type Camera struct {
	dir       string
	minioDir  string
	saveLocal bool
	saveMinio bool
	model.T0x0801
	openNats   bool
	urlPrefix  string
	id         string
	httpPrefix string
}

func NewCamera() *Camera {
	cameraConfig := conf.GetData().FileConfig.CameraConfig
	return &Camera{
		dir:        cameraConfig.Dir,
		minioDir:   cameraConfig.MinioDir,
		saveLocal:  cameraConfig.Enable,
		saveMinio:  cameraConfig.MinioConfig.Enable,
		urlPrefix:  cameraConfig.URLPrefix,
		openNats:   conf.GetData().NatsConfig.Open,
		id:         conf.GetData().JTConfig.ID,
		httpPrefix: conf.GetData().JTConfig.HTTPPrefix,
	}
}

func (c *Camera) OnReadExecutionEvent(msg *service.Message) {
	if err := c.Parse(msg.JTMessage); err == nil {
		const timing = 1
		if c.T0x0801.EventItemEncode == timing {
			phone := msg.Header.TerminalPhoneNo
			now := time.Now().Format("150405")
			name := fmt.Sprintf("%s_%s_%d", now, phone, c.MultimediaID)
			go c.SaveData(name, msg.Key, phone)
		}
	}
}

func (c *Camera) OnWriteExecutionEvent(_ service.Message) {}

func (c *Camera) SaveData(name string, key string, phone string) {
	name = name + c.getDataType()
	data := shared.T0x0801File{
		LocalFileURL:        "",
		MinioURL:            "",
		Name:                name,
		ObjectName:          "",
		T0x0200LocationItem: c.T0x0200LocationItem,
		Phone:               phone,
	}

	if c.saveLocal {
		if err := os.WriteFile(c.dir+name, c.T0x0801.MultimediaPackage, os.ModePerm); err != nil {
			slog.Warn("local save fail",
				slog.String("path", c.dir+name),
				slog.String("err", err.Error()))
			return
		}
		data.LocalFileURL = c.urlPrefix + name
	}

	if c.saveMinio {
		date := time.Now().Format("20060102")
		objName := fmt.Sprintf("%s/%s", date, name)
		// 简单一点 把路径保存到txt中 也可以把name当key保存到redis 另一边获取路径
		minioUrl, err := file.Default().Upload(objName, c.T0x0801.MultimediaPackage)
		if err != nil {
			slog.Warn("minio upload fail",
				slog.String("name", objName),
				slog.String("err", err.Error()))
			return
		}
		if c.minioDir != "" {
			if err := os.WriteFile(c.minioDir+name+".txt", []byte(minioUrl), os.ModePerm); err != nil {
				slog.Warn("local save fail",
					slog.String("path", c.minioDir+name+".txt"),
					slog.String("err", err.Error()))
				return
			}
		}
		data.MinioURL = minioUrl
		data.ObjectName = objName
	}
	if c.openNats {
		c.pub(shared.NewEventData(shared.OnCustom, key,
			shared.WithIDAndHTTPPrefix(c.id, c.httpPrefix),
			shared.WithCustomData(phone, uint16(c.T0x0801.Protocol()), data)))
	} else {
		slog.Debug("pub custom",
			slog.String("0801", c.String()))
	}
	record.PutImageURL(phone, data.LocalFileURL, data.MinioURL)
}

func (c *Camera) getDataType() string {
	format := ".jpg"
	switch c.MultimediaFormatEncode {
	case 0:
		format = ".jpeg"
	case 1:
		format = ".tlf"
	case 2:
		format = ".mp3"
	case 3:
		format = ".wav"
	case 4:
		format = ".wmv"
	}
	return format
}

func (c *Camera) pub(data *shared.EventData) {
	sub := data.Subject
	if err := mq.Default().Pub(sub, data.ToBytes()); err != nil {
		slog.Error("pub fail",
			slog.String("sub", sub),
			slog.String("err", err.Error()))
	}
}
