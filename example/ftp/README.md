# FTP例子

## 启动FTP服务

1. https://github.com/fclairamb/ftpserver/releases
2. 下载ftpserver 使用如下配置启动 ./ftpserver

``` json
{
  "version": 1,
  "accesses": [
    {
      "user": "test",
      "pass": "test",
      "fs": "os",
      "params": {
        "basePath": "/tmp/ftp"
      }
    }
  ],
  "passive_transfer_port_range": {
    "start": 2122,
    "end": 2130
  }
}

```
## 启动模拟ftp服务

``` go
go build && ./ftp
```

## 模拟9205和9206请求

1. 使用apifox模拟请求

https://vsh9jdgg5d.apifox.cn/257779934e0

2. 使用curl命令模拟请求
- 9205请求
```
curl --location --request POST 'http://127.0.0.1:8080/api/v1/ftp/9205' \
--data-raw '{
    "key": "1001",
    "command": 37381,
    "data": {
        "channelNo": 1,
        "startTime": "2024-11-02 00:00:00",
        "endTime": "2024-11-02 23:59:59",
        "alarmFlag": 0,
        "mediaType": 0,
        "streamType": 0,
        "storageType": 1
    }
}'

```

- 9206请求
```
curl --location --request POST 'http://127.0.0.1:8080/api/v1/ftp/9206' \
--data-raw '{
    "key": "1001",
    "command": 37382,
    "data": {
        "ftpAddrLen": 9,
        "ftpAddr": "127.0.0.1",
        "port": 2121,
        "usernameLen": 4,
        "username": "test",
        "passwordLen": 4,
        "password": "test",
        "fileUploadPathLen": 5,
        "fileUploadPath": "/1001",
        "channelNo": 1,
        "startTime": "2024-11-02 00:00:00",
        "endTime": "2024-11-02 00:01:02",
        "alarmFlag": 0,
        "mediaType": 0,
        "streamType": 1,
        "memoryPosition": 1,
        "taskExecuteCondition": 1
    }
}'

```
## 结果

ftp保存到指定目录
``` shell
ls /tmp/ftp/1001
atop_cpu.png
```