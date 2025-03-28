package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/spf13/cast"
	"net/http"
	"strings"
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
		Sim  string `json:"sim" binding:"required"`
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

	result, err := videoPlay(c, req)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "播放失败",
			Data: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  "成功",
		Data: result,
	})
}

func onStreamNotFound(_ context.Context, c *app.RequestContext) {
	type MediaInfo struct {
		MediaServerID string `json:"mediaServerId"`
		App           string `json:"app"`
		ID            string `json:"id"`
		IP            string `json:"ip"`
		Params        string `json:"params"`
		Port          int    `json:"port"`
		Schema        string `json:"schema"`
		Protocol      string `json:"protocol"`
		Stream        string `json:"stream"`
		Vhost         string `json:"vhost"`
	}
	var info MediaInfo
	if err := c.BindJSON(&info); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}

	// 流id固定格式 sim卡号_通道号 注意zlm的sim
	list := strings.Split(info.Stream, "_")
	if len(list) != 2 {
		c.JSON(http.StatusOK, Response{
			Code: int(InvalidArgs),
			Msg:  fmt.Sprintf("stream id 格式不正确 %s", info.Stream),
		})
		return
	}

	var (
		ip         = GlobalConfig.Zlm.OnStreamNotFound.IP
		dataType   = GlobalConfig.Zlm.OnStreamNotFound.DataType
		streamType = GlobalConfig.Zlm.OnStreamNotFound.StreamType
		key        = ""
	)
	sim := list[0]
	for i := 0; i < len(sim); i++ {
		if sim[i] != '0' {
			key = sim[i:]
			break
		}
	}
	req := Request[*model.P0x9101]{
		Key: key,
		Sim: list[0],
		Data: &model.P0x9101{
			ServerIPLen:  byte(len(ip)),
			ServerIPAddr: ip,
			TcpPort:      GlobalConfig.Zlm.Port,
			UdpPort:      0,
			ChannelNo:    cast.ToUint8(list[1]),
			DataType:     dataType,
			StreamType:   streamType,
		},
	}
	if _, err := videoPlay(c, req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: int(OtherFailed), // 0 表示允许播放
			Msg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code: int(Success), // 0 表示允许播放
		Msg:  "success",
	})
}

func videoPlay(c *app.RequestContext, req Request[*model.P0x9101]) (any, error) {
	// 发送指令
	if err := handleCommand(c, req.Key, req.Data); err != nil {
		return nil, fmt.Errorf("发送指令失败: %w", err)
	}

	// zlm播放规则 https://github.com/zlmediakit/ZLMediaKit/wiki/%E6%92%AD%E6%94%BEurl%E8%A7%84%E5%88%99
	type Result struct {
		StreamID string `json:"streamID"`
		MP4      string `json:"mp4"`
	}

	id := fmt.Sprintf("%s_%d", req.Sim, req.Data.ChannelNo)
	return Result{
		StreamID: id,
		MP4:      fmt.Sprintf(GlobalConfig.Zlm.PlayURLFormat, id),
	}, nil
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
