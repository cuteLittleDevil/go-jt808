package m7s_jt1078

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/m7s-jt1078/pkg"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt1078"
	"golang.org/x/net/context"
	"log/slog"
	"m7s.live/v5"
	m7sPkg "m7s.live/v5/pkg"
	"net"
	"os"
	"strings"
	"time"
)

var _ = m7s.InstallPlugin[JT1078Plugin]()

type (
	JT1078Plugin struct {
		m7s.Plugin
		AudioPorts  [2]int              `default:"[10000,10010]" desc:"音频端口 用于下发数据"`
		RealTime    jt1078Stream        `default:"{}" desc:"实时推流"`
		Playback    jt1078Stream        `default:"{}" desc:"回放推流"`
		Simulations []jt1078Simulations `default:"[]" desc:"模拟客户端推流"`
		sessions    *pkg.AudioManager
	}

	jt1078Stream struct {
		Addr       string `default:"0.0.0.0:1078" desc:"视频端口"`
		OnJoinURL  string `default:"http://127.0.0.1:10011/api/v1/join" desc:"第一个包正确解析时触发"`
		OnLeaveURL string `default:"http://127.0.0.1:10011/api/v1/leave" desc:"推流客户端离开时"`
		Prefix     string `default:"live/jt1078" desc:"推流前缀"`
	}

	jt1078Simulations struct {
		Name string `default:"./data/data.txt" desc:"文件名"`
		Addr string `default:"127.0.0.1:1078" desc:"地址"`
	}
)

func (j *JT1078Plugin) OnInit() (err error) {
	if j.RealTime.Addr != "" {
		j.sessions = pkg.NewAudioManager(j.AudioPorts)
		if err := j.sessions.Init(); err != nil {
			j.Error("init error",
				slog.String("err", err.Error()))
			return err
		}
		j.Info("audio init",
			slog.Any("limits", j.AudioPorts))
		go j.sessions.Run()
		service := pkg.NewService(j.RealTime.Addr, j.Logger,
			pkg.WithURL(j.RealTime.OnJoinURL, j.RealTime.OnLeaveURL),
			pkg.WithPubFunc(func(ctx context.Context, pack *jt1078.Packet) (publisher *m7s.Publisher, err error) {
				streamPath := strings.Join([]string{
					j.RealTime.Prefix, pack.Sim, fmt.Sprintf("%d", pack.LogicChannel),
				}, "-")
				if pub, err := j.Publish(ctx, streamPath); err == nil {
					return pub, nil
				} else if errors.Is(err, m7sPkg.ErrStreamExist) { // 实时的流名称重复了 在给一次机会
					streamPath += fmt.Sprintf("-%d", time.Now().UnixMilli())
					return j.Publish(ctx, streamPath)
				} else {
					return pub, err
				}
			}),
			pkg.WithSessions(j.sessions),
			pkg.WithPTSFunc(func(_ *jt1078.Packet) time.Duration {
				return time.Duration(time.Now().UnixMilli()) * 90 // 实时视频使用本机时间戳
			}),
		)
		go service.Run()
	}
	if j.Playback.Addr != "" {
		service := pkg.NewService(j.Playback.Addr, j.Logger,
			pkg.WithURL(j.Playback.OnJoinURL, j.Playback.OnLeaveURL),
			pkg.WithPubFunc(func(ctx context.Context, pack *jt1078.Packet) (publisher *m7s.Publisher, err error) {
				streamPath := strings.Join([]string{
					j.Playback.Prefix, pack.Sim, fmt.Sprintf("%d", pack.LogicChannel),
				}, "-")
				return j.Publish(ctx, streamPath) // 回放唯一
			}),
			pkg.WithPTSFunc(func(pack *jt1078.Packet) time.Duration {
				return time.Duration(pack.Timestamp) * 90 // 录像回放使用设备的
			}),
		)
		go service.Run()
	}
	if len(j.Simulations) > 0 {
		go j.simulationPull()
	}
	return nil
}

func (j *JT1078Plugin) simulationPull() {
	time.Sleep(1 * time.Second) // 等待jt1078服务都启动好
	for _, v := range j.Simulations {
		go func(name string, addr string) {
			j.Warn("simulation pull",
				slog.String("name", name),
				slog.String("addr", addr))
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				j.Warn("simulation pull",
					slog.String("name", name),
					slog.String("addr", addr),
					slog.String("err", err.Error()))
				return
			}
			defer func() {
				_ = conn.Close()
			}()
			content, err := os.ReadFile(name)
			if err != nil {
				j.Warn("simulation pull",
					slog.String("name", name),
					slog.String("addr", addr),
					slog.String("err", err.Error()))
			}
			data, _ := hex.DecodeString(string(content))
			const groupSum = 1023
			for {
				start := 0
				end := 0
				for i := 0; i < len(data)/groupSum; i++ {
					start = i * groupSum
					end = start + groupSum
					_, _ = conn.Write(data[start:end])
					time.Sleep(50 * time.Millisecond)
				}
				_, _ = conn.Write(data[end:])
			}
		}(v.Name, v.Addr)
	}
}
