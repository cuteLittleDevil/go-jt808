# GB28181 客户端示例

本目录演示使用 `github.com/cuteLittleDevil/go-jt808/gb28181` 模拟国标（GB/T 28181）设备端，实现向国标信令平台注册、心跳保活、目录查询以及点播推流（TCP 被动模式）。

## 目录结构

| 子目录 | 说明 | 配置文件 |
| --- | --- | --- |
| [client](./client) | 单设备 GB28181 客户端，收到 Invite 点播后读取本地 JT1078 测试文件并转封装为 PS 流推送 | [client/config.yaml](./client/config.yaml) |
| [rtsp](./rtsp) | 自定义 `ToGB28181er` 实现示例，通过 ZLMediaKit 接口将 RTSP 流拉流代理并转为国标 RTP/PS 推送 | [rtsp/config.yaml](./rtsp/config.yaml) |
| [simulator](./simulator) | 批量 GB28181 设备模拟器，支持并发创建多台设备注册、保活与点播压测 | [simulator/config.yaml](./simulator/config.yaml) |
| [testdata](./testdata) | JT1078 模拟码流文件（`data.txt` 为纯 H264，`audio_data.txt` 为 H264 + G711A） | - |

## 前置条件

1. 已部署可用的 GB28181 信令与流媒体平台（如 `m7s`、`wvp-GB28181-pro`、`gowvp/gb28181` 等）。
2. 修改对应子目录的 `config.yaml`，将 `gb28181.platform` 下的 `domain`、`id`、`password`、`ip`、`port` 替换为你的平台参数。
3. 如使用 `rtsp` 示例，还需配置 `zlm` 下的 `secret`、RTSP 地址及 ZLMediaKit HTTP API 地址。

## 运行方式

### 1. 单设备 JT1078 转 PS 点播

```bash
cd example/gb28181/client
go run .
```

### 2. RTSP 经 ZLMediaKit 转 GB28181 推送

```bash
cd example/gb28181/rtsp
go run .
```

### 3. 批量 GB28181 设备并发模拟

按需调整 `simulator/config.yaml` 中的 `device.maxID`、`createIntervalMicrosecond`、`stopSecond` 后运行：

```bash
cd example/gb28181/simulator
go run .
```

> 说明：如需将真实 JT808 设备通过适配器无侵入接入 GB28181 平台，请参考 [example/jt808_to_gb28181](../jt808_to_gb28181/README.md)。
