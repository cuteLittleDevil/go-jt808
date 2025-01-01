package router

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"net/http"
	"time"
	"web/internal/shared"
	"web/service/command"
	"web/service/conf"
)

func p9208(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x9208]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p8801(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x8801]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p8201(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x8201]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p8202(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x8202]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p8300(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x8300]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p8302(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x8302]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p9206(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x9206]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p9205(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x9205]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p9202(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x9202]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p9201(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x9201]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p9102(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x9102]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p9101(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x9101]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p8103(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x8103]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func p8104(_ context.Context, c *app.RequestContext) {
	var req shared.Request[*model.P0x8104]
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}
	handleCommand(c, req.Key, req.Data)
}

func handleCommand(c *app.RequestContext, key string, handle shared.PlatformHandler) {
	if v, ok := c.Value(conf.GetData().JTConfig.ID).(*service.GoJT808); ok {
		replyMsg := v.SendActiveMessage(&service.ActiveMessage{
			Key:              key,
			Command:          handle.Protocol(),
			Body:             handle.Encode(),
			OverTimeDuration: 3 * time.Second,
		})
		if replyMsg.ExtensionFields.Err != nil {
			c.JSON(http.StatusOK, shared.Response{
				Code: http.StatusInternalServerError,
				Msg:  replyMsg.ExtensionFields.Err.Error(),
			})
			return
		}
		if replyMsg.Command != handle.ReplyProtocol() {
			slog.Warn("command",
				slog.String("reality", replyMsg.Command.String()),
				slog.String("expect", replyMsg.Command.String()))
		}
		c.JSON(http.StatusOK, shared.Response{
			Code: http.StatusOK,
			Msg:  "success",
			Data: replyParse(handle.Protocol(), replyMsg.Command, replyMsg),
		})
		return
	}

	c.JSON(http.StatusOK, shared.Response{
		Code: http.StatusInternalServerError,
		Msg:  "jt808服务不存在",
	})
}

func replyParse(commandType, replyCommandType consts.JT808CommandType, msg *service.Message) any {
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
		Command:      commandType,
		ReplyCommand: replyCommandType,
		Remark:       fmt.Sprintf("%s -> %s", commandType.String(), replyCommandType.String()),
	}
	type Handler interface {
		Parse(jtMsg *jt808.JTMessage) error
		Protocol() consts.JT808CommandType
	}
	var handle Handler
	switch replyCommandType {
	case consts.T0001GeneralRespond:
		handle = &model.T0x0001{}
	case consts.T0104QueryParameter:
		handle = &model.T0x0104{}
	case consts.T0201QueryLocation:
		handle = &model.T0x0201{}
	case consts.T1003UploadAudioVideoAttr:
		handle = &model.T0x1003{}
	case consts.T1205UploadAudioVideoResourceList:
		handle = &model.T0x1205{}
	case consts.T1206FileUploadCompleteNotice:
		handle = &model.T0x1206{}
	case consts.T0805CameraShootImmediately:
		handle = &command.CameraShootImmediately{}
	}
	if handle == nil {
		reply.ErrDescribe = fmt.Sprintf("暂未支持的命令 %s", commandType.String())
	} else {
		_ = handle.Parse(msg.JTMessage)
		reply.Details = handle
		reply.TerminalMessage = fmt.Sprintf("%x", msg.ExtensionFields.TerminalData)
		reply.PlatformMessage = fmt.Sprintf("%x", msg.ExtensionFields.PlatformData)
	}
	return reply
}
