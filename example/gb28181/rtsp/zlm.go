package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"github.com/go-resty/resty/v2"
	"log/slog"
	"rtsp/conf"
	"time"
)

type zlmHandle struct {
	ssrc string
}

func (z *zlmHandle) OnAck(info *command.InviteInfo) {
	z.ssrc = info.SSRC
	if err := addStreamProxy(); err != nil {
		slog.Error("addStreamProxy",
			slog.String("err", err.Error()))
	}
	// 等待zlm拉流成功 或者改成轮询有没有成功？
	time.Sleep(3 * time.Second)
	if err := startSendRtp(info.SSRC, info.IP, info.Port); err != nil {
		slog.Error("startSendRtp",
			slog.String("err", err.Error()))
	}
	slog.Info("rtsp转ps流",
		slog.String("ip", info.IP),
		slog.Int("port", info.Port),
		slog.String("ssrc", info.SSRC),
		slog.String("rtsp", conf.GetData().ZLM.Rtsp))
}

func (z *zlmHandle) ConvertToGB28181(_ []byte) ([][]byte, error) {
	return nil, nil
}

func (z *zlmHandle) OnBye(msg string) {
	slog.Info("点播结束",
		slog.String("msg", msg))
	stopSendRtp(z.ssrc)
	closeStreams()
}

type zlmCode int

const (
	Exception   zlmCode = -400 //代码抛异常
	InvalidArgs zlmCode = -300 //参数不合法
	SqlFailed   zlmCode = -200 //sql执行失败
	AuthFailed  zlmCode = -100 //鉴权失败
	OtherFailed zlmCode = -1   //业务代码执行失败，
	Success     zlmCode = 0    //执行成功
)

func (z zlmCode) String() string {
	switch z {
	case Exception:
		return "代码抛异常"
	case InvalidArgs:
		return "参数不合法"
	case SqlFailed:
		return "sql执行失败"
	case AuthFailed:
		return "鉴权失败"
	case OtherFailed:
		return "业务代码执行失败"
	case Success:
		return "执行成功"
	default:
		return "执行失败"
	}
}

func addStreamProxy() error {
	// http://127.0.0.1/index/api/addStreamProxy?vhost=__defaultVhost__&app=proxy&stream=0&url=rtmp://live.hkstv.hk.lxdns.com/live/hks2
	url := conf.GetData().ZLM.AddStreamProxy
	params := map[string]string{
		"secret": conf.GetData().ZLM.Secret,
		"vhost":  conf.GetData().ZLM.Vhost,
		"app":    conf.GetData().ZLM.App,
		"stream": conf.GetData().ZLM.Stream,
		"url":    conf.GetData().ZLM.Rtsp,
	}
	return zlmHTTPHandle(url, params)
}

func startSendRtp(ssrc string, dstURL string, dstPort int) error {
	// http://127.0.0.1/index/api/startSendRtp?secret=035c73f7-bb6b-4889-a715-d9eb2d1925cc&vhost=__defaultVhost__&
	//app=live&stream=test&ssrc=1&dst_url=127.0.0.1&dst_port=10000&is_udp=0
	url := conf.GetData().ZLM.StartSendRtp
	params := map[string]string{
		"secret":   conf.GetData().ZLM.Secret,
		"vhost":    conf.GetData().ZLM.Vhost,
		"app":      conf.GetData().ZLM.App,
		"stream":   conf.GetData().ZLM.Stream,
		"ssrc":     ssrc,
		"dst_url":  dstURL,
		"dst_port": fmt.Sprintf("%d", dstPort),
		"is_udp":   "0",
	}
	return zlmHTTPHandle(url, params)
}

func stopSendRtp(ssrc string) {
	// http://127.0.0.1/index/api/stopSendRtp?secret=035c73f7-bb6b-4889-a715-d9eb2d1925cc&vhost=__defaultVhost__&app=live&stream=test
	url := conf.GetData().ZLM.StopSendRtp
	params := map[string]string{
		"secret": conf.GetData().ZLM.Secret,
		"vhost":  conf.GetData().ZLM.Vhost,
		"app":    conf.GetData().ZLM.App,
		"stream": conf.GetData().ZLM.Stream,
		"ssrc":   ssrc,
	}
	if err := zlmHTTPHandle(url, params); err != nil {
		slog.Error("stopSendRtp",
			slog.String("err", err.Error()))
	}
}

func closeStreams() {
	// http://127.0.0.1/index/api/close_streams?schema=rtmp&vhost=__defaultVhost__&app=live&stream=0&force=1
	url := conf.GetData().ZLM.CloseStreams
	params := map[string]string{
		"secret": conf.GetData().ZLM.Secret,
		"vhost":  conf.GetData().ZLM.Vhost,
		"app":    conf.GetData().ZLM.App,
		"stream": conf.GetData().ZLM.Stream,
		"force":  "1",
	}
	if err := zlmHTTPHandle(url, params); err != nil {
		slog.Error("closeStreams",
			slog.String("err", err.Error()))
	}
}

func zlmHTTPHandle(url string, params map[string]string) error {
	type Response struct {
		Code zlmCode `json:"code"`
		Msg  string  `json:"msg"`
	}
	var res Response
	client := resty.New()
	client.SetDebug(true)
	client.SetTimeout(3 * time.Second)
	if _, err := client.R().
		SetQueryParams(params).
		SetResult(&res).
		Get(url); err != nil {
		return err
	}
	if res.Code != Success {
		return fmt.Errorf("code is %d[%s]", res.Code, res.Code)
	}
	return nil
}
