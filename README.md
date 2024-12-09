# go-jt808

- 本项目已更好支持二次开发为目标 可通过各种自定义事件去完成相应功能 常见案例如下
1. jt1078视频 [详情](./example/jt1078/README.md)

``` txt
jt808服务端 jt1078服务端 模拟器在2核4G腾讯云服务器
平台下发0x9101等指令 模拟器开始推流等动作
```
| 流媒体服务 | 语言 | 描述       | 说明 |
|----------|-----|-------------------|-----|
| RTVS | 不开源可使用  | 在线测试页面 http://49.234.235.7:17001 | [详情点击](./example/jt1078/README.md#rtvs)  |
| LAL | go  | 在线播放地址 http://49.234.235.7:8080/live/295696659617_1.flv | [详情点击](./example/jt1078/README.md#lal)  |
| sky-java | java  | 需要部署后 HTTP请求 10秒内拉流 参考格式如下 <br/> http://222.244.144.181:7777/video/1001-1-0-0.live.mp4 | [详情点击](./example/jt1078/README.md#sky-java)  |
| m7s | go  | 在线播放地址 http://49.234.235.7:8088/mp4/live/jt1078-295696659617-1.mp4 | [详情点击](https://github.com/cuteLittleDevil/m7s-jt1078)  |

2. 主动安全附件 [流程](./example/attachment/README.md#主动安全)
``` txt
默认支持苏标 可自定义各事件扩展（开始、传输进度、补传情况、完成、退出等事件）
```

3. 存储经纬度 [详情](./README.md#save)
``` txt
jt808服务端 模拟器 消息队列 数据库都运行在2核4G腾讯云服务器
测试每秒保存5000条的情况 约5.5小时保存了近1亿的经纬度
```

4. 分布式集群方案 [详情](./example/distributed_cluster/README.md)
``` txt
使用nginx把终端分配到多个808服务上 下发数据使用广播
存在则回复终端应答到新主题 不存在则忽略
```

5. 平台下发指令给终端 [获取参数](./example/protocol/active_reply/main.go) [立即拍摄](./example/protocol/camera/main.go)
``` txt
主动下发给设备指令 获取应答的情况
```

6. 协议交互详情 [代码参考](./example/protocol/register/main.go)
``` txt
使用自定义模拟器 可以轻松生成测试用的报文 有详情描述
```

7. 自定义协议扩展 [代码参考](./example/protocol/custom_parse/main.go)
``` txt
自定义附加信息处理 获取想要的扩展内容
```

---
- 看飞哥的单机TCP百万并发 好奇有数据情况的表现 因此国庆准备试一试有数据的情况
- 性能测试 单机[2核4G机器]并发10w+ 每日保存4亿+经纬度 [详情](./README.md#save)
- 支持JT808(2011/2013/209) JT1078(需要其他流媒体服务)
- 支持分包和自动补传 支持主动安全扩展(苏标 黑标 广东标 湖南标 四川标)

| 特点  |   描述   |
| :---:   | -------- |
|  安全可靠 | 核心协议部分测试覆盖率100% 纯原生go实现(不依赖任何库)  |
|  简洁优雅 | 核心代码不到1000行 不使用任何锁 仅使用channel完成  |
|  易于扩展 | 方便二次开发 有JT1078流媒体对接 保存经纬度等案例  |

---

[快速开始](./example/quick_start/main.go)
``` go
package main

import (
	"github.com/cuteLittleDevil/go-jt808/attachment"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"log/slog"
	"os"
)

var goJt808 *service.GoJT808

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(logger)

	attach := attachment.New(
		attachment.WithNetwork("tcp"),
		attachment.WithHostPorts("0.0.0.0:10001"),
		attachment.WithActiveSafetyType(consts.ActiveSafetyJS), // 默认苏标 支持黑标 广东标 湖南标 四川标
		attachment.WithFileEventerFunc(func() attachment.FileEventer {
			return &meFileEvent{} // 自定义文件处理 开始 结束 当前进度 补传 完成等事件
		}),
	)
	go attach.Run()
}

func main() {
	goJt808 = service.New(
		service.WithHostPorts("0.0.0.0:808"),
		service.WithNetwork("tcp"),
		service.WithCustomTerminalEventer(func() service.TerminalEventer {
			return &meTerminal{} // 自定义终端事件 终端进入 离开 读写报文事件
		}),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				consts.T0200LocationReport: &meLocation{}, // 自定义0x0200位置解析等
			}
		}),
	)
	goJt808.Run()
}

```

---
- 目前(2024-10-01)前的go语言版本个人觉得都不好因此都不推荐参考 推荐参考资料如下
## 参考资料
| 项目名称           | 语言   | 日期       | Star 数 | 链接                                       |
|--------------------|--------|------------|---------|--------------------------------------------|
| JT808           | C#     | 2024-10-01 | 534       | [JT808 C#](https://github.com/SmallChi/JT808.git) |
| jt808-server    | Java   | 2024-10-01 | 1.4k+     | [JT808 Java](https://gitee.com/yezhihao/jt808-server.git) |

- [飞哥的开发内功修炼](https://github.com/yanfeizhang/coder-kung-fu?tab=readme-ov-file)
- [协议文档 (PDF整理)](https://gitee.com/yezhihao/jt808-server/tree/master/协议文档 )
- [协议文档 (官网)](https://jtst.mot.gov.cn/hb/search/stdHBView?id=a3011cd31e6602ec98f26c35329e88e4)
- [协议解析网站](https://jttools.smallchi.cn/jt808)
- [bcd转dec编码](https://github.com/deatil/lakego-admin/tree/main/pkg/lakego-pkg/go-encoding/bcd)
- [lal流媒体文档](https://pengrl.com/lal/#/streamurllist)

## 性能测试
- java模拟器(QQ群下载 373203450)
- go模拟器 [详情点击](./example/simulator/client/main.go#go模拟器)

### 连接数测试
[详情点击](./example/simulator/README.md#online)

- 2台云服务器各开5w+客户端 总计10w+

| 服务端版本  |   场景   | 并发数 |  服务器配置  | 服务器使用资源情况 |  描述  |
| :---:   | :-------: | :--: | :------: | :-------------- | :----------------------------: |
|  v0.3.0 | 连接数测试  | 10w+ |  2核4G | 120%+cpu 1.7G内存  | 10.0.16.5上开启服务端和模拟器  <br/> 10.0.16.14机器上开启模拟器 |

<h3 id="save"> 模拟经纬度存储测试 </h3>

[详情点击](./example/simulator/README.md#save)

- save进程丢失了部分数据 channel队列溢出抛弃 (测试channel队列为100)
- 保存1亿丢失826条 保存4.32亿丢失1216条（分两次测试)

| 服务端版本  | 客户端 |  服务器配置  | 服务使用资源情况 |  描述  |
| :---:   | :--: | :------: | :-------------- | :----------------------------: |
|  v0.3.0 | 1w go模拟器 |  2核4G | 35%cpu 180.4MB内存 | 每秒5000 一共保存经纬度1亿  <br/> 实际保存99999174 成功率99.999% |

| 服务  |   cpu   | 内存 | 描述 |
| :---:   | :-------: | :--: | :--: |
|  server | 35% | 180.4MB | 808服务端 |
|  client | 23% | 196MB | 模拟客户端 |
|  save |  18% | 68.8MB | 存储数据服务 |
|  nats-server | 20% | 14.8MB | 消息队列 |
|  taosadapter | 37% | 124.3MB | tdengine数据库适配 |
|  taosd | 15% | 124.7MB | tdengine数据库 |
## 使用案例

### 1. 协议处理

#### 1.1 协议解析
``` go
func main() {
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, "1")) // 自定义模拟器 2013版本
	data := t.CreateDefaultCommandData(consts.T0100Register) // 生成的预值注册指令 0x0100
	fmt.Println(fmt.Sprintf("模拟器生成的[%x]", data))

	jtMsg := jt808.NewJTMessage()
	_ = jtMsg.Decode(data) // 解析固定请求头

	var t0x0100 model.T0x0100 // 把body数据解析到结构体中
	_ = t0x0100.Parse(jtMsg)
	fmt.Println(jtMsg.Header.String())
	fmt.Println(t0x0100.String())
}

```

部分输出 [输出详情](./example/protocol/README.md#register)
``` txt
模拟器生成的[7e010000300000000000010001001f006e63643132337777772e3830382e636f6d0000000000000000003736353433323101b2e2413132333435363738797e]
[0100] 消息ID:[256] [终端-注册]
消息体属性对象: {
        [0000000000110000] 消息体属性对象:[48]
        版本号:[JT2013]
        [bit15] [0]
        [bit14] 协议版本标识:[0]
        [bit13] 是否分包:[false]
        [bit10-12] 加密标识:[0] 0-不加密 1-RSA
        [bit0-bit9] 消息体长度:[48]
}
[000000000001] 终端手机号:[1]
[0001] 消息流水号:[1]
数据体对象:{
        终端-注册:[001f006e63643132337777772e3830382e636f6d0000000000000000003736353433323101b2e2413132333435363738]
        [001f] 省域ID:[31]
        [006e] 市县域ID:[110]
        [6364313233] 制造商ID(5):[cd123]
        [7777772e3830382e636f6d000000000000000000] 终端型号(20):[www.808.com]
        [37363534333231] 终端ID(7):[7654321]
        [01] 车牌颜色:[1]
        [b2e2413132333435363738] 车牌号:[测A12345678]
}
```

#### 1.2 自定义协议扩展 (附加信息)

自定义解析扩展 0x33为例 关键代码如下
``` go
type Location struct {
	model.T0x0200
	customMile  int
	customValue uint8
}

func (l *Location) Parse(jtMsg *jt808.JTMessage) error {
	l.T0x0200AdditionDetails.CustomAdditionContentFunc = func(id uint8, content []byte) (model.AdditionContent, bool) {
		if id == uint8(consts.A0x01Mile) {
			l.customMile = 100
		}
		if id == 0x33 {
			value := content[0]
			l.customValue = value
			return model.AdditionContent{
				Data:        content,
				CustomValue: value,
			}, true
		}
		return model.AdditionContent{}, false
	}
	return l.T0x0200.Parse(jtMsg)
}
```

部分输出 [输出详情](./example/protocol/README.md#custom)
``` txt
里程[11] 自定义辅助里程[100]
自定义未知信息扩展 32 32
```

#### 1.3 平台下发给终端参数 (8104查询终端参数)

关键代码如下
``` go
	replyMsg := goJt808.SendActiveMessage(&service.ActiveMessage{
		Key:              phone,                           // 默认使用手机号作为唯一key 根据key找到对应终端的TCP链接
		Command:          consts.P8104QueryTerminalParams, // 下发的指令
		Body:             nil,                             // 下发的body数据 8104为空
		OverTimeDuration: 3 * time.Second,                 // 超时时间 设备这段时间没有回复则失败
	})
	var t0x0104 model.T0x0104
	if err := t0x0104.Parse(replyMsg.JTMessage); err != nil {
		panic(err)
	}
	fmt.Println(t0x0104.String())
```

部分输出 [输出详情](./example/protocol/README.md#active_reply)
``` txt
数据体对象:{
        [0003] 应答消息流水号:[3]
        [5b] 应答参数个数:[91]
        终端-查询参数:

        {
                [0001]终端参数ID:1 终端心跳发送间隔,单位为秒(s)
                参数长度[4] 是否存在[true]
                [0000000a]参数值:[10]
        }
		...
        {
                [0110]终端参数ID:272 CAN总线ID单独采集设置:
                参数长度[8] 是否存在[true]
                [0000000000000101]参数值:[[0 0 0 0 0 0 1 1]]
        }
        未知终端参数id:[33 117 118 119 121 122 123 124]
}
```

### 2. 自定义保存经纬度
[详情点击](./example/simulator/server/main.go)
- 自定义实现0x0200的消息处理 把数据发送到nats 关键代码如下
``` go
type T0x0200 struct {
	model.T0x0200
}

func (t *T0x0200) OnReadExecutionEvent(message *service.Message) {
	var t0x0200 model.T0x0200
	if err := t0x0200.Parse(message.JTMessage); err != nil {
		fmt.Println(err)
		return
	}
	location := shared.NewLocation(message.Header.TerminalPhoneNo, t0x0200.Latitude, t0x0200.Longitude)
	if err := mq.Default().Pub(shared.SubLocation, location.Encode()); err != nil {
		fmt.Println(err)
		return
	}
}

func (t *T0x0200) OnWriteExecutionEvent(_ service.Message) {}
```

-  使用自定义0x0200消息处理启动

``` go
	goJt808 := service.New(
		service.WithHostPorts("0.0.0.0:8080"),
		service.WithNetwork("tcp"),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				consts.T0200LocationReport:      &T0x0200{},
			}
		}),
	)
	goJt808.Run()
```

### 3. 主动安全附件

附件服务调用方式
``` go
	attach := attachment.New(
		attachment.WithNetwork("tcp"),
		attachment.WithHostPorts(address),
		attachment.WithFileEventerFunc(func() attachment.FileEventer {
			f, _ := os.OpenFile("file.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
			return &meFileEvent{file: f} // 自定义附件事件
		}),
	)
	attach.Run()

```
#### 3.1 苏标
[详情点击](./example/su_biao/main.md)
- 自定义0x0200指令 自定义苏标相关附加指令扩展 获取报警标识号
``` go
type meLocation struct {
	model.T0x0200
	model.T0x0200AdditionExtension0x64
	model.T0x0200AdditionExtension0x65
	model.T0x0200AdditionExtension0x66
	model.T0x0200AdditionExtension0x67
	model.T0x0200AdditionExtension0x70
}

func (l *meLocation) Parse(jtMsg *jt808.JTMessage) error {
	l.T0x0200.CustomAdditionContentFunc = func(id uint8, content []byte) (model.AdditionContent, bool) {
		switch id {
		case 0x64:
			return l.T0x0200AdditionExtension0x64.Parse(id, content)
		case 0x65:
			return l.T0x0200AdditionExtension0x65.Parse(id, content)
		case 0x66:
			return l.T0x0200AdditionExtension0x66.Parse(id, content)
		case 0x67:
			return l.T0x0200AdditionExtension0x67.Parse(id, content)
		case 0x70:
			return l.T0x0200AdditionExtension0x70.Parse(id, content)
		}
		return model.AdditionContent{}, false
	}
	return l.T0x0200.Parse(jtMsg)
}

```

### 4. jt1078相关

#### 4.1 流媒体服务使用lal
[详情点击](./example/jt1078/lal/main.go)
-  把1078格式流转对应格式 放入lal服务中 核心代码参考

``` go
func (j *jt1078) createStream(name string) chan<- *Packet {
	...
	ch := make(chan *Packet, 100)
	go func(session logic.ICustomizePubSessionContext, ch <-chan *Packet) {
		for v := range ch {
				...
				switch v.Flag.PT {
				case PTG711A:
					tmp.PayloadType = base.AvPacketPtG711A
				case PTG711U:
					tmp.PayloadType = base.AvPacketPtG711U
				case PTH264:
				case PTH265:
					tmp.PayloadType = base.AvPacketPtHevc
				default:
					slog.Warn("未知类型",
						slog.Any("pt", v.Flag.PT))
				}
				if err := session.FeedAvPacket(tmp); err != nil {
					slog.Warn("session.FeedAvPacket",
						slog.Any("err", err))
				}
			}
		}
	}(session, ch)
}
```

#### 4.2 流媒体服务使用rtvs
[详情点击](./example/jt1078/rtvs/main.go)

``` go
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

```

## 协议对接完成情况
### JT808 终端通讯协议消息对照表

| 序号  |    消息 ID    | 完成情况 |  测试情况  | 消息体名称                     |  2019 版本   | 2011 版本 |
| :---: | :-----------: | :------: | :--------: | :----------------------- | :----------: | :-------: |
|   1   |    0x0001     |    ✅    |     ✅     | 终端通用应答				|				|			|
|   2   |    0x8001     |    ✅    |     ✅     | 平台-通用应答				|				|           |
|   3   |    0x0002     |    ✅    |     ✅     | 终端心跳					|				|           |
|   5   |    0x0100     |    ✅    |     ✅     | 终端注册					|     修改		|  被修改	|
|   4   |    0x8003     |    ✅    |     ✅     | 补传分包请求                |               |  被新增    |
|   6   |    0x8100     |    ✅    |     ✅     | 平台-注册应答				|				|           |
|   8   |    0x0102     |    ✅    |     ✅     | 终端鉴权					|     修改		|			|
|   9   |    0x8103     |    ✅    |     ✅     | 设置终端参数                |  修改且增加  	|  被修改    |
|  10   |    0x8104     |    ✅    |     ✅     | 平台-查询终端参数			|				|           |
|  11   |    0x0104     |    ✅    |     ✅     | 查询终端参数应答			|				|           |
|  18   |    0x0200     |    ✅    |     ✅     | 位置信息汇报				| 增加附加信息 	|  被修改	|
|  49   |    0x0704     |    ✅    |     ✅     | 定位数据批量上传			|     修改		|  被新增	|
|  51   |    0x0800     |    ✅    |     ✅     | 多媒体事件信息上传           |              |  被修改   |
|  52   |    0x0801     |    ✅    |     ✅     | 多媒体数据上传               |     修改     |  被修改   |
|  53   |    0x8800     |    ✅    |     ✅     | 平台-多媒体数据上传应答       |              |  被修改   |
|  54   |    0x8801     |    ✅    |     ✅     | 平台-摄像头立即拍摄命令       |     修改     |           |
|  55   |    0x0805     |    ✅    |     ✅     | 摄像头立即拍摄命令应答        |     修改     |  被新增   |

### JT1078 扩展

| 序号  |    消息 ID     | 完成情况 	| 测试情况 | 消息体名称 |
| :---: | :-----------: | :------: | :--------: | :----------------------- |
|  13   |    0x1003     |    ✅    |    ✅    | 终端上传音视频属性       |
|  14   |    0x1005     |    ✅    |    ✅    | 终端上传乘客流量         |
|  15   |    0x1205     |    ✅    |    ✅    | 终端上传音视频资源列表   |
|  16   |    0x1206     |    ✅    |    ✅    | 文件上传完成通知         |
|  17   |    0x9003     |    ✅    |    ✅    | 平台-查询终端音视频属性       |
|  18   |    0x9101     |    ✅    |    ✅    | 平台-实时音视频传输请求       |
|  19   |    0x9102     |    ✅    |    ✅    | 平台-音视频实时传输控制       |
|  20   |    0x9105     |    ✅    |    ✅    | 平台-实时音视频传输状态通知   |
|  21   |    0x9201     |    ✅    |    ✅    | 平台-下发远程录像回放请求 |
|  22   |    0x9202     |    ✅    |    ✅    | 平台-下发远程录像回放控制 |
|  23   |    0x9205     |    ✅    |    ✅    | 平台-查询资源列表             |
|  24   |    0x9206     |    ✅    |    ✅    | 平台-文件上传指令             |
|  25   |    0x9207     |    ✅    |    ✅    | 平台-文件上传控制             |

### 主动安全（苏标）扩展

| 序号  |    消息 ID    | 完成情况 | 测试情况 | 消息体名称                 |
| :---: | :-----------: | :------: | :------: | :------------------------- |
|   1   |    0x1210     |    ✅    |    ✅    | 报警附件信息消息           |
|   2   |    0x1211     |    ✅    |    ✅    | 文件信息上传               |
|   3   |    0x1212     |    ✅    |    ✅    | 文件上传完成消息           |
|   4   |    0x9208     |    ✅    |    ✅    | 报警附件上传指令           |
|   5   |    0x9212     |    ✅    |    ✅    | 文件上传完成消息应答       |