# 对接完成的协议

## JT808 终端通讯协议消息对照表

| 序号  |    消息 ID    | 完成情况 |  测试情况  | 消息体名称                     |  2019 版本   | 2011 版本 |
| :---: | :-----------: | :------: | :--------: | :----------------------- | :----------: | :-------: |
|   1   |    0x0001     |    ✅    |     ✅     | 终端通用应答				|				|			|
|   2   |    0x8001     |    ✅    |     ✅     | 平台-通用应答				|				|           |
|   3   |    0x0002     |    ✅    |     ✅     | 终端心跳					|				|           |
|   5   |    0x0100     |    ✅    |     ✅     | 终端注册					|     修改		|  被修改	|
|   4   |    0x8003     |    ✅    |     ✅     | 补传分包请求                |               |  被新增    |
|   6   |    0x8100     |    ✅    |     ✅     | 平台-注册应答				|				|           |
|   8   |    0x0102     |    ✅    |     ✅     | 终端鉴权					|     修改		|			|
|   9   |    0x8103     |    ✅    |     ✅     | 设置终端参数                |  修改且增加  	|  被修改    |
|  10   |    0x8104     |    ✅    |     ✅     | 平台-查询终端参数			|				|           |
|  11   |    0x0104     |    ✅    |     ✅     | 查询终端参数应答			|				|           |
|  18   |    0x0200     |    ✅    |     ✅     | 位置信息汇报				| 增加附加信息 	|  被修改	|
|  49   |    0x0704     |    ✅    |     ✅     | 定位数据批量上传			|     修改		|  被新增	|
|  51   |    0x0800     |    ✅    |     ✅     | 多媒体事件信息上传           |              |  被修改   |
|  52   |    0x0801     |    ✅    |     ✅     | 多媒体数据上传               |     修改     |  被修改   |
|  53   |    0x8800     |    ✅    |     ✅     | 平台-多媒体数据上传应答       |              |  被修改   |
|  54   |    0x8801     |    ✅    |     ✅     | 平台-摄像头立即拍摄命令       |     修改     |           |
|  55   |    0x0805     |    ✅    |     ✅     | 摄像头立即拍摄命令应答        |     修改     |  被新增   |

## JT1078 扩展

| 序号  |    消息 ID     | 完成情况 	| 测试情况 | 消息体名称 |
| :---: | :-----------: | :------: | :--------: | :----------------------- |
|  13   |    0x1003     |    ✅    |    ✅    | 终端上传音视频属性       |
|  14   |    0x1005     |    ✅    |    ✅    | 终端上传乘客流量         |
|  15   |    0x1205     |    ✅    |    ✅    | 终端上传音视频资源列表   |
|  16   |    0x1206     |    ✅    |    ✅    | 文件上传完成通知         |
|  17   |    0x9003     |    ✅    |    ✅    | 平台-查询终端音视频属性       |
|  18   |    0x9101     |    ✅    |    ✅    | 平台-实时音视频传输请求       |
|  19   |    0x9102     |    ✅    |    ✅    | 平台-音视频实时传输控制       |
|  20   |    0x9105     |    ✅    |    ✅    | 平台-实时音视频传输状态通知   |
|  21   |    0x9201     |    ✅    |    ✅    | 平台-下发远程录像回放请求 |
|  22   |    0x9202     |    ✅    |    ✅    | 平台-下发远程录像回放控制 |
|  23   |    0x9205     |    ✅    |    ✅    | 平台-查询资源列表             |
|  24   |    0x9206     |    ✅    |    ✅    | 平台-文件上传指令             |
|  25   |    0x9207     |    ✅    |    ✅    | 平台-文件上传控制             |

## 主动安全

| 序号  |    消息 ID    | 完成情况 | 测试情况 | 消息体名称                 |
| :---: | :-----------: | :------: | :------: | :------------------------- |
|   1   |    0x1210     |    ✅    |    ✅    | 报警附件信息消息           |
|   2   |    0x1211     |    ✅    |    ✅    | 文件信息上传               |
|   3   |    0x1212     |    ✅    |    ✅    | 文件上传完成消息           |
|   4   |    0x9208     |    ✅    |    ✅    | 报警附件上传指令           |
|   5   |    0x9212     |    ✅    |    ✅    | 文件上传完成消息应答       |