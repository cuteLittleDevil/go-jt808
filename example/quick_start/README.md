# 快速开始例子

包含最简单的0x0200位置示例
1. 模拟一个设备连接
2. 打印经纬度等信息

```
数据体对象:{
        终端-位置上报:[000004000000080006eeb6ad02633df7013800030063241001235959]
        [00000400] 报警标志:[1024]
        [00000800] 状态标志:[2048]
        [06eeb6ad] 纬度:[116307629]
        [02633df7] 经度:[40058359]
        [0138] 海拔高度:[312]
        [0003] 速度:[3]
        [0063] 方向:[99]
        [241001235959] 时间:[2024-10-01 23:59:59]
                [bit31]非法开门报警:[false]
                [bit30]侧翻预警:[false]
                [bit29]碰撞预警[false]
                [bit28]车辆非法位移:[false]
                [bit27]车辆非法点火:[false]
                [bit26]车辆被盗(通过车辆防盗器):[false]
                [bit25]车辆油量异常:[false]
                [bit24]车辆VSS故障:[false]
                [bit23]路线偏离报警:[false]
                [bit22]路段行驶时间不足/过长:[false]
                [bit21]进出路线:[false]
                [bit20]进出区域:[false]
                [bit19]超时停车:[false]
                [bit18]当天累计驾驶超时:[false]
                [bit17]:右转盲区预警[false]
                [bit16]:胎压预警[false]
                [bit15]:违规行驶预警[false]
                [bit14]疲劳驾驶预警:[false]
                [bit13]超速预警:[false]
                [bit12]道路运输证IC卡模块故障:[false]
                [bit11]摄像头故障:[false]
                [bit10]TTS模块故障]:[true]
                [bit9]终端LCD或显示器故障:[false]
                [bit8]终端主电源掉电:[false]
                [bit7]终端主电源欠压:[false]
                [bit6]GNSS天线短路:[false]
                [bit5]GNSS天线未接或被剪断:[false]
                [bit4]GNSS模块发生故障:[false]
                [bit3]危险预警:[false]
                [bit2]疲劳驾驶:[false]
                [bit1]超速报警:[false]
                [bit0]紧急报警,触动报警开关后触发:[false]
                [bit22]车辆状态 是否运行:[false] 2019版本增加的
                [bit21]使用Galileo卫星进行定位:[false]
                [bit20]未使用GLONASS卫星进行定位:[false]
                [bit19]未使用北斗卫星进行定位:[false]
                [bit18]未使用GPS卫星进行定位:[false]
                [bit17]自定义门 门5:[false]
                [bit16]驾驶席门 门4:[false]
                [bit15]后门情况 门3:[false]
                [bit14]中门情况 门2:[false]
                [bit13]前门情况 门1:[false]
                [bit12]车门情况 是否解锁:[false]
                [bit11]电路情况 是否断开:[true]
                [bit10]油路情况 是否正常:[false]
                [bit8-bit9]载客情况 0-空车 1-半载 2-保留 3-满载:[00]
                [bit7]车道偏移预警:[false] 2019版本增加的
                [bit6]紧急刹车系统的前撞预警:[false] 2019版本增加的
                [bit5]经纬度是否加密:[false]
                [bit4]运营状态 是否停运:[false]
                [bit3]东西经 0-东经 1-西经:[false]
                [bit2]南北纬 0-北纬 1-南纬:[false]
                [bit1]定位状态 是否开:[false]
                [bit0]ACC 是否开:[false]
}

```