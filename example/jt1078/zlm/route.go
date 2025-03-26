package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type (
	PlatformHandler interface {
		Encode() []byte
		Protocol() consts.JT808CommandType
		ReplyProtocol() consts.JT808CommandType
	}

	Request[T PlatformHandler] struct {
		Key  string `json:"key" binding:"required"`
		Data T      `json:"data" binding:"required"`
	}

	Response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data any    `json:"data"`
	}
)

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
	id := fmt.Sprintf("%s-%d-%d", req.Key, req.Data.ChannelNo, req.Data.TcpPort)
	// 先关闭zlm的收流端口
	if err := closeRtpServer(GlobalConfig.Zlm.CloseURL, map[string]string{
		"secret":    GlobalConfig.Zlm.Secret,
		"stream_id": id,
	}); err != nil {
		slog.Warn("关闭zlm收流端口失败",
			slog.String("id", id),
			slog.String("err", err.Error()))
	}
	// 打开zlm的收流端口
	if err := openRtpServer(GlobalConfig.Zlm.OpenURL, map[string]string{
		"port":      strconv.Itoa(int(req.Data.TcpPort)),
		"secret":    GlobalConfig.Zlm.Secret,
		"tcp_mode":  "1", // 0:udp 1:tcp被动 2:tcp主动
		"stream_id": id,
	}); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "打开zlm收流端口失败",
			Data: err.Error(),
		})
		return
	}

	// 发送指令
	if err := handleCommand(c, req.Key, req.Data); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "发送指令失败",
			Data: err.Error(),
		})
		return
	}

	// zlm播放规则 https://github.com/zlmediakit/ZLMediaKit/wiki/%E6%92%AD%E6%94%BEurl%E8%A7%84%E5%88%99
	type Result struct {
		StreamID string `json:"streamID"`
		MP4      string `json:"mp4"`
	}
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  "成功",
		Data: Result{
			StreamID: id,
			MP4:      fmt.Sprintf(GlobalConfig.Zlm.PlayURLFormat, id),
		},
	})
}

func handleCommand(c *app.RequestContext, key string, handle PlatformHandler) error {
	if v, ok := c.Value("jt808").(*service.GoJT808); ok {
		replyMsg := v.SendActiveMessage(&service.ActiveMessage{
			Key:              key,
			Command:          handle.Protocol(),
			Body:             handle.Encode(),
			OverTimeDuration: 3 * time.Second,
		})
		if replyMsg.ExtensionFields.Err != nil {
			return replyMsg.ExtensionFields.Err
		}
		if replyMsg.Command != handle.ReplyProtocol() {
			return fmt.Errorf("command not in conform %s-%s",
				replyMsg.Command.String(), handle.ReplyProtocol().String())
		}
		return nil
	}

	return fmt.Errorf("jt808 not found")
}
