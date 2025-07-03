package internal

import "encoding/xml"

// DeviceInfo 查询通道情况
/*
<?xml version="1.0" encoding="GB2312"?>
<Query>
	<CmdType>DeviceInfo</CmdType>
 	<SN>0</SN>
	<DeviceID>34020000001320000016</DeviceID>
</Query>
*/
type DeviceInfo struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  string   `xml:"CmdType"`  // 命令类型 (必选 固定DeviceInfo)
	SN       int64    `xml:"SN"`       // 命令序列号 (必选 )
	DeviceID string   `xml:"DeviceID"` // 目标设备/区域/系统的编码 (必选 )
}

// DeviceInfoResponse 查询通道的返回值 GB/T28181—2016 66页
/*
<?xml version="1.0" encoding="GB2312"?>
<Response>
	<CmdType>DeviceInfo</CmdType>
	<SN>3</SN>
	<DeviceID>34020000001110000005</DeviceID>
	<DeviceName>jt808-simulation</DeviceName>
	<Result>OK</Result>
	<DeviceType>132</DeviceType>
	<Manufacturer>go-jt808</Manufacturer>
	<Model>simulation</Model>
	<Firmware>v0.1.0</Firmware>
	<Channel>4</Channel>
</Response>
*/
type DeviceInfoResponse struct {
	XMLName      xml.Name `xml:"Response"`
	CmdType      string   `xml:"CmdType"`      // 命令类型 (必选 固定DeviceInfo)
	SN           int64    `xml:"SN"`           // 命令序列号 (必选 )
	DeviceID     string   `xml:"DeviceID"`     // 目标设备/区域/系统的编码 (必选 )
	DeviceName   string   `xml:"DeviceName"`   // 目标设备/区域/系统的名称(可选)
	Result       string   `xml:"Result"`       // 查询结果 (必选 )
	DeviceType   string   `xml:"DeviceType"`   // 设备类型
	Manufacturer string   `xml:"Manufacturer"` // 设备生产商 (可选 )
	Model        string   `xml:"Model"`        // 设备型号 (可选 )
	Firmware     string   `xml:"Firmware"`     // 设备固件版本 (可选 )
	Channel      int      `xml:"Channel"`      // 视频输入通道数(可选)
}

func NewDeviceInfoResponse(info DeviceInfo) *DeviceInfoResponse {
	return &DeviceInfoResponse{
		CmdType:      info.CmdType,
		SN:           info.SN,
		DeviceID:     info.DeviceID,
		DeviceName:   "jt808-simulation",
		Result:       "OK",
		DeviceType:   "132",
		Manufacturer: "go-jt808",
		Model:        "simulation",
		Firmware:     "v0.1.0",
		Channel:      4,
	}
}
