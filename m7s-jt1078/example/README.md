<h2 id="m7s"> m7s </h2>

- [m7s官方地址](https://monibuca.com)
- [部署文档参考](https://blog.csdn.net/vanjoge/article/details/108319078)
- [代码参考](./main.go)

目前v5版本还没有发布 需要把monibuca项目拉入到本地 使用go.work

```
.
├── example
│    ├── config.yaml
│    ├── main.go
│    ├── go.work
│    ├── go.mod
│
├── monibuca
├── m7s-jt1078

go.work文件内容

go 1.23.2

use (
    .
    ../monibuca
    ../../m7s-jt1078
)
```

<h3>参数说明</h3>

``` yaml
global:

jt1078:
  enable: true
  audioports: [10000, 10010] # 音频端口 用于下发数据[min,max]
  simulations:
    # jt1078文件 默认循环发送
      - name: ../data/data.txt
        addr: 127.0.0.1:1078 # 模拟实时
      - name: ../data/data.txt
        addr: 127.0.0.1:1079 # 模拟回放

  realtime: # 实时视频
    addr: '0.0.0.0:1078'
    onjoinurl: "http://127.0.0.1:10011/api/v1/real-time-join"
    onleaveurl: "http://127.0.0.1:10011/api/v1/real-time-leave"
    prefix: "live/jt1078" # 默认自定义前缀-手机号-通道 如：live/jt1078-295696659617-1

  playback: # 回放视频
    addr: '0.0.0.0:1079'
    onjoinurl: "http://127.0.0.1:10011/api/v1/play-back-join"
    onleaveurl: "http://127.0.0.1:10011/api/v1/play-back-leave"
    prefix: "live/jt1079" # 默认自定义前缀-手机号-通道 如：live/jt1079-295696659617-1

mp4:
  enable: true

```

| 参数           | 描述                                                              | 例子                                  |
|---------------------|------------------------------------------------------------|------------------------------------------|
| jt1078          	| jt1078配置                                               |                                          |
|  &nbsp; enable        | 是否启用                                                  | `true`                                   |
|  &nbsp; audioports    | 使用的音频端口列表 [min,max]                               | `[10000, 10010]`                         |
|  &nbsp;simulations   	| 自定义模拟器测试1078流                                     |                                          |
|   &nbsp;&nbsp;name    | 读取的文件                                                | `../data/data.txt`                       |
|   &nbsp;&nbsp;addr    | 连接的地址 `IP:Port`.                                     | `127.0.0.1:1078`                         |
| &nbsp;realtime        | 实时视频                                                  |                                          |
|   &nbsp;&nbsp;addr          | 监听地址                                            | `0.0.0.0:1078`                           |
|   &nbsp;&nbsp;onjoinurl     | 有新流的时候 http回调                                 | `http://127.0.0.1:10011/api/v1/real-time-join` |
|   &nbsp;&nbsp;onleaveurl    | 新流结束的时候 http回调                               | `http://127.0.0.1:10011/api/v1/real-time-leave` |
|   &nbsp;&nbsp;prefix        | 流名称前缀 实际流名称为前缀-手机号-通道号                | `live/jt1078-{phone_number}-{channel}`   |
| &nbsp;playback              | 回放视频                                            |                                          |
|   &nbsp;&nbsp;addr          | 监听地址                                            | `0.0.0.0:1079`                           |
|   &nbsp;&nbsp;onjoinurl     | 有新流的时候 http回调                                 | `http://127.0.0.1:10011/api/v1/real-time-join` |
|   &nbsp;&nbsp;onleaveurl    | 新流结束的时候 http回调                               | `http://127.0.0.1:10011/api/v1/real-time-leave` |
|   &nbsp;&nbsp;prefix        | 流名称前缀 实际流名称为前缀-手机号-通道号                 | `live/jt1078-{phone_number}-{channel}`   |

回调统一使用http post 参数如下
``` json
{
    "audioPort": 10005,
    "sim": "295696659617",
    "channel": 1,
    "streamPath": "live/jt1078-295696659617-1"
}
```

| 参数           | 描述               | 例子          |
|----------------|----------------------------|-----------|
| audioPort      | 使用的音频端口 不使用则为0 不够了则为-1   | 10005            |
| sim            | 手机号                      | 295696659617   |
| channel        | 通道号                      | 1                 |
| streamPath     | m7s流名称                   | live/jt1078-295696659617-1   |

<h3>m7s-jt1078插件开发说明</h3>

1. 加入

``` go
var _ = m7s.InstallPlugin[JT1078Plugin]()

JT1078Plugin struct {
    m7s.Plugin
}
```

2. 自定义配置文件参数

- [m7s配置文件官方地址](https://monibuca.com/docs/guide/config.html)
- config.yaml中都使用小写
``` yaml
global:

jt1078:
  audioports: [10000, 10010] # 音频端口 用于下发数据[min,max]
```

``` go
JT1078Plugin struct {
    AudioPorts  [2]int   `default:"[10000,10010]" desc:"音频端口 用于下发数据"`
}
```

3. 把数据传入到m7s中

```
m7s.Plugin自带Publish方法 使用ctx控制退出
publisher.WriteVideo 写入视频数据
publisher.WriteAudio 写入音频数据
```

4. 常见问题
4.1 时间戳问题
![m7s时间戳](./testdata/m7s.png)

- 如本插件获取设备时间戳(毫秒)后 统一*90
- [关于帧率、比特率、DTS、PTS、分辨率](https://maxwellqi.github.io/ios-edcoder-fps-dts-etc/)

``` go
pkg.WithPTSFunc(func(_ *jt1078.Packet) time.Duration {
    return time.Duration(time.Now().UnixMilli()) * 90 // 实时视频使用本机时间戳 毫秒
}),

pkg.WithPTSFunc(func(pack *jt1078.Packet) time.Duration {
    return time.Duration(pack.Timestamp) * 90 // 录像回放使用设备的
}),
```