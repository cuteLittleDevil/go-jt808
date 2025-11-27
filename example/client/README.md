# 并发测试

mac系统的参数 临时修改
``` shell
# 增大系统默认的最大连接数限制
sudo sysctl -w kern.maxfiles=8880000
# 增大单个进程默认最大连接数限制
sudo sysctl -w kern.maxfilesperproc=8990000
# 设置当前shell能打开的最大文件数
ulimit -n 1100000
# 设置当前shell可以创建的最大用户线程数
ulimit -u 26600
# 调整可用端口数量
sysctl -w net.inet.ip.portrange.first=5000
```

linux系统的 永久修改
``` shell
vi /etc/sysctl.conf
# 服务器系统级参数 最大打开文件数
fs.file-max=1100000
# 服务器进程级参数 最大打开文件数
fs.nr_open=1100000
# 增加进程可用端口
net.ipv4.ip_local_port_range = 5000 65000

vi /etc/security/limits.conf
# 限制用户进程最大打开文件数量限制 soft(软限制) hard(硬限制)
*  soft  nofile  1000000
*  hard  nofile  1010000

# 生效
sysctl -p
# 查看
sysctl -a
```

---

- 获取发送的报文信息的url http://127.0.0.1:8080/api/v1/all
```
{
    "allSum": 54,
    "records": [
        {
            "sim": "3",
            "sum": 18,
            "commands": {
                "0002:终端-心跳": 2,
                "0200:终端-位置上报": 16
            },
            "createTime": "2025-11-26T20:46:31.886216+08:00",
            "updateTime": "2025-11-26T20:47:51.885295+08:00"
        },
        {
            "sim": "2",
            "sum": 18,
            "commands": {
                "0002:终端-心跳": 2,
                "0200:终端-位置上报": 16
            },
            "createTime": "2025-11-26T20:46:31.885081+08:00",
            "updateTime": "2025-11-26T20:47:51.884639+08:00"
        },
        {
            "sim": "1",
            "sum": 18,
            "commands": {
                "0002:终端-心跳": 2,
                "0200:终端-位置上报": 16
            },
            "createTime": "2025-11-26T20:46:31.88451+08:00",
            "updateTime": "2025-11-26T20:47:51.884646+08:00"
        }
    ]
}
```

- 配置文件参考
``` yaml
server:
  addr: "127.0.0.1:808" # 服务端地址

client:
  ip: "127.0.0.1" # 客户端ip
  intervalMicrosecond: 1000 # 多久生成一个新客户端 默认1毫秒一个
  sum: 1000 # 测试的最大客户端数量
  sim: 1 # 初始的sim卡号 后面的就一直+1
  version: 2013 # 客户端版本 2013 2019
  commands: # 按周期循环的指令
    - name: 0x0200 #位置信息
      enable: true # 是否启用
      interval: 5 # 5秒钟发送一次
      sum: 1000 # 至多发送多少报文 0-不限制 到了sum后就不发了
    - name: 0x0002 #心跳
      enable: true
      interval: 30
      sum: 0
```