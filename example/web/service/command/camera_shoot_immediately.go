package command

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"web/service/conf"
	"web/service/record"
)

type CameraShootImmediately struct {
	ImageURLs []string `json:"imageURLs"`
	MinioURLs []string `json:"minioURLs"`
	model.T0x0805
}

func (c *CameraShootImmediately) Parse(jtMsg *jt808.JTMessage) error {
	if err := c.T0x0805.Parse(jtMsg); err != nil {
		return err
	}
	sim := jtMsg.Header.TerminalPhoneNo
	cameraConfig := conf.GetData().FileConfig.CameraConfig
	if cameraConfig.Enable {
		c.localFiles(sim, cameraConfig)
	}
	if minio := cameraConfig.MinioConfig; minio.Enable {
		c.minioFiles(sim, cameraConfig.Dir)
	}
	return nil
}

func (c *CameraShootImmediately) localFiles(sim string, cameraConfig conf.CameraConfig) {
	for _, id := range c.T0x0805.MultimediaIDList {
		name := fmt.Sprintf("%s_%d", sim, id)
		_ = filepath.WalkDir(cameraConfig.Dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())) == name {
				savePath := cameraConfig.URLPrefix + d.Name()
				record.PutImageURL(sim, savePath)
				c.ImageURLs = append(c.ImageURLs, savePath)
				return filepath.SkipAll
			}
			return nil
		})
	}
}

func (c *CameraShootImmediately) minioFiles(sim string, dir string) {
	for _, id := range c.T0x0805.MultimediaIDList {
		name := fmt.Sprintf("%s_%d.txt", sim, id)
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if name == d.Name() {
				if data, err := os.ReadFile(path); err == nil {
					url := string(data)
					c.MinioURLs = append(c.MinioURLs, url)
					record.PutMinioURL(sim, url)
				} else {
					slog.Warn("read file fail",
						slog.String("path", path),
						slog.Any("err", err))
				}
				return filepath.SkipAll
			}
			return nil
		})
	}
}
