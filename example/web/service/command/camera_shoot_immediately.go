package command

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
)

type CameraShootImmediately struct {
	model.T0x0805
	CustomData any      `json:"customData"`
	Names      []string `json:"names"`
}

func (c *CameraShootImmediately) Parse(jtMsg *jt808.JTMessage) error {
	if infos, ok := c.CustomData.(map[uint32]string); ok {
		for _, v := range infos {
			c.Names = append(c.Names, v)
		}
	}
	return c.T0x0805.Parse(jtMsg)
}
