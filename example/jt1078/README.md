# JT1078流媒体

<h2 id="rtvs-dev"> RTVS终端模拟器 </h2>

```
rtvsdev（1078终端模拟器docker版本）
命令行运行
docker run --restart always -p 5288:80 -d vanjoge/rtvsdevice
然后访问你的http://IP:5288即可

```

<h2 id="lal"> LAL流媒体服务 </h2>

使用模拟器默认的数据 持续推送到LAL服务
- [在线播放地址 FLV](http://49.234.235.7:8080/live/295696659617_1.flv)
- [LAL官方文档](https://pengrl.com/lal/#/streamurllist)
- [代码参考](./lal/main.go)


<h2 id="sky-java"> JT1078 sky-java </h2>

1. 启动服务
2. 使用RTVS终端模拟器连接到服务
3. 调用sky-java的JT1078 HTTP接口发送请求(默认10秒内需要去拉流)
- [sky-java官方地址](https://gitee.com/hui_hui_zhou/open-source-repository)
- [sky-java HTTP文档](http://222.244.144.181:9991/doc.html)
- [代码参考](./sky-java/main.go)