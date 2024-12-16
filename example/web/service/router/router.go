package router

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"net/http"
	"time"
	"web/service/conf"
)

type (
	PlatformHandler interface {
		Encode() []byte
		Protocol() consts.JT808CommandType
		ReplyProtocol() consts.JT808CommandType
	}

	Request[T PlatformHandler] struct {
		Key     string                  `json:"key" binding:"required"`
		Command consts.JT808CommandType `json:"command" binding:"required"`
		Data    T                       `json:"data" binding:"required"`
	}

	Response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data any    `json:"data"`
	}
)

func Register(h *server.Hertz) {
	group := h.Group("/api/v1/jt808/")
	group.POST("/8103", p8103)
	group.POST("/8104", p8104)
	group.POST("/8801", p8801)
	group.POST("/9101", p9101)
	group.POST("/9102", p9102)
	group.POST("/9201", p9201)
	group.POST("/9202", p9202)
	group.POST("/9205", p9205)
	group.POST("/9206", p9206)
}

func p8801(_ context.Context, c *app.RequestContext) {
	var req Request[*model.P0x8801]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p9206(_ context.Context, c *app.RequestContext) {
	var req Request[*model.P0x9206]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p9205(_ context.Context, c *app.RequestContext) {
	var req Request[*model.P0x9205]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p9202(_ context.Context, c *app.RequestContext) {
	var req Request[*model.P0x9202]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p9201(_ context.Context, c *app.RequestContext) {
	var req Request[*model.P0x9201]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p9102(_ context.Context, c *app.RequestContext) {
	var req Request[*model.P0x9102]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p9101(_ context.Context, c *app.RequestContext) {
	var req Request[*model.P0x9101]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p8103(_ context.Context, c *app.RequestContext) {
	var req Request[*model.P0x8103]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p8104(_ context.Context, c *app.RequestContext) {
	var req Request[*model.P0x8104]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func handleCommand(c *app.RequestContext, key string, handle PlatformHandler) {
	if v, ok := c.Value(conf.GetData().JTConfig.ID).(*service.GoJT808); ok {
		replyMsg := v.SendActiveMessage(&service.ActiveMessage{
			Key:              key,
			Command:          handle.Protocol(),
			Body:             handle.Encode(),
			OverTimeDuration: 3 * time.Second,
		})
		if replyMsg.ExtensionFields.Err != nil {
			c.JSON(http.StatusOK, Response{
				Code: http.StatusInternalServerError,
				Msg:  replyMsg.ExtensionFields.Err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  "success",
			Data: replyParse(handle.Protocol(), handle.ReplyProtocol(), replyMsg),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: http.StatusInternalServerError,
		Msg:  "jt808服务不存在",
	})
}

func replyParse(command, replyCommand consts.JT808CommandType, msg *service.Message) any {
	type Reply struct {
		ErrDescribe     string                  `json:"errDescribe"`
		Command         consts.JT808CommandType `json:"command"`
		ReplyCommand    consts.JT808CommandType `json:"replyCommand"`
		Details         any                     `json:"details"`
		TerminalMessage string                  `json:"terminalMessage"`
		PlatformMessage string                  `json:"platformMessage"`
		Remark          string                  `json:"remark"`
	}
	reply := Reply{
		Command:      command,
		ReplyCommand: replyCommand,
		Remark:       fmt.Sprintf("%s -> %s", command.String(), replyCommand.String()),
	}
	type Handler interface {
		Parse(jtMsg *jt808.JTMessage) error
		Protocol() consts.JT808CommandType
	}
	var handle Handler
	switch replyCommand {
	case consts.T0001GeneralRespond:
		handle = &model.T0x0001{}
	case consts.T0104QueryParameter:
		handle = &model.T0x0104{}
	case consts.T1003UploadAudioVideoAttr:
		handle = &model.T0x1003{}
	case consts.T1205UploadAudioVideoResourceList:
		handle = &model.T0x1205{}
	case consts.T1206FileUploadCompleteNotice:
		handle = &model.T0x1206{}
	case consts.T0805CameraShootImmediately:
		handle = &model.T0x0805{}
	}
	if handle == nil {
		reply.ErrDescribe = fmt.Sprintf("暂未支持的命令 %s", command.String())
	} else {
		_ = handle.Parse(msg.JTMessage)
		reply.Details = handle
		reply.TerminalMessage = fmt.Sprintf("%x", msg.ExtensionFields.TerminalData)
		reply.PlatformMessage = fmt.Sprintf("%x", msg.ExtensionFields.PlatformData)
	}
	return reply
}
