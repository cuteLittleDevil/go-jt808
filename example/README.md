# go-jt808 示例总览

本目录包含 `go-jt808` 各核心能力与生产场景的可运行示例。每个示例（或子目录）均为独立的 Go 模块，可进入对应目录直接运行或构建。

## 当前依赖版本说明

除历史压测示例 `simulator` 外，所有示例均已对齐仓库最新子模块版本：

| 子模块 | 版本 | 说明 |
| --- | --- | --- |
| `protocol` | `v1.18.2` | JT/T 808 / 1078 / 主动安全协议编解码 |
| `service` | `v1.15.0` | JT808 服务端与会话管理 |
| `shared` | `v1.6.0` | 公共常量与枚举 |
| `terminal` | `v1.4.0` | 终端模拟器与报文构造 |
| `attachment` | `v0.5.0` | 主动安全附件上传服务（苏标等） |
| `adapter` | `v0.4.0` | 多 808 服务流量适配与指令放行 |
| `gb28181` | `v0.5.2` | GB28181 模拟客户端与 JT1078 转 PS 流 |

> 注：`example/simulator` 为历史性能基准测试示例，按原测试基线保留在 `service v0.4.0` / `terminal v0.2.0`（`v0.x.x` 系列），未升级至 `service v1.x.x`。

## 示例导航

| 目录 | 场景说明 | 前置依赖 | 文档 |
| --- | --- | --- | --- |
| [quick_start](./quick_start) | 最简入门：启动 808 服务、模拟终端连接、解析 `0x0200` 位置上报与超时断开 | 无 | [README](./quick_start/README.md) |
| [protocol](./protocol) | 协议解析、自定义附加信息、主动下发与回复映射、自定义分包、分包补传等 8 个子示例 | 无（部分需外部模拟器） | [README](./protocol/README.md) |
| [adapter](./adapter) | 适配器方案：单设备连接同时转发到多个 808 服务（Leader / Follower 指令放行控制） | 无 | [README](./adapter/README.md) |
| [attachment](./attachment) | 主动安全（苏标）附件上传服务，含本地模拟、苏标对接、带 HTTP 回调的独立附件服务 | 无 | [README](./attachment/README.md) |
| [client](./client) | 多客户端并发压测工具，支持 YAML 配置心跳/位置上报周期与 HTTP 统计查询 | JT808 服务端 | [README](./client/README.md) |
| [distributed_cluster](./distributed_cluster) | 分布式集群方案：Nginx 四层负载均衡 + 多节点 808 服务 + NATS 广播下发指令 | Nginx、NATS | [README](./distributed_cluster/README.md) |
| [ftp](./ftp) | 平台下发 `0x9205` / `0x9206` 指令，终端上传文件到 FTP 服务 | FTP 服务 | [README](./ftp/README.md) |
| [gb28181](./gb28181) | GB28181 客户端模拟（JT1078 转 PS 推流、RTSP 转 PS、批量设备压测） | GB28181 信令平台（如 m7s / wvp / gowvp） | [README](./gb28181/README.md) |
| [jt1078](./jt1078) | JT1078 音视频推流与播放，对接 RTVS、LAL、sky-java、m7s、ZLMediaKit、SRS | 对应流媒体服务 | [README](./jt1078/README.md) |
| [jt808_to_gb28181](./jt808_to_gb28181) | 无侵入式 JT808 转 GB28181 方案（适配器 + JT1078 转 PS），支持 Docker 部署 | GB28181 信令平台 | [README](./jt808_to_gb28181/README.md) |
| [simulator](./simulator) | 10 万连接与 1 亿经纬度入库性能基准测试（保留 `service v0.4.0`） | NATS、TDengine | [README](./simulator/README.md) |
| [web](./web) | 完整生产级 Web 工程：HTTP API + 808 服务 + 附件 + 告警轨迹 + 通知订阅 | NATS（可选 MinIO / MongoDB / TDengine） | [README](./web/README.md) |

## 运行方式

进入对应示例目录直接执行 `go run` 或 `go build` 即可（建议 Go 1.23+，部分依赖较新第三方包的示例需 Go 1.25+）：

```bash
cd example/quick_start
go run .
```
