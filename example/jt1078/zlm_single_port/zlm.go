package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"log/slog"
	"time"
)

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
	default:
		return "执行成功"
	}
}

func zlmStartSendRtpTalk(url string, params map[string]string) error {
	type Response struct {
		Code zlmCode `json:"code"`
		Port int     `json:"port"`
	}
	var res Response
	client := resty.New()
	client.SetDebug(true)
	client.SetTimeout(3 * time.Second)
	// 范例：http://127.0.0.1/index/api/startSendRtpTalk?
	// secret=5xGbdUpfXnsiW3uZq2CApzSyxSFrIWpc&vhost=__defaultVhost__&app=rtp&stream=test&ssrc=1&recv_stream_id=000000001003_1_0_0
	if _, err := client.R().
		SetQueryParams(params).
		SetResult(&res).
		ForceContentType("application/json; charset=utf-8").
		Get(url); err != nil {
		return err
	}
	if res.Code != Success {
		return fmt.Errorf("code is %d[%s]", res.Code, res.Code)
	}
	return nil
}

func isExistMediaInfo(url string, params map[string]string) bool {
	type Response struct {
		Code zlmCode `json:"code"`
		Port int     `json:"port"`
	}
	var res Response
	client := resty.New()
	client.SetDebug(true)
	client.SetTimeout(3 * time.Second)
	// 范例：http://127.0.0.1/index/api/startSendRtpTalk?
	// secret=5xGbdUpfXnsiW3uZq2CApzSyxSFrIWpc&vhost=__defaultVhost__&app=rtp&stream=test&ssrc=1&recv_stream_id=000000001003_1_0_0
	if _, err := client.R().
		SetQueryParams(params).
		SetResult(&res).
		ForceContentType("application/json; charset=utf-8").
		Get(url); err != nil {
		slog.Warn("isExistMediaInfo",
			slog.String("url", url),
			slog.Any("params", params),
			slog.Any("err", err))
		return false
	}
	// 不存在的话 code=-500
	return res.Code == Success
}
