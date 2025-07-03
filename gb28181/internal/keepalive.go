package internal

import (
	"encoding/xml"
)

// Keepalive 保活 示例如下
// <?xml version="1.0" encoding="GB2312"?>
// <Notify>
// <CmdType>Keepalive</CmdType>
// <SN>2</SN>
// <DeviceID>34020000001320000011</DeviceID>
// <Status>OK</Status>
// </Notify>
type Keepalive struct {
	XMLName  xml.Name `xml:"Notify"`
	CmdType  string   `xml:"CmdType"`
	SN       uint32   `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	Status   string   `xml:"Status"`
}

func NewKeepalive(deviceID string, SN uint32) *Keepalive {
	return &Keepalive{
		SN:       SN,
		DeviceID: deviceID,
		CmdType:  "Keepalive",
		Status:   "OK",
	}
}
