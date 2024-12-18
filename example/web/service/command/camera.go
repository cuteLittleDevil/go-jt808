package command

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"os"
)

type Camera struct {
	Dir string
	model.T0x0801
}

func (c *Camera) OnReadExecutionEvent(msg *service.Message) {
	_ = c.T0x0801.Parse(msg.JTMessage)
	name := c.saveName(msg.Header.TerminalPhoneNo)
	_ = os.WriteFile(c.Dir+name, c.T0x0801.MultimediaPackage, os.ModePerm)
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
