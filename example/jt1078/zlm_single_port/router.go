package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/patrickmn/go-cache"
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

	_, result, err := videoPlay(c, req)
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

	// 流id固定格式 sim卡号_通道号_数据类型_主子码流 注意zlm的sim卡号是保留前面0的
	list := strings.Split(info.Stream, "_")
	fmt.Println("onStreamNotFound", info.Stream)

	if len(list) != 4 {
		c.JSON(http.StatusOK, Response{
			Code: int(InvalidArgs),
			Msg:  fmt.Sprintf("stream id 格式不正确 %s", info.Stream),
		})
		return
	}

	var (
		ip  = GlobalConfig.Zlm.OnStreamNotFound.IP
		key = ""
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
			DataType:     cast.ToUint8(list[2]),
			StreamType:   cast.ToUint8(list[3]),
		},
	}
	if _, _, err := videoPlay(c, req); err != nil {
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

func onPublish(_ context.Context, c *app.RequestContext) {
	type Publish struct {
		MediaServerId string `json:"mediaServerId"`
		App           string `json:"app"`
		Id            string `json:"id"`
		IP            string `json:"ip"`
		Params        string `json:"params"`
		Port          int    `json:"port"`
		Schema        string `json:"schema"`
		Protocol      string `json:"protocol"`
		Stream        string `json:"stream"`
		Vhost         string `json:"vhost"`
	}
	// https://github.com/ZLMediaKit/ZLMediaKit/wiki/MediaServer%E6%94%AF%E6%8C%81%E7%9A%84HTTP-HOOK-API#7on_publish
	var info Publish
	if err := c.BindJSON(&info); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}

	reviseID := info.Stream
	// 如果流id是sim卡号+通道号的格式 则改变成sim卡号_通道号_数据类型_主子码流
	if strings.Count(info.Stream, "_") == 1 {
		reviseID = info.Stream + "_0_0"
		if v, ok := c.Value("cache").(*cache.Cache); ok {
			if id, ok := v.Get(info.Stream); ok {
				reviseID = cast.ToString(id)
			}
		}
	}

	fmt.Println("修改流程名称", info.Stream, reviseID)

	type ZlmReply struct {
		Code           int    `json:"code"`
		AddMuteAudio   bool   `json:"add_mute_audio"`
		ContinuePushMs int    `json:"continue_push_ms"`
		EnableAudio    bool   `json:"enable_audio"`
		EnableFmp4     bool   `json:"enable_fmp4"`
		EnableHls      bool   `json:"enable_hls"`
		EnableHlsFmp4  bool   `json:"enable_hls_fmp4"`
		EnableMp4      bool   `json:"enable_mp4"`
		EnableRtmp     bool   `json:"enable_rtmp"`
		EnableRtsp     bool   `json:"enable_rtsp"`
		EnableTs       bool   `json:"enable_ts"`
		HlsSavePath    string `json:"hls_save_path"`
		ModifyStamp    bool   `json:"modify_stamp"`
		Mp4AsPlayer    bool   `json:"mp4_as_player"`
		Mp4MaxSecond   int    `json:"mp4_max_second"`
		Mp4SavePath    string `json:"mp4_save_path"`
		AutoClose      bool   `json:"auto_close"`
		StreamReplace  string `json:"stream_replace"`
	}
	c.JSON(http.StatusOK, ZlmReply{
		Code:           int(Success),
		AddMuteAudio:   true,
		ContinuePushMs: 10000,
		EnableAudio:    true,
		EnableFmp4:     true,
		EnableHls:      true,
		EnableHlsFmp4:  false,
		EnableMp4:      false,
		EnableRtmp:     true,
		EnableRtsp:     true,
		EnableTs:       true,
		HlsSavePath:    "/hls_save_path/",
		ModifyStamp:    false,
		Mp4AsPlayer:    false,
		Mp4MaxSecond:   3600,
		Mp4SavePath:    "/mp4_save_path/",
		AutoClose:      false,
		StreamReplace:  reviseID,
	})
}

func startSendRtpTalk(_ context.Context, c *app.RequestContext) {
	type RtpTalk struct {
		Request[*model.P0x9101]
		Stream string `json:"stream"`
	}

	var req RtpTalk
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
			Data: err.Error(),
		})
		return
	}

	receiveID, _, err := videoPlay(c, req.Request)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "播放失败",
			Data: err.Error(),
		})
		return
	}

	// 等待设备流推送到zlm服务
	intercom := GlobalConfig.Zlm.Intercom
	ticker := time.NewTicker(time.Duration(intercom.IntervalSecond) * time.Second)
	overTime := time.After(time.Duration(intercom.OvertimeSecond) * time.Second)
	defer ticker.Stop()
	var (
		isExist    = false
		isOverTime = false
	)
	for !isExist && !isOverTime {
		select {
		case <-ticker.C:
			isExist = isExistMediaInfo(intercom.GetMediaInfoURL, map[string]string{
				"secret": GlobalConfig.Zlm.Secret,
				"schema": "rtsp",
				"vhost":  intercom.Vhost,
				"app":    intercom.App,
				"stream": receiveID,
			})
			break
		case <-overTime:
			isOverTime = true
		}
	}

	if isOverTime {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "收流超时",
		})
		return
	}

	// 确保完全成功
	time.Sleep(500 * time.Millisecond)
	url := intercom.Url
	params := map[string]string{
		"secret":         GlobalConfig.Zlm.Secret,
		"vhost":          intercom.Vhost,
		"app":            intercom.App,
		"stream":         req.Stream,
		"ssrc":           fmt.Sprintf("%s_%d", req.Sim, req.Data.ChannelNo),
		"recv_stream_id": receiveID,
	}
	if err := zlmStartSendRtpTalk(url, params); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "播放失败",
			Data: err.Error(),
		})
		return
	}
	type Reply struct {
		Stream    string `json:"stream"`
		ReceiveID string `json:"receiveID"`
	}
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  "播放成功",
		Data: Reply{
			ReceiveID: receiveID,
			Stream:    req.Stream,
		},
	})
}

func videoPlay(c *app.RequestContext, req Request[*model.P0x9101]) (string, any, error) {
	// 修改zlm的流id 支持主子码流
	zlmStreamID := fmt.Sprintf("%s_%d", req.Sim, req.Data.ChannelNo)
	id := fmt.Sprintf("%s_%d_%d_%d", req.Sim, req.Data.ChannelNo, req.Data.DataType, req.Data.StreamType)
	if v, ok := c.Value("cache").(*cache.Cache); ok {
		v.Set(zlmStreamID, id, 5*time.Second)
	}
	// 发送指令
	if err := handleCommand(c, req.Key, req.Data); err != nil {
		return "", nil, fmt.Errorf("发送指令失败: %w", err)
	}

	// zlm播放规则 https://github.com/zlmediakit/ZLMediaKit/wiki/%E6%92%AD%E6%94%BEurl%E8%A7%84%E5%88%99
	type Result struct {
		StreamID string `json:"streamID"`
		MP4      string `json:"mp4"`
		HTTPSMP4 string `json:"httpsMP4"`
	}
	return id, Result{
		StreamID: id,
		MP4:      fmt.Sprintf(GlobalConfig.Zlm.PlayURLFormat, id),
		HTTPSMP4: fmt.Sprintf(GlobalConfig.Zlm.HttpsPlayURLFormat, id),
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
