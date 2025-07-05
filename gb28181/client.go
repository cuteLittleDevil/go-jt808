package gb28181

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"github.com/cuteLittleDevil/go-jt808/gb28181/internal/stream"
	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"log/slog"
	"sync"
	"time"
)

type Client struct {
	Options
	client    *sipgo.Client
	server    *sipgo.Server
	stopOnce  sync.Once
	stopChan  chan struct{}
	sn        uint32
	online    bool
	recipient sip.Uri
	from      *sip.FromHeader
	to        *sip.ToHeader
	callID    sip.CallIDHeader
	contact   *sip.ContactHeader
	// keepaliveReplyCount 心跳有多少次没有回复了.
	keepaliveReplyCount int
	manage              *stream.Manage
}

func New(sim string, opts ...Option) *Client {
	client := &Client{
		Options: Options{
			Sim:       sim,
			UserAgent: fmt.Sprintf("jt808-sim:%s", sim),
			KeepAlive: 30 * time.Second, // 默认30秒
			Transport: "UDP",            // 默认UDP
			JT1078ToGB28181erFunc: func() command.JT1078ToGB28181er {
				return stream.NewJT1078T0GB289181()
			},
		},
		stopChan: make(chan struct{}),
		sn:       1,
	}
	for _, op := range opts {
		op.F(&client.Options)
	}
	return client
}

func (c *Client) Init() error {
	// 不知道查询目录时 通道>4时如何按要求一个个分 干脆一次性发送完成
	sip.UDPMTUSize = 15000
	ua, err := sipgo.NewUA(
		sipgo.WithUserAgent(fmt.Sprintf("jt808-sim:%s", c.Sim)),
	)
	if err != nil {
		return err
	}

	if client, err := sipgo.NewClient(ua); err != nil {
		return err
	} else {
		c.client = client
	}

	if server, err := sipgo.NewServer(ua); err != nil {
		return err
	} else {
		c.server = server
		c.server.OnMessage(c.message)
		c.server.OnInvite(c.invite)
		c.server.OnAck(c.ack)
		c.server.OnBye(c.bye)
		c.server.OnNoRoute(func(req *sip.Request, tx sip.ServerTransaction) {
			err := tx.Respond(sip.NewResponseFromRequest(req, sip.StatusNotFound, "目前不支持的请求", nil))
			slog.Warn("OnNoRoute",
				slog.String("req", req.String()),
				slog.Any("err", err))
		})
	}

	c.recipient = sip.Uri{
		User: c.Options.PlatformInfo.ID,
		Host: c.Options.PlatformInfo.IP,
		Port: c.Options.PlatformInfo.Port,
	}
	c.from = &sip.FromHeader{
		Address: sip.Uri{
			User: c.Options.DeviceInfo.ID,
			Host: c.Options.DeviceInfo.IP,
			Port: c.Options.DeviceInfo.Port,
		},
		Params: sip.NewParams().Add("tag", sip.GenerateTagN(16)),
	}
	c.to = &sip.ToHeader{
		Address: sip.Uri{
			User: c.Options.PlatformInfo.ID,
			Host: c.Options.PlatformInfo.IP,
			Port: c.Options.PlatformInfo.Port,
		},
		Params: sip.NewParams(),
	}
	c.contact = &sip.ContactHeader{
		// 这里客户端都用虚假的内网ip 反正客户端的内网ip 服务器那边用不到
		Address: sip.Uri{
			User: c.Options.DeviceInfo.ID,
			Host: c.Options.DeviceInfo.IP,
			Port: c.Options.DeviceInfo.Port,
		},
	}
	customCallID := fmt.Sprintf("%d@%s", time.Now().Unix(), c.Options.Sim)
	c.callID = sip.CallIDHeader(customCallID)

	c.manage = stream.NewManage(c.Options.OnInviteEventFunc, c.Options.JT1078ToGB28181erFunc)
	go c.manage.Run()
	return nil
}

func (c *Client) Run() {
	ticker := time.NewTicker(c.Options.KeepAlive)
	registerChan := make(chan bool, 1)
	defer func() {
		ticker.Stop()
		close(registerChan)
	}()
	registerChan <- true

	for {
		select {
		case <-c.stopChan:
			if c.online {
				if _, err := c.handleRegister(false); err != nil {
					slog.Warn("unregister fail",
						slog.String("sim", c.Options.Sim),
						slog.String("id", c.Options.DeviceInfo.ID),
						slog.Any("err", err))
				}
			}
			return
		case <-registerChan:
			if ok, err := c.handleRegister(true); err != nil {
				slog.Warn("register fail",
					slog.String("sim", c.Options.Sim),
					slog.String("id", c.Options.DeviceInfo.ID),
					slog.Any("err", err))
			} else if ok {
				c.online = true
				slog.Info("register success",
					slog.String("sim", c.Options.Sim),
					slog.String("id", c.Options.DeviceInfo.ID))
			}
		case <-ticker.C: // 心跳保活
			if !c.online {
				registerChan <- true // 不在线就注册
				continue
			}
			if err := c.handleKeepalive(); err != nil {
				c.keepaliveReplyCount++
				if c.keepaliveReplyCount == 3 { // 服务器一段时间不回复就触发注册
					c.online = false
					c.keepaliveReplyCount = 0
					registerChan <- true // 保活指令失败就重新注册
				}
			} else {
				c.keepaliveReplyCount = 0
			}
		}
	}
}

func (c *Client) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopChan)
		if c.client != nil {
			if err := c.client.Close(); err != nil {
				slog.Warn("client close fail",
					slog.String("sim", c.Options.Sim),
					slog.String("id", c.Options.DeviceInfo.ID),
					slog.Any("err", err))
			}
		}
		if c.server != nil {
			if err := c.server.Close(); err != nil {
				slog.Warn("server close fail",
					slog.String("sim", c.Options.Sim),
					slog.String("id", c.Options.DeviceInfo.ID),
					slog.Any("err", err))
			}
		}
		if c.manage != nil {
			c.manage.Stop()
		}
	})
}

func (c *Client) getResponse(tx sip.ClientTransaction) (*sip.Response, error) {
	select {
	case <-tx.Done():
		return nil, fmt.Errorf("事务已终止")
	case res := <-tx.Responses():
		return res, nil
	}
}
