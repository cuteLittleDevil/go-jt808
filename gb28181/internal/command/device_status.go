package command

import "encoding/xml"

// DeviceStatus 查询通道情况
/*
<?xml version="1.0" encoding="GB2312"?>
<Query>
    <CmdType>DeviceStatus</CmdType>
    <SN>1</SN>
    <DeviceID>34020000001320000105</DeviceID>
</Query>
.*/
type DeviceStatus struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  string   `xml:"CmdType"`  // 命令类型 (必选 固定DeviceInfo)
	SN       int64    `xml:"SN"`       // 命令序列号 (必选 )
	DeviceID string   `xml:"DeviceID"` // 目标设备/区域/系统的编码 (必选 )
}

// DeviceStatusResponse 查询通道的返回值 GB/T28181—2016 67页
/*
<?xml version="1.0" encoding="GB2312"?>
<Response>
	<CmdType>DeviceStatus</CmdType>
	<SN>1676968400</SN>
	<DeviceID>34020000001320000041</DeviceID>
	<Result>OK</Result>
	<Online>ONLINE</Online>
	<Status>OK</Status>
	<Encode>ON</Encode>
	<Record>ON</Record>
</Response>
.*/
type DeviceStatusResponse struct {
	XMLName  xml.Name `xml:"Response"`
	CmdType  string   `xml:"CmdType"`
	SN       int64    `xml:"SN"`               // 命令序列号 (必选 )
	DeviceID string   `xml:"DeviceID"`         // 目标设备/区域/系统的编码 (必选 )
	Result   string   `xml:"Result"`           // 查询结果标志 (必选 )
	Online   string   `xml:"Online"`           // 是否在线 (必选 ) 在线-ONLINE 离线-OFFLINE
	Status   string   `xml:"Status"`           // 是否正常工作 (必选 )
	Reason   string   `xml:"Reason,omitempty"` // 不正常工作原因(可选)
	Encode   string   `xml:"Encode,omitempty"` // 是否编码 (可选 )
	Record   string   `xml:"Record,omitempty"` // 是否录像 (可选 )
}

func NewDeviceStatusResponse(status DeviceStatus) *DeviceStatusResponse {
	return &DeviceStatusResponse{
		CmdType:  status.CmdType,
		SN:       status.SN,
		DeviceID: status.DeviceID,
		Result:   "OK",
		Online:   "ONLINE",
		Status:   "OK",
		Reason:   "",
		Encode:   "ON",
		Record:   "ON",
	}
}
