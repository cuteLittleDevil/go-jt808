![MIT License](https://img.shields.io/github/license/cuteLittleDevil/go-jt808)
[![Go Doc](https://godoc.org/github.com/cuteLittleDevil/go-jt808?status.svg)](https://pkg.go.dev/github.com/cuteLittleDevil/go-jt808#readme-jt808)
[![Perf](https://img.shields.io/badge/perf-save-blue.svg)](https://github.com/cuteLittleDevil/go-jt808/blob/main/example/simulator/README.md)
[![WEB](https://img.shields.io/badge/example-web-red.svg)](https://github.com/cuteLittleDevil/go-jt808/tree/main/example/web#web)
[![API](https://img.shields.io/badge/web%20doc-apifox-red.svg)](https://vsh9jdgg5d.apifox.cn/)
[![codecov](https://codecov.io/github/cuteLittleDevil/go-jt808/graph/badge.svg?token=KZXKKIJUSA)](https://codecov.io/github/cuteLittleDevil/go-jt808)
[![Go Report Card](https://goreportcard.com/badge/github.com/cuteLittleDevil/go-jt808/protocol)](https://goreportcard.com/report/github.com/cuteLittleDevil/go-jt808/protocol)
[![build status](https://github.com/cuteLittleDevil/go-jt808/actions/workflows/ci.yml/badge.svg)](https://github.com/cuteLittleDevil/go-jt808/actions/workflows/ci.yml)

# go-jt808

- 设计说明 https://dkpt1fpoxb.feishu.cn/docx/FUlPda09roSnN0x7SJAc7yPbnke

``` txt
 本项目已更好支持二次开发为目标 可通过各种自定义事件去完成相应功能
 看飞哥的单机TCP百万并发 好奇有数据情况的表现 因此国庆准备试一试有数据的情况

```

---
- 性能测试 单机[2核4G机器]并发10w+ 每日保存4亿+经纬度 [详情](./README.md#save)
- 支持JT808(2011/2013/2019) JT1078(需要其他流媒体服务)
- 支持分包和自动补传 支持主动安全扩展(苏标 黑标 广东标 湖南标 四川标)

| 特点  |   描述   |
| :---:   | -------- |
|  安全可靠 | 核心协议部分测试覆盖率100% 纯原生go实现(不依赖任何库)  |
|  简洁优雅 | 核心代码不到1000行 不使用任何锁 仅使用channel完成  |
|  易于扩展 | 方便二次开发 有适配任意jt808服务 分布式集群等案例  |

| 包名 |      描述       | Go Report Card |
|----------|--------------------|-----|
| shared | jt808和1078指令常量 | [![Go Report Card](https://goreportcard.com/badge/github.com/cuteLittleDevil/go-jt808/shared)](https://goreportcard.com/report/github.com/cuteLittleDevil/go-jt808/shared) |
| protocol | jt808和1078协议实现 | [![Go Report Card](https://goreportcard.com/badge/github.com/cuteLittleDevil/go-jt808/protocol)](https://goreportcard.com/report/github.com/cuteLittleDevil/go-jt808/protocol) |
| service | jt808服务端 | [![Go Report Card](https://goreportcard.com/badge/github.com/cuteLittleDevil/go-jt808/service)](https://goreportcard.com/report/github.com/cuteLittleDevil/go-jt808/service) |
| adapter | jt808适配器 | [![Go Report Card](https://goreportcard.com/badge/github.com/cuteLittleDevil/go-jt808/adapter)](https://goreportcard.com/report/github.com/cuteLittleDevil/go-jt808/adapter) |
| attachment | jt808附件服务 | [![Go Report Card](https://goreportcard.com/badge/github.com/cuteLittleDevil/go-jt808/attachment)](https://goreportcard.com/report/github.com/cuteLittleDevil/go-jt808/attachment) |
| terminal | jt808客户端模拟器 | [![Go Report Card](https://goreportcard.com/badge/github.com/cuteLittleDevil/go-jt808/terminal)](https://goreportcard.com/report/github.com/cuteLittleDevil/go-jt808/terminal) |
| gb28181 | gb28181客户端 | [![Go Report Card](https://goreportcard.com/badge/github.com/cuteLittleDevil/go-jt808/gb28181)](https://goreportcard.com/report/github.com/cuteLittleDevil/go-jt808/gb28181) |

---

## 常见案例

### 1. 真实项目对接 [apifox文档](https://vsh9jdgg5d.apifox.cn/) [web详情](./example/web) [releases下载](https://github.com/cuteLittleDevil/go-jt808/releases)
``` txt
web例子在线网页 http://124.221.30.46:18000/
真实案例 根据壹品信息技术有限公司对接中农云设备修改
```

### 2. jt1078视频 [详情](./example/jt1078/README.md)

``` txt
平台下发0x9101等指令 模拟器开始推流等动作
```

| 流媒体 | 语言 | 描述       | 说明 |
|----------|-----|-------------------|-----|
| rtvs | 不开源<br/> 可使用  | 在线测试页面 https://go-jt808.online:44300/index.html <br/> 点击实时视频(0x9101)按钮播放| [详情点击](./example/jt1078/README.md#rtvs)  |
| lal | go  | 在线播放地址 http://go-jt808.online:8080/live/1001_1.flv | [详情点击](./example/jt1078/README.md#lal)  |
| sky-java | java  | 需要部署后 HTTP请求 10秒内拉流 参考格式如下 <br/> http://222.244.144.181:7777/video/1001-1-0-0.live.mp4 | [详情点击](./example/jt1078/README.md#sky-java)  |
| monibuca | go  | 对讲示例 https://go-jt808.online:12000 | [详情点击](https://github.com/cuteLittleDevil/m7s-jt1078)  |
| ZLMediaKit | c++  | 对讲测试 https://go-jt808.online/static/?type=push <br/> http://go-jt808.online:80/rtp/000000001003_1_0_0.live.mp4 | [详情点击](./example/jt1078/README.md#zlm)  |

### 3. jt808模拟gb28181客户端 [gb28181使用](./gb28181/example_test.go) [jt1078转ps流](./gb28181/internal/stream/jt1078_to_gb28181.go)
``` txt
原: 设备连接到原808服务
现: 设备连接到适配器 适配器产生两个模拟链接 一个连接到原808服务 保证不影响原服务
另一个连接到gb28181模拟服务 产生一个gb28181客户端 (目前仅支持注册 目录查询 点播[jt1078转ps流])
```

| 信令服务 | 流媒体 | 在线测试  |  说明 |
|----------|-----|-------------------| --- |
| monibuca | monibuca  | http://101.35.2.3:12079/#/0/device/gb28181 | [详情](./example/jt808_to_gb28181/README.md) |
| gb28181 | ZLMediaKit  |   | [详情](./example/jt808_to_gb28181/README.md#gb28181) |
| wvp-GB28181-pro | ZLMediaKit  |   | [详情](./example/jt808_to_gb28181/README.md#wvp) |

- 参考配置 [请点击](./example/jt808_to_gb28181/README.md#config)

```
docker pull cdcddcdc/jt808-to-gb28181:latest
```
```
docker run -d \
-v /home/config.yaml:/app/jt808-to-gb28181/config.yaml \
--network host \
cdcddcdc/jt808-to-gb28181:latest
```

### 4. 兼容任意808服务 [详情](./example/adapter/README.md)
``` txt
真实设备连接到适配器 适配器产生多个模拟设备连接多个808服务
```

### 5. 主动安全附件 [流程](./example/attachment/README.md#主动安全)
``` txt
默认支持苏标 可自定义各事件扩展（开始、传输进度、补传情况、完成、退出等事件）
```

### 6. 存储经纬度 [详情](./README.md#save)
``` txt
jt808服务端 模拟器 消息队列 数据库都运行在2核4G腾讯云服务器
测试每秒保存5000条的情况 约5.5小时保存了近1亿的经纬度
```

### 7. 分布式集群方案 [详情](./example/distributed_cluster/README.md)
``` txt
使用nginx把终端分配到多个808服务上 下发数据使用广播
存在则回复终端应答到新主题 不存在则忽略
```

### 8. 平台下发指令给终端 [获取参数](./example/protocol/active_reply/main.go) [立即拍摄](./example/protocol/camera/main.go)
``` txt
主动下发给设备指令 获取应答的情况
```

### 9. 协议交互详情 [代码参考](./example/protocol/register/register_test.go)
``` txt
使用自定义模拟器 可以轻松生成测试用的报文 有详情描述
可在apifox文档页面 使用测试环境查看报文详情 https://vsh9jdgg5d.apifox.cn/250573462e0
```

### 10. 自定义协议扩展 [代码参考](./example/protocol/custom_parse/main.go)
``` txt
自定义附加信息处理 获取想要的扩展内容
```

### 11. ftp例子 [详情](./example/ftp/README.md)
``` txt
把atop_cpu.png传输到ftp目录 (需要ftp服务)
```

[快速开始](./example/quick_start) [完整项目例子](./example/web)
``` go
package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/cuteLittleDevil/go-jt808/terminal"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"
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
	goJt808 := service.New(
		service.WithHostPorts("0.0.0.0:808"),
		service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
			return map[consts.JT808CommandType]service.Handler{
				consts.T0200LocationReport: &meLocation{}, // 自定义0x0200位置解析等
			}
		}),
	)
	go client("1001", "127.0.0.1:808") // 模拟一个设备连接
	goJt808.Run()
}

type meLocation struct {
	model.T0x0200
}

func (l *meLocation) OnReadExecutionEvent(msg *service.Message) {
	_ = l.Parse(msg.JTMessage)
	fmt.Println(time.Now().Format(time.DateTime), l.String()) // 打印经纬度等信息
}

func (l *meLocation) OnWriteExecutionEvent(_ service.Message) {}

func (l *meLocation) String() string {
	body := l.T0x0200.Encode()
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", l.Protocol(), body),
		l.T0x0200LocationItem.String(),
		l.AlarmSignDetails.String(),
		l.StatusSignDetails.String(),
		"}",
	}, "\n")
}

func client(phone string, address string) {
	time.Sleep(time.Second)
	t := terminal.New(terminal.WithHeader(consts.JT808Protocol2013, phone))
	location := t.CreateDefaultCommandData(consts.T0200LocationReport)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		_, _ = conn.Write(location)
	}
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
|  v0.3.0 | 连接数测试  | 10w+ |  2核4G | 120%+cpu 1.7G内存  | 10.0.16.5: 服务端和模拟器  <br/> 10.0.16.14: 模拟器 |

<h3 id="save"> 模拟经纬度存储测试 </h3>

[详情点击](./example/simulator/README.md#save)

- save进程丢失了部分数据 channel队列溢出抛弃 (测试channel队列为100)
- 保存1亿丢失826条 保存4.32亿丢失1216条（分两次测试)

| 服务端版本  | 客户端 |  服务器配置  |  描述  |
| :---:   | :--: | :------: | :----------------------------: |
|  v0.3.0 | 1w go模拟器 |  2核4G | 每秒5000 一共保存经纬度1亿  <br/> 实际保存99999174 成功率99.999% |

| 服务  |   cpu   | 内存 | 描述 |
| :---:   | :-------: | :--: | :--: |
|  server | 35% | 180.4MB | 808服务端 |
|  client | 23% | 196MB | 模拟客户端 |
|  save |  18% | 68.8MB | 存储数据服务 |
|  nats-server | 20% | 14.8MB | 消息队列 |
|  taosadapter | 37% | 124.3MB | tdengine数据库适配 |
|  taosd | 15% | 124.7MB | tdengine数据库 |

## 协议对接完成情况
### JT808终端通讯协议

| 序号  |    消息 ID    | 完成情况 |  测试情况  | 消息体名称       |  2019 版本   | 2011 版本 |
| :---: | :-----------: | :------: | :--------: | :------------------------------------------------------------ | :----------: | :-------: |
|   1   |    0x0001     |    ✅    |     ✅     | [终端通用应答](./protocol/model/t_0x0001.go#L12) 				|       		|  			|
|   2   |    0x8001     |    ✅    |     ✅     | [平台-通用应答](./protocol/model/p_0x8001.go#L12) 				| 				|   		|
|   3   |    0x0002     |    ✅    |     ✅     | [终端心跳](./protocol/model/t_0x0002.go#L9) 					|			    |           |
|   4   |    0x8003     |    ✅    |     ✅     | [补传分包请求](./protocol/model/p_0x8003.go#L12)  				|               |  被新增    |
|   5   |    0x0100     |    ✅    |     ✅     | [终端注册](./protocol/model/t_0x0100.go#L14)					|     修改		|  被修改	|
|   6   |    0x8100     |    ✅    |     ✅     | [平台-注册应答](./protocol/model/p_0x8100.go#L13)				|				|           |
|   8   |    0x0102     |    ✅    |     ✅     | [终端鉴权](./protocol/model/t_0x0102.go#L12)					|     修改		|			|
|   9   |    0x8103     |    ✅    |     ✅     | [平台-设置终端参数](./protocol/model/p_0x8103.go#L11)            |  修改且增加  	|  被修改    |
|  10   |    0x8104     |    ✅    |     ✅     | [平台-查询终端参数](./protocol/model/p_0x8104.go#L10)			|				|           |
|  11   |    0x0104     |    ✅    |     ✅     | [查询终端参数应答](./protocol/model/t_0x0104.go#L12)			|				|           |
|  18   |    0x0200     |    ✅    |     ✅     | [位置信息汇报](./protocol/model/t_0x0200.go#L10)				 | 增加附加信息 	|  被修改	|
|  19   |    0x8201     |    ✅    |     ✅     | [平台-位置信息查询](./protocol/model/p_0x8201.go#L10)            |              |           |
|  20   |    0x0201     |    ✅    |     ✅     | [位置信息查询应答](./protocol/model/t_0x0201.go#L12)             |              |           |
|  21   |    0x8202     |    ✅    |     ✅     | [平台-临时位置跟踪控制](./protocol/model/p_0x8202.go#L12)         |              |           |
|  23   |    0x8300     |    ✅    |     ✅     | [平台-文本信息下发](./protocol/model/p_0x8300.go#L13)            |     修改      |  被修改   |
|  26   |    0x8302     |    ✅    |     ✅     | [平台-提问下发](./protocol/model/p_0x8302.go#L14)               |     删除      |           |
|  27   |    0x0302     |    ✅    |     ✅     | [提问应答](./protocol/model/t_0x0302.go#L12)                   |     删除      |           |
|  49   |    0x0704     |    ✅    |     ✅     | [定位数据批量上传](./protocol/model/t_0x0704.go#L13)			|     修改		|  被新增	|
|  51   |    0x0800     |    ✅    |     ✅     | [多媒体事件信息上传](./protocol/model/t_0x0800.go#L12)           |              |  被修改   |
|  52   |    0x0801     |    ✅    |     ✅     | [多媒体数据上传](./protocol/model/t_0x0801.go#L12)               |     修改     |  被修改   |
|  53   |    0x8800     |    ✅    |     ✅     | [平台-多媒体数据上传应答](./protocol/model/p_0x8800.go#L12)       |              |  被修改   |
|  54   |    0x8801     |    ✅    |     ✅     | [平台-摄像头立即拍摄命令](./protocol/model/p_0x8801.go#L12)       |     修改     |           |
|  55   |    0x0805     |    ✅    |     ✅     | [摄像头立即拍摄命令应答](./protocol/model/t_0x0805.go#L12)        |     修改     |  被新增   |

### JT1078扩展

| 序号  |    消息 ID     | 完成情况 	| 测试情况 | 消息体名称 |
| :---: | :-----------: | :------: | :--------: | :----------------------- |
|  13   |    0x1003     |    ✅    |    ✅    | [终端上传音视频属性](./protocol/model/t_0x1003.go#L12)         |
|  14   |    0x1005     |    ✅    |    ✅    | [终端上传乘客流量](./protocol/model/t_0x1005.go#L13)           |
|  15   |    0x1205     |    ✅    |    ✅    | [终端上传音视频资源列表](./protocol/model/t_0x1205.go#L14)      |
|  16   |    0x1206     |    ✅    |    ✅    | [文件上传完成通知](./protocol/model/t_0x1206.go#L12)           |
|  17   |    0x9003     |    ✅    |    ✅    | [平台-查询终端音视频属性](./protocol/model/p_0x9003.go#L10)     |
|  18   |    0x9101     |    ✅    |    ✅    | [平台-实时音视频传输请求](./protocol/model/p_0x9101.go#L13)     |
|  19   |    0x9102     |    ✅    |    ✅    | [平台-音视频实时传输控制](./protocol/model/p_0x9102.go#L11)     |
|  20   |    0x9105     |    ✅    |    ✅    | [平台-实时音视频传输状态通知](./protocol/model/p_0x9105.go#L11)  |
|  21   |    0x9201     |    ✅    |    ✅    | [平台-下发远程录像回放请求](./protocol/model/p_0x9201.go#L13)    |
|  22   |    0x9202     |    ✅    |    ✅    | [平台-下发远程录像回放控制](./protocol/model/p_0x9202.go#L12)    |
|  23   |    0x9205     |    ✅    |    ✅    | [平台-查询资源列表](./protocol/model/p_0x9205.go#L13)           |
|  24   |    0x9206     |    ✅    |    ✅    | [平台-文件上传指令](./protocol/model/p_0x9206.go#L13)           |
|  25   |    0x9207     |    ✅    |    ✅    | [平台-文件上传控制](./protocol/model/p_0x9207.go#L12)           |

### 主动安全扩展

| 序号  |    消息 ID    | 完成情况 | 测试情况 | 消息体名称  |
| :---: | :-----------: | :------: | :------: | :------------------------- |
|   1   |    0x1210     |    ✅    |    ✅    | [报警附件信息消息](./protocol/model/t_0x1210.go#L15)     |
|   2   |    0x1211     |    ✅    |    ✅    | [文件信息上传](./protocol/model/t_0x1211.go#L12)         |
|   3   |    0x1212     |    ✅    |    ✅    | [文件上传完成消息](./protocol/model/t_0x1212.go#L8)       |
|   4   |    0x9208     |    ✅    |    ✅    | [报警附件上传指令](./protocol/model/p_0x9208.go#L15)      |
|   5   |    0x9212     |    ✅    |    ✅    | [文件上传完成消息应答](./protocol/model/p_0x9212.go#L13)   |