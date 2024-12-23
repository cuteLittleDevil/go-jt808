package file

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/go-resty/resty/v2"
	_ "github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"log/slog"
	"time"
	"web/alarm/command"
	"web/alarm/conf"
	"web/internal/shared"
)

func OnAlarmEvent(data shared.EventData, location command.Location) {
	if location.T0x0200AdditionExtension0x64.ParseSuccess {
		send9208(data, location.T0x0200AdditionExtension0x64.P9208AlarmSign)
	}

	if location.T0x0200AdditionExtension0x65.ParseSuccess {
		send9208(data, location.T0x0200AdditionExtension0x65.P9208AlarmSign)
	}

	if location.T0x0200AdditionExtension0x66.ParseSuccess {
		send9208(data, location.T0x0200AdditionExtension0x66.P9208AlarmSign)
	}

	if location.T0x0200AdditionExtension0x67.ParseSuccess {
		send9208(data, location.T0x0200AdditionExtension0x67.P9208AlarmSign)
	}

	if location.T0x0200AdditionExtension0x70.ParseSuccess {
		send9208(data, location.T0x0200AdditionExtension0x70.P9208AlarmSign)
	}
}

func send9208(data shared.EventData, p9208AlarmSign model.P9208AlarmSign) {
	attachIP := data.AttachIP
	attachPort := data.AttachPort
	p9208 := &model.P0x9208{
		ServerIPLen:    byte(len(attachIP)),
		ServerAddr:     attachIP,
		TcpPort:        uint16(attachPort),
		UdpPort:        0,
		P9208AlarmSign: p9208AlarmSign,
		// AlarmID 报警编号 byte[32] 平台给报警分配的唯一编号
		AlarmID: uuid.New().String(),
		Reserve: make([]byte, 16),
	}

	client := resty.New()
	client.SetTimeout(5 * time.Second)
	url := data.Address + conf.GetData().ServerConfig.OnFileApi
	var result shared.Response
	_, err := client.R().
		SetBody(shared.Request[*model.P0x9208]{
			Key:     data.Key,
			Command: p9208.Protocol(),
			Data:    p9208,
		}).
		SetResult(&result).
		ForceContentType("application/json; charset=utf-8").
		Post(url)
	if err != nil {
		slog.Error("9208",
			slog.String("key", data.Key),
			slog.Any("data", p9208.String()),
			slog.String("url", url))
		return
	}
	slog.Debug("9208",
		slog.String("key", data.Key),
		slog.Any("data", p9208.String()),
		slog.Any("result", result))
}
