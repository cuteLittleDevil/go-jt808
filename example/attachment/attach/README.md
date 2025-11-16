
-  回调示例

``` json
{
    "alarmID":"2024-11-22_10_00_00_",
    "phone": "1001",
    "remark": "初始化状态-收到0x1210",
    "status": 1
}

```

``` json
{
    "alarmID":"2024-11-22_10_00_00_",
    "phone": "1001",
    "remark": "开始状态-收到0x1211",
    "status": 2
}

```

``` json
{
    "alarmID":"2024-11-22_10_00_00_",
    "phone": "1001",
    "remark": "文件码流数据收集中",
    "status": 3
}

```

``` json
{
    "alarmID":"2024-11-22_10_00_00_",
    "phone": "1001",
    "remark": "补传状态-等待最终完成",
    "status": 4
}

```

``` json
{
    "alarmID":"2024-11-22_10_00_00_",
    "phone": "1001",
    "remark": "文件码流数据收集完成",
    "status": 5
}

```

``` json
{
    "alarmID":"2024-11-22_10_00_00_",
    "phone": "1001",
    "remark": "完成状态-收到0x1212 并且没有需要补传的",
    "status": 6
}

```

``` json
{
    "alarmID":"2024-11-22_10_00_00_",
    "files": [
        {
            "name": "2025-11-15_10_00_00_client",
            "path": "/home/attachment/1001/2024-11-22_10_00_00_client",
            "size": 8570946
        },
        {
            "name": "2025-11-15_10_00_00_client.exe",
            "path": "/home/attachment/1001/2024-11-22_10_00_00_client.exe",
            "size": 4363776
        },
        {
            "name": "2025-11-15_10_00_00_main.go",
            "path": "/home/attachment/1001/2024-11-22_10_00_00_main.go",
            "size": 5466
        }
    ],
    "phone": "1001",
    "remark": "成功退出状态-所有的文件都接收完成",
    "status": 7
}
```

``` json
{
    "alarmID":"2024-11-22_10_00_00_",
    "phone": "1001",
    "remark": "异常退出状态-文件没有全部接收完成",
    "status": 8
}

```