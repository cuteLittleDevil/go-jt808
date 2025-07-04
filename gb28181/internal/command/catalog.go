package command

import (
	"encoding/xml"
	"fmt"
	"strconv"
)

// Catalog 查询通道情况
/*
<?xml version="1.0" encoding="GB2312"?>
<Query>
    <CmdType>Catalog</CmdType>
    <SN>2</SN>
    <DeviceID>34020000001320000105</DeviceID>
</Query>
.*/
type Catalog struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  string   `xml:"CmdType"`  // 命令类型 (必选 固定DeviceInfo)
	SN       int64    `xml:"SN"`       // 命令序列号 (必选)
	DeviceID string   `xml:"DeviceID"` // 目标设备/区域/系统的编码 (必选)
}

type (
	// CatalogResponse 目录查询回复
	/*
		<?xml version="1.0" encoding="GB2312"?>
		<Response>
		    <CmdType>Catalog</CmdType>
		    <SN>2</SN>
		    <DeviceID>34020000001320000310</DeviceID>
		    <SumNum>4</SumNum>
		    <DeviceList Num="4">
		        <Item>
		            <DeviceID>34020000001320000311</DeviceID>
		            <Name>???-1</Name>
		            <Manufacturer>go-jt808</Manufacturer>
		            <Model>simulation</Model>
		            <Owner>Owner</Owner>
		            <CivilCode>330100</CivilCode>
		            <Address>test</Address>
		            <Parental>0</Parental>
		            <ParentID>65010200002160000001</ParentID>
		            <SafetyWay>0</SafetyWay>
		            <RegisterWay>1</RegisterWay>
		            <Secrecy>0</Secrecy>
		            <Status>ON</Status>
		        </Item>
		        <Item>
		            <DeviceID>34020000001320000312</DeviceID>
		            <Name>???-2</Name>
		            <Manufacturer>go-jt808</Manufacturer>
		            <Model>simulation</Model>
		            <Owner>Owner</Owner>
		            <CivilCode>330100</CivilCode>
		            <Address>test</Address>
		            <Parental>0</Parental>
		            <ParentID>65010200002160000001</ParentID>
		            <SafetyWay>0</SafetyWay>
		            <RegisterWay>1</RegisterWay>
		            <Secrecy>0</Secrecy>
		            <Status>ON</Status>
		        </Item>
		        <Item>
		            <DeviceID>34020000001320000313</DeviceID>
		            <Name>???-3</Name>
		            <Manufacturer>go-jt808</Manufacturer>
		            <Model>simulation</Model>
		            <Owner>Owner</Owner>
		            <CivilCode>330100</CivilCode>
		            <Address>test</Address>
		            <Parental>0</Parental>
		            <ParentID>65010200002160000001</ParentID>
		            <SafetyWay>0</SafetyWay>
		            <RegisterWay>1</RegisterWay>
		            <Secrecy>0</Secrecy>
		            <Status>ON</Status>
		        </Item>
		        <Item>
		            <DeviceID>34020000001320000314</DeviceID>
		            <Name>???-4</Name>
		            <Manufacturer>go-jt808</Manufacturer>
		            <Model>simulation</Model>
		            <Owner>Owner</Owner>
		            <CivilCode>330100</CivilCode>
		            <Address>test</Address>
		            <Parental>0</Parental>
		            <ParentID>65010200002160000001</ParentID>
		            <SafetyWay>0</SafetyWay>
		            <RegisterWay>1</RegisterWay>
		            <Secrecy>0</Secrecy>
		            <Status>ON</Status>
		        </Item>
		    </DeviceList>
		</Response>
	.*/
	CatalogResponse struct {
		XMLName    xml.Name `xml:"Response"`
		CmdType    string   `xml:"CmdType"`
		SN         int64    `xml:"SN"`
		DeviceID   string   `xml:"DeviceID"`
		SumNum     int      `xml:"SumNum"`
		DeviceList struct {
			Num   int           `xml:"Num,attr"`
			Items []CatalogItem `xml:"Item"`
		} `xml:"DeviceList"`
	}

	// CatalogItem 目录选项.
	CatalogItem struct {
		DeviceID     string `xml:"DeviceID"`        // 设备/区域/系统编码(必选) 就是通道ID
		Name         string `xml:"Name"`            // 设备/区域/系统名称(必选)
		Manufacturer string `xml:"Manufacturer"`    // 当为设备时,设备厂商(必选)
		Model        string `xml:"Model"`           // 当为设备时,设备型号(必选)
		Owner        string `xml:"Owner"`           // 当为设备时,设备归属(必选)
		CivilCode    string `xml:"CivilCode"`       // 行政区域 (必选)
		Block        string `xml:"Block,omitempty"` // 警区 (可选)
		Address      string `xml:"Address"`         // 当为设备时,安装地址(必选)
		Parental     int    `xml:"Parental"`        // 当为设备时,是否有子设备(必选) 0:无 1:有
		ParentID     string `xml:"ParentID"`        // 父设备/区域/系统ID(必选)
		// SafetyWay 信令安全模式(可选)缺省为0; 0:不采用;2:S/MIME签名方式;3:S/ MIME 加密签名同时采用方式;4:数字摘要方式.
		SafetyWay int `xml:"SafetyWay"`
		// RegisterWay 注册方式 (必选)缺省为1; 1:符合IETF RFC 3261标准的认证注册模式 2:基于口令的双向认证注册模式; 3:基于数字证书的双向认证注册模式.
		RegisterWay int     `xml:"RegisterWay"`
		CertNum     string  `xml:"CertNum,omitempty"`     // 证书序列号(有证书的设备必选)
		Certifiable int     `xml:"Certifiable,omitempty"` // 证书有效标识(有证书的设备必选)缺省为0;证书有效标识: 0:无效 1:有效
		ErrCode     int     `xml:"ErrCode,omitempty"`     // 无效原因码(有证书且证书无效的设备必选)
		EndTime     string  `xml:"EndTime,omitempty"`     // 证书终止有效期(有证书的设备必选)
		Secrecy     int     `xml:"Secrecy"`               // 保密属性(必选)缺省为0;0:不涉密,1:涉密
		IPAddress   string  `xml:"IPAddress,omitempty"`   // 设备 / 区域 / 系统 IP 地址 (可选)
		Port        int     `xml:"Port,omitempty"`        // 设备/区域/系统端口(可选)
		Password    string  `xml:"Password,omitempty"`    // 设备口令(可选)
		Status      string  `xml:"Status"`                // 设备状态 (必选) 正常-OK 离线-OFF
		Longitude   float32 `xml:"Longitude,omitempty"`   // 经度 (可选)
		Latitude    float32 `xml:"Latitude,omitempty"`    // 纬度 (可选)
		// PTZType 摄像机类型扩展,标识摄像机类型:1-球机;2-半球;3-固定枪机;4-遥控枪 机。当目录项为摄像机时可选。
		PTZType int `xml:"PTZType,omitempty"`
		// PositionType 摄像机位置类型扩展。1-省际检查站、2-党政机关、3-车站码头、4-中心广场 、5-体育场馆 、6-商业中心 、7-宗教场所 、8-校园周边 、9-治安复杂区域 、10-交通干线。当目录项为摄像机时可选。
		PositionType int `xml:"PositionType,omitempty"`
		// RoomType 摄像机安装位置室外、室内属性。1-室外、2-室内。当目录项为摄像机时可选 ,缺省为1。
		RoomType int `xml:"RoomType,omitempty"`
		// UseType 摄像机用途属性。1-治安、2-交通、3-重点。当目录项为摄像机时可选。
		UseType int `xml:"UseType,omitempty"`
		// SupplyLightType 摄像机补光属性。1-无补光、2-红外补光、3-白光补光。当目录项为摄像机时可选 ,缺省为1。
		SupplyLightType int `xml:"SupplyLightType,omitempty"`
		// DirectionType 摄像机监视方位属性。1-东、2-西、3-南、4-北、5-东南、6-东北、7-西南、8-西北。当目录项为摄像机时且为固定摄像机或设置看守位摄像机时可选。
		DirectionType int `xml:"DirectionType,omitempty"`
		// Resolution 摄像机支持的分辨率,可有多个分辨率值,各个取值间以 / 分隔。分辨率 取值参见附录F中SDPf字段规定。当目录项为摄像机时可选。
		Resolution string `xml:"Resolution,omitempty"`
	}
)

func NewCatalogResponse(catalog Catalog, channelNum int) *CatalogResponse {
	items := make([]CatalogItem, 0, 4)
	for i := 0; i < channelNum; i++ {
		// 通道id 是把设备id的最后一位换了
		channelID := catalog.DeviceID[:len(catalog.DeviceID)-1] + strconv.Itoa(i+1)
		items = append(items, CatalogItem{
			DeviceID:     channelID,
			Name:         fmt.Sprintf("通道-%d", i+1),
			Manufacturer: "go-jt808",
			Model:        "simulation",
			Owner:        "Owner",
			CivilCode:    "330100", // 杭州西湖区
			Block:        "",
			Address:      "192.168.1.1",
			Parental:     0, // 没有子设备
			ParentID:     "65010200002160000001",
			RegisterWay:  1,
			Secrecy:      0,
			Status:       "ON",
		})
	}
	return &CatalogResponse{
		CmdType:  catalog.CmdType,
		SN:       catalog.SN,
		DeviceID: catalog.DeviceID,
		SumNum:   len(items),
		DeviceList: struct {
			Num   int           `xml:"Num,attr"`
			Items []CatalogItem `xml:"Item"`
		}{
			Num:   len(items),
			Items: items,
		},
	}
}
