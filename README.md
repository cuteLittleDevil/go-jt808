# go-jt808

- 本项目已更好支持二次开发为目标 核心是协议的解析交互  <br/>
通过各种自定义事件去完成相应功能 如记录消息、主动发送消息(支持等待设备回复)等
---

- 看飞哥的单机TCP百万并发 好奇有数据情况的表现 因此国庆准备试一试有数据的情况
- 性能测试 每日保存亿+经纬度[2核4G机器] [详情](#模拟经纬度存储测试)
- 安全可靠 核心协议交互不基于任何框架完成 测试覆盖率100%
- 简洁优雅 信令交互不使用任何锁 仅使用channel完成
- 支持JT808(2011/2013/209)
---

## 快速开始
``` go
package main

import (
	"github.com/cuteLittleDevil/go-jt808/service"
	"log/slog"
	"os"
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
		service.WithHostPorts("0.0.0.0:8080"),
		service.WithNetwork("tcp"),
	)
	goJt808.Run()
}

```

---
- 目前(2024-10-01)前的go语言版本个人觉得都不好因此都不推荐参考 推荐参考资料如下
## 参考项目
| 项目名称           | 语言   | 日期       | Star 数 | 链接                                       |
|--------------------|--------|------------|---------|--------------------------------------------|
| JT808           | C#     | 2024-10-01 | 534       | [JT808 C#](https://github.com/SmallChi/JT808.git) |
| jt808-server    | Java   | 2024-10-01 | 1.4k+     | [JT808 Java](https://gitee.com/yezhihao/jt808-server.git) |

| 描述                | 链接                         |
|--------------------|------------------------------|
| 飞哥的开发内功修炼    | https://github.com/yanfeizhang/coder-kung-fu?tab=readme-ov-file |
| 协议文档       | https://gitee.com/yezhihao/jt808-server/tree/master/协议文档 |
| 协议解析网站  | https://jttools.smallchi.cn/jt808 |
| bcd转dec编码   | https://github.com/deatil/lakego-admin/tree/main/pkg/lakego-pkg/go-encoding/bcd |

## 性能测试
- java模拟器(QQ群下载 373203450)
- go模拟器 [详情点击](./example/simulator/client/client.go#go模拟器)

### 连接数测试 [详情点击](./example/simulator/README.md#连接数测试)

| 服务端版本  |   客户端   | 并发数 |  服务器配置  | jt808服务使用资源情况 |
| :---:   | :-------: | :--: | :------: | :-------------- |
|  v0.3.0 | 10w+ go模拟器  | 10w+ |  10核32G | 20%cpu 1.4G内存  |

### 模拟经纬度存储测试 [详情点击](./example/simulator/README.md#模拟经纬度存储测试)
- 1w个客户端 每一个客户端发送100个0x0200

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

### 1. 协议解析

### 2. 自定义接收发送数据

### 3. jt808附件上传

### 4. jt1078相关

## 协议对接完成情况

| 序号  |    消息 ID    | 完成情况 |  测试情况  | 消息体名称                     |  2019 版本   | 2011 版本 |
| :---: | :-----------: | :------: | :--------: | :----------------------------- | :----------: | :-------: |
|   1   |    0x0001     |    ✅    |     ✅     | 终端通用应答                   |              |           |
|   2   |    0x8001     |    ✅    |     ✅     | 平台-通用应答                   |              |           |
|   3   |    0x0002     |    ✅    |     ✅     | 终端心跳                       |              |           |
|   5   |    0x0100     |    ✅    |     ✅     | 终端注册                       |     修改     |  被修改   |
|   6   |    0x8100     |    ✅    |     ✅     | 平台-注册应答                   |              |           |
|   8   |    0x0102     |    ✅    |     ✅     | 终端鉴权                       |     修改     |
|  18   |    0x0200     |    ✅    |     ✅     | 位置信息汇报                   | 增加附加信息 |  被修改   |
|  49   |    0x0704     |    ✅    |     ✅     | 定位数据批量上传               |     修改     |  被新增   |
