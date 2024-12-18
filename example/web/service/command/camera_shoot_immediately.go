package command

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"os"
	"path/filepath"
	"strings"
	"web/service/conf"
	"web/service/record"
)

type CameraShootImmediately struct {
	ImageURLs []string `json:"imageURLs"`
	model.T0x0805
}

func (c *CameraShootImmediately) Parse(jtMsg *jt808.JTMessage) error {
	if err := c.T0x0805.Parse(jtMsg); err != nil {
		return err
	}
	sim := jtMsg.Header.TerminalPhoneNo
	for _, id := range c.T0x0805.MultimediaIDList {
		name := fmt.Sprintf("%s_%d", sim, id)
		_ = filepath.WalkDir(conf.GetData().JTConfig.CameraDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())) == name {
				savePath := conf.GetData().JTConfig.CameraURLPrefix + d.Name()
				record.PutImageURL(sim, savePath)
				c.ImageURLs = append(c.ImageURLs, savePath)
				return nil
			}
			return nil
		})
	}
	return nil
}
