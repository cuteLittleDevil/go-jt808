package internal

import (
	"github.com/cuteLittleDevil/go-jt808/gb28181"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/go-resty/resty/v2"
	"jt808_to_gb2818108/conf"
	"log/slog"
	"time"
)

func send9101(req Request[*model.P0x9101]) {
	config := conf.GetData().JT808.JT1078
	var res Response[any]
	httpClient := resty.New()
	httpClient.SetDebug(false)
	httpClient.SetTimeout(5 * time.Second)
	_, err := httpClient.R().
		SetBody(req).
		SetResult(&res).
		ForceContentType("application/json; charset=utf-8").
		Post(config.OnPlayURL)
	if err != nil {
		slog.Warn("send 9101 fail",
			slog.Any("data", req.Data),
			slog.Any("err", err))
		return
	}
	slog.Info("send 9101",
		slog.Any("res", res))
}

func sendDevice(sim string) gb28181.DeviceInfo {
	config := conf.GetData().JT808.GB28181
	var res Response[gb28181.DeviceInfo]
	httpClient := resty.New()
	httpClient.SetDebug(false)
	httpClient.SetTimeout(5 * time.Second)
	_, err := httpClient.R().
		SetBody(map[string]string{
			"sim": sim,
		}).
		SetResult(&res).
		ForceContentType("application/json; charset=utf-8").
		Post(config.Device.OnConfigURL)
	if err != nil {
		slog.Warn("send device fail",
			slog.String("sim", sim),
			slog.Any("err", err))
		return defaultDeviceInfo(sim)
	}
	if res.Data.ID == "" {
		return defaultDeviceInfo(sim)
	}
	return res.Data
}
