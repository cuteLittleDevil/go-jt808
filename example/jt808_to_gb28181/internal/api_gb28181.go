package internal

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/gb28181"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Device(c *gin.Context) {
	type Request struct {
		Sim string `json:"sim" binding:"required"`
	}
	var req Request
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response[string]{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, Response[gb28181.DeviceInfo]{
		Code: http.StatusOK,
		Msg:  "使用默认配置",
		Data: defaultDeviceInfo(req.Sim),
	})
}

func defaultDeviceInfo(sim string) gb28181.DeviceInfo {
	// 默认的设备id 就是3402000000132 + sim卡号最后6位 + 0
	sim = "000000" + sim
	return gb28181.DeviceInfo{
		ID:   fmt.Sprintf("3402000000132%s0", sim[len(sim)-6:]),
		IP:   "127.0.0.1",
		Port: 5060,
	}
}
