package internal

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/gb28181"
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"jt808_to_gb2818108/conf"
	"log/slog"
	"math/rand/v2"
	"time"
)

type AdapterTerminal struct {
	gb         *gb28181.Client
	hasDetails bool
	*terminal.Terminal
	portRule int
}

func NewAdapterTerminal(hasDetails bool) *AdapterTerminal {
	portRule := conf.GetData().JT808.JT1078.PortRule
	if portRule == 0 {
		portRule = -100
	}
	return &AdapterTerminal{
		hasDetails: hasDetails,
		portRule:   portRule,
	}
}

func (a *AdapterTerminal) OnJoinEvent(msg *service.Message, key string, err error) {
	if a.hasDetails {
		a.Terminal = terminal.New(terminal.WithHeader(msg.Header.ProtocolVersion, key))
	}
	if err == nil {
		go func() {
			// 随机延迟创建gb28181客户端 防止设备加入就创建 一下子太多
			time.Sleep(time.Duration(rand.IntN(3000)+100) * time.Millisecond)
			platform := conf.GetData().JT808.GB28181.Platform
			sim := key // 默认sim卡号就是key
			device := sendDevice(sim)
			slog.Debug("开始创建gb28181模拟设备",
				slog.String("key", key),
				slog.String("id", device.ID))
			a.gb = gb28181.New(sim, gb28181.WithPlatformInfo(gb28181.PlatformInfo{
				Domain:   platform.Domain,
				ID:       platform.ID,
				Password: platform.Password,
				IP:       platform.IP,
				Port:     platform.Port,
			}), gb28181.WithDeviceInfo(device),
				gb28181.WithTransport(conf.GetData().JT808.GB28181.Transport),             // 信令默认使用UDP 也可以TCP
				gb28181.WithKeepAliveSecond(conf.GetData().JT808.GB28181.KeepAliveSecond), // 默认30秒
				gb28181.WithInviteEventFunc(func(info *command.InviteInfo) *command.InviteInfo {
					// 默认jt1078收流端口是 gb28181 - 100
					// 如gb28181收流端口是10100 则jt1078收流端口是10000
					// 流媒体默认选择的是音视频流 视频h264 音频g711a
					info.JT1078Info.StreamTypes = []jt1078.PTType{jt1078.PTH264, jt1078.PTG711A}
					info.JT1078Info.Port = info.Port + a.portRule
					info.JT1078Info.RtpTypeConvert = func(pt jt1078.PTType) (byte, bool) {
						// 默认是按照h264=98的国标 但是zlm会失败 因此可以自行修改
						if pt == jt1078.PTH264 {
							return 96, true
						}
						// 其他情况使用默认的国标规范
						return 0, false
					}
					// 完成9101请求 让设备发送jt1078流
					// 只支持TCP被动模式 即设备TCP链接到gb28181平台
					if info.SessionName == "Play" {
						// 目前只支持点播
						ip := conf.GetData().JT808.JT1078.IP
						go send9101(Request[*model.P0x9101]{
							Key: info.JT1078Info.Sim,
							Data: &model.P0x9101{
								ServerIPLen:  byte(len(ip)),
								ServerIPAddr: ip,
								TcpPort:      uint16(info.JT1078Info.Port),
								UdpPort:      0,
								ChannelNo:    byte(info.JT1078Info.Channel),
								// gb28181 2016 96页
								// m=video6000RTP/AVP96”标识媒体类型为视频或视音频
								// m=audio8000RTP/AVP8”标识媒体类型为音频,传输端口为8000
								// 因此可以说协议没地方判断是不是音视频 就默认使用音视频和主码流
								DataType:   0, //  0-音视频 1-视频 2-双向对讲 3-监听 4-中心广播 5-透传
								StreamType: 0,
							},
						})
					}

					return info
				}),
			)
			if err := a.gb.Init(); err != nil {
				slog.Warn("init gb error",
					slog.String("key", key),
					slog.Any("err", err))
				return
			}
			go a.gb.Run()
		}()
	}
}

func (a *AdapterTerminal) OnLeaveEvent(key string) {
	fmt.Println("设备退出:", key)
	if a.gb != nil {
		a.gb.Stop()
	}
}

func (a *AdapterTerminal) OnNotSupportedEvent(msg *service.Message) {
	slog.Warn("暂不支持的指令",
		slog.String("key", msg.Key),
		slog.String("cmd", fmt.Sprintf("%x", msg.ExtensionFields.TerminalCommand)))
}

func (a *AdapterTerminal) OnReadExecutionEvent(msg *service.Message) {
	if a.hasDetails {
		str := fmt.Sprintf(" %s: [%x]", msg.ExtensionFields.TerminalCommand, msg.ExtensionFields.TerminalData)
		fmt.Println(time.Now().Format(time.DateTime), str)
		details := a.Terminal.ProtocolDetails(fmt.Sprintf("%x", msg.ExtensionFields.TerminalData))
		fmt.Println(details)
		fmt.Println()
	}
}

func (a *AdapterTerminal) OnWriteExecutionEvent(msg service.Message) {
	if a.hasDetails {
		str := "回复的报文"
		if msg.ExtensionFields.ActiveSend {
			str = "主动发送的报文"
		}
		str += fmt.Sprintf(" %s: [%x]", msg.ExtensionFields.PlatformCommand, msg.ExtensionFields.PlatformData)
		fmt.Println(time.Now().Format(time.DateTime), str)
		details := a.Terminal.ProtocolDetails(fmt.Sprintf("%x", msg.ExtensionFields.PlatformData))
		fmt.Println(details)
		fmt.Println()
	}
}
