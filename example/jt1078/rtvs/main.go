package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/hertz-contrib/cors"
	"jt1078/help"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	goJt808    *service.GoJT808
	address    string
	webAddress string
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(logger)
}

func main() {
	flag.StringVar(&address, "address", "0.0.0.0:8082", "address")
	flag.StringVar(&webAddress, "webAddress", "0.0.0.0:17001", "gatewayAddress")
	flag.Parse()

	goJt808 = service.New(
		service.WithHostPorts(address),
		service.WithNetwork("tcp"),
		service.WithCustomTerminalEventer(func() service.TerminalEventer {
			return &help.LogTerminal{}
		}),
	)
	go goJt808.Run()

	h := server.Default(
		server.WithHostPorts(webAddress),
	)
	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Token", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	customRoute(h)
	h.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		fmt.Println("未知路由", string(c.Request.Method()),
			string(c.Request.Path()), string(c.Request.QueryString()))
		c.String(http.StatusOK, "-1")
	})
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.Redirect(http.StatusMovedPermanently, []byte("index.html"))
	})
	h.StaticFile("/index.html", "tstrtvs.html")
	fmt.Println("访问页面", fmt.Sprintf("http://127.0.0.1:%s%s", strings.Split(webAddress, ":")[1], "/index.html"))
	h.Spin()
}

func customRoute(h *server.Hertz) {
	apiRtvsV1 := h.Group("/api/")
	apiRtvsV1.GET("/VideoControl", videoControl)
	apiRtvsV1.POST("/WCF0x9105", wcf0x9105)
}

func wcf0x9105(_ context.Context, c *app.RequestContext) {
	content := c.DefaultPostForm("Content", "")
	type RtvsResponse []struct {
		Sim        string `json:"Sim"`
		NotifyList []struct {
			Channel        byte `json:"Channel"`
			PacketLossRate byte `json:"PacketLossRate"`
		} `json:"NotifyList"`
	}
	var rtvsResponse RtvsResponse
	_ = json.Unmarshal([]byte(content), &rtvsResponse)
	var (
		wg       sync.WaitGroup
		once     sync.Once
		complete = make(chan error, 1)
	)
	defer close(complete)
	for _, v := range rtvsResponse {
		sim := v.Sim
		for _, notify := range v.NotifyList {
			wg.Add(1)
			go func(sim string, channel, packetLossRate byte) {
				defer wg.Done()
				tmp := &model.P0x9105{
					ChannelNo:       channel,
					PackageLossRate: packetLossRate,
				}
				replyMsg := goJt808.SendActiveMessage(&service.ActiveMessage{
					Key:              strings.TrimLeft(sim, "0"), //注意去掉前面的0
					Command:          consts.P9105AudioVideoControlStatusNotice,
					Body:             tmp.Encode(),
					OverTimeDuration: 3 * time.Second,
				})
				if replyMsg.ExtensionFields.Err != nil {
					once.Do(func() {
						complete <- replyMsg.ExtensionFields.Err
					})
					return
				}
			}(sim, notify.Channel, notify.PacketLossRate)
		}
	}
	wg.Wait()
	select {
	case err := <-complete:
		slog.Warn("9105",
			slog.Any("err", err))
		c.String(http.StatusInternalServerError, "-1")
		return
	default:
	}
	c.String(http.StatusOK, "1")

}

func videoControl(_ context.Context, c *app.RequestContext) {
	content := c.DefaultQuery("Content", "")
	activeMsg, err := rtvs2jt1078Pack(content)
	if err != nil {
		slog.Warn("rtvs to jt1078 fail",
			slog.String("content", content),
			slog.Any("err", err))
		c.String(http.StatusBadRequest, "-1")
		return
	}
	replyMsg := goJt808.SendActiveMessage(activeMsg)
	if replyMsg.ExtensionFields.Err != nil {
		slog.Warn("send active message fail",
			slog.String("content", content),
			slog.Any("err", replyMsg.ExtensionFields.Err))
		c.String(http.StatusBadRequest, "-1")
		return
	}

	type Handler struct {
		Parse          func(jtMsg *jt808.JTMessage) error
		CustomSignFunc func() string
	}

	var handler Handler
	switch activeMsg.Command {
	case consts.P9205QueryResourceList:
		tmp := &model.T0x1205{}
		handler.Parse = tmp.Parse
		handler.CustomSignFunc = func() string {
			return "9205-" + replyMsg.JTMessage.Header.TerminalPhoneNo
		}
	case consts.P9201SendVideoRecordRequest:
		tmp := &model.P0x9201{}
		handler.Parse = tmp.Parse
		handler.CustomSignFunc = func() string {
			return "9201-" + replyMsg.JTMessage.Header.TerminalPhoneNo + fmt.Sprintf("-%d", tmp.ChannelNo)
		}
	case consts.P9206FileUploadInstructions:
		tmp := &model.P0x9206{}
		handler.Parse = tmp.Parse
		handler.CustomSignFunc = func() string {
			return fmt.Sprintf("%d", activeMsg.ExtensionFields.PlatformSeq)
		}
	}

	if replyMsg.JTMessage.Header.ID == uint16(consts.T0001GeneralRespond) {
		tmp := &model.T0x0001{}
		handler.Parse = tmp.Parse
		handler.CustomSignFunc = func() string {
			if tmp.Result == 0 {
				return "1"
			}
			return "-1"
		}
	}

	if handler.Parse == nil || handler.CustomSignFunc == nil {
		c.String(http.StatusBadRequest, "-1")
		return
	}
	if err := handler.Parse(replyMsg.JTMessage); err != nil {
		slog.Warn("send active message fail",
			slog.String("content", content),
			slog.String("reply", fmt.Sprintf("%x", replyMsg.ExtensionFields.TerminalData)),
			slog.Any("err", replyMsg.ExtensionFields.Err))
		c.String(http.StatusBadRequest, "-1")
	}

	sign := handler.CustomSignFunc()
	c.String(http.StatusOK, sign)
	return
}

func rtvs2jt1078Pack(content string) (*service.ActiveMessage, error) {
	if len(content) < 24 {
		return nil, errors.New("content too short")
	}
	data := rtvsContent2Data(content)
	jtMsg := jt808.NewJTMessage()
	if err := jtMsg.Decode(data); err != nil {
		return nil, err
	}

	type Handler interface {
		Protocol() consts.JT808CommandType
		Parse(jtMsg *jt808.JTMessage) error
		Encode() []byte
	}
	var (
		handler  Handler
		overTime time.Duration
	)
	overTime = 3 * time.Second
	switch content[:4] {
	case "9101":
		handler = &model.P0x9101{}
	case "9102":
		handler = &model.P0x9102{}
	case "9201":
		handler = &model.P0x9201{}
		overTime = 5 * time.Second
	case "9202":
		handler = &model.P0x9202{}
	case "9205":
		handler = &model.P0x9205{}
		overTime = 10 * time.Second
	case "9206":
		handler = &model.P0x9206{}
	}
	if handler == nil {
		return nil, fmt.Errorf("unknown command: %s", content[:4])
	}
	if err := handler.Parse(jtMsg); err != nil {
		return nil, err
	}
	return &service.ActiveMessage{
		Key:              jtMsg.Header.TerminalPhoneNo,
		Command:          handler.Protocol(),
		Body:             handler.Encode(),
		OverTimeDuration: overTime,
	}, nil
}

func rtvsContent2Data(content string) []byte {
	content = strings.ToLower(content)
	effectiveBody, _ := hex.DecodeString(content)
	code := utils.CreateVerifyCode(effectiveBody)
	content = strings.ReplaceAll(content, "7d", "7d01")
	content = strings.ReplaceAll(content, "7e", "7d02")
	body, _ := hex.DecodeString(content)
	data := make([]byte, 0, 20)
	data = append(data, 0x7e)
	data = append(data, body...)
	data = append(data, code)
	data = append(data, 0x7e)
	return data
}
