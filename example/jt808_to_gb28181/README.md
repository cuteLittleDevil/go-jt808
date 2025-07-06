# jt808转gb28181

![流程](./testdata/jt808-to-gb2818.jpg)

- 流程情况

```
原: jt808服务启动在808端口 设备连接到121.40.19.233:808

现: 808端口给适配器服务 设备连接到适配器 适配器产生2个jt808连接
一个连接到原jt808服务 (adapter.leader 默认127.0.0.1:20001) 保证原jt808服务业务正常
另一个连接到jt808转gb28181服务 (followers[0].address 默认127.0.0.1:20002)
只读写限制(只下发jt1078相关的指令) 不影响原jt808服务
```

<h2 id="m7s"> 1. m7s </h2>

- [m7s官网](https://monibuca.com/)
- admin.zip 加QQ下载(751639168) 或 https://download.m7s.live/bin/admin.zip
- 参考的配置文件的外网ip是121.40.19.233 需要修改成自己的

```
# 使用./testdata/m7s/config.yaml配置文件
# 配置文件和admin.zip(不用解压)放在m7s可执行文件同目录
./m7s

访问默认网页 选择国标设备->设备管理->通道详情->播放
http://127.0.0.1:12079/
```
![m7s-play](./testdata/m7s/m7s-play.jpg)

<h2 id="gb28181"> 1. gb28181 </h2>

- 信令使用gb28181 https://github.com/gowvp/gb28181
- 流媒体使用zlm https://github.com/ZLMediaKit/ZLMediaKit
- 参考的配置文件的外网ip是121.40.19.233 需要修改成自己的

```
# 可以根据gb28181官方 打包docker使用
# 这里使用直接编译的gb28181和下载好的zlm

# zlm下载 https://github.com/ZLMediaKit/ZLMediaKit/issues/483
# 使用的配置文件 ./testdata/gb28181/config.ini 放到和MediaServer同级的目录
./MediaServer -c config.ini

# 使用Makefile文件上的构建linux应用 可执行文件为bin
# 前端使用github上的下载链接 https://github.com/gowvp/gb28181_web/releases
# 下载www.zip 解压放到和bin同级的目录
# 配置文件 ./testdata/gb28181/config.toml 放到./configs/config.toml
./bin

默认首页 左侧国标通道 -> 选择任意一个通道点击
http://121.40.19.233:15123
```
![zlm-play](./testdata/gb28181/zlm-play.jpg)
