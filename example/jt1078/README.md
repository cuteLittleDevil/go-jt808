# JT1078流媒体

<h2 id="rtvs-dev"> RTVS终端模拟器 </h2>

```
rtvsdev（1078终端模拟器docker版本）
命令行运行
docker run --restart always -p 5288:80 -d vanjoge/rtvsdevice
然后访问你的//IP:5288即可

```

<h2 id="rtvs"> RTVS </h2>

- [RTVS官方地址](https://gitee.com/vanjoge/RTVS)
- [部署文档参考](https://blog.csdn.net/vanjoge/article/details/108319078)
- [代码参考](./rtvs/main.go)

address是设备连接的地址 webAddress是页面的
```  go
./rtvs -address 0.0.0.0:8082 -webAddress 0.0.0.0:17001
```

1. 测试部署网页 http://49.234.235.7:17001
2. 让终端(模拟器)默认连接到了49.234.235.7:8082地址
3. 根据测试部署网页进行测试 如点击9101观看在线视频

测试模拟器的手机号为 013777883241

![9101实时视频测试](./data/rtvs9101.png)

需要对讲的话 则在本地打开tsrtvs.html测试

<h2 id="lal"> LAL流媒体服务 </h2>

1. 使用模拟器默认的数据 持续推送到LAL服务
2. 在线播放地址 http://49.234.235.7:8080/live/1001_1.flv

ip是外网的ip用于下发9101的1078的ip 可以用phone新建一个模拟终端 使用dataPath的数据推送1078流
```  go
./lal2 -ip=49.234.235.7
./lal2 -ip=49.234.235.7  -phone=1 -dataPath=./data.txt
```

```
运行后 让设备连接到808端口 默认3秒发送9101
打印flv的播放地址（其他播放地址参考lal文档）
默认的播放地址格式 http://49.234.235.7:8080/live/手机号_通道号.flv
如 http://49.234.235.7:8080/live/156987000796_1.flv
```

- [LAL官方文档](https://pengrl.com/lal/#/streamurllist)
- [代码参考](./lal/main.go)

<h2 id="sky-java"> JT1078 sky-java </h2>

1. 启动服务
2. 使用RTVS终端模拟器连接到服务
3. 调用sky-java的JT1078 HTTP接口发送请求(默认10秒内需要去拉流)
- [sky-java官方地址](https://gitee.com/hui_hui_zhou/open-source-repository)
- [sky-java HTTP文档](http://222.244.144.181:9991/doc.html)
- [代码参考](./sky/java/main.go)

<h2 id="m7s"> m7s-jt1078 </h2>

- [插件详情](https://github.com/cuteLittleDevil/m7s-jt1078)
- [代码参考](./m7s/main.go)

<h2 id="zlm"> ZLMediaKit </h2>
- [代码参考](./zlm/main.go)

1. 使用ZLMediaKit试用版
- https://github.com/ziyuexiachu/ci/actions/runs/13678145491/artifacts/2696568677

2. 启动ZLMediaKit 复制secret
```
unzip LinuxTry_feature_1078_2025-03-05.zip -d /home/zlm
cd /home/zlm/linux/Release
# 可以把./example/jt1078/zlm/config.ini放到 /home/zlm/linux/Release目录 使用这个配置文件
./MediaServer
# 案例的secret是5xGbdUpfXnsiW3uZq2CApzSyxSFrIWpc
cat /home/zlm/linux/Release/config.ini | grep secret

```

3. 启动go zlm的示例
```
cd ./example/jt1078/zlm
GOOS=linux GOARCH=amd64 go build -o go-zlm
# config.yaml中的secret换成步骤2生成的secret
# 这里把go-zlm和config.yaml放到了 /home/zlm/go/ 目录
./go-zlm

```

4. 使用模拟器连接到808服务
- 测试案例的808服务默认端口是8083
如下所示 终端sim卡号1004的加入了
```
终端加入 key=[1004] command=[7e01020004000000001004002631303034307e] err=[nil]

```

5. 调用http接口发送9101请求
- ip换成部署的机器 案例云服务器ip 124.221.30.46 serverIPLen换ip的长度
``` curl
curl --location --request POST 'http://124.221.30.46:17002/api/v1/9101' \
--header 'Content-Type: application/json' \
--data-raw '{
    "key": "1004",
    "data": {
        "serverIPLen": 13,
        "serverIPAddr": "124.221.30.46",
        "tcpPort": 1078,
        "udpPort": 0,
        "channelNo": 1,
        "dataType": 0,
        "streamType": 0
    }
}'

```

- 返回结果如下
```
{
  "code": 200,
  "msg": "成功",
  "data": {
    "streamID": "1004-1-1078",
    "mp4": "http://124.221.30.46:80/rtp/1004-1-1078.live.mp4"
  }
}
```

6. 扩展
```
在config.ini的配置文件中添加
hook.enable=1
hook.on_stream_not_found=http://127.0.0.1:17002/api/v1/on_stream_not_found

用户访问http://124.221.30.46:80/rtp/1004-1-1078.live.mp4
当流不存在的时候 主动下发9101让这个流存在
```