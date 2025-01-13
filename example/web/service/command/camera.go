package command

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"log/slog"
	"os"
	"time"
	"web/internal/file"
	"web/service/conf"
)

type Camera struct {
	dir       string
	saveLocal bool
	saveMinio bool
	model.T0x0801
}

func NewCamera() *Camera {
	cameraConfig := conf.GetData().FileConfig.CameraConfig
	return &Camera{
		dir:       cameraConfig.Dir,
		saveLocal: cameraConfig.Enable,
		saveMinio: cameraConfig.MinioConfig.Enable,
	}
}

func (c *Camera) OnReadExecutionEvent(msg *service.Message) {
	_ = c.T0x0801.Parse(msg.JTMessage)
	name := c.saveName(msg.Header.TerminalPhoneNo)
	if c.saveLocal {
		_ = os.WriteFile(c.dir+name, c.T0x0801.MultimediaPackage, os.ModePerm)
	}
	if c.saveMinio {
		date := time.Now().Format("20060102")
		path := fmt.Sprintf("/%s/%s", date, name)
		// 简单一点 把路径保存到txt中 也可以把name当key保存到redis 另一边获取路径
		minioUrl, err := file.Default().Upload(path, c.T0x0801.MultimediaPackage)
		if err != nil {
			slog.Warn("minio upload fail",
				slog.String("path", path),
				slog.String("err", err.Error()))
			return
		}
		_ = os.WriteFile(c.dir+name+".txt", []byte(minioUrl), os.ModePerm)
	}
}

func (c *Camera) OnWriteExecutionEvent(_ service.Message) {}

func (c *Camera) saveName(sim string) string {
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
	return fmt.Sprintf("%s_%d%s", sim, c.MultimediaID, format)
}
