package command

import (
	"os"
	"strings"
	"testing"
)

func TestParseXML(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		instance any
	}{
		{
			name:     "心跳保活 KEEPALIVE",
			filePath: "../../testdata/keepalive.xml",
			instance: &Keepalive{},
		},
		{
			name:     "查询设备信息 DEVICE_INFO",
			filePath: "../../testdata/device_info.xml",
			instance: &DeviceInfo{},
		},
		{
			name:     "查询设备信息-回复 DEVICE_INFO_RESPONSE",
			filePath: "../../testdata/device_info_response.xml",
			instance: &DeviceInfoResponse{},
		},
		{
			name:     "查询设备状态 DEVICE_STATUS",
			filePath: "../../testdata/device_status.xml",
			instance: &DeviceStatus{},
		},
		{
			name:     "查询设备状态-回复 DEVICE_STATUS_RESPONSE",
			filePath: "../../testdata/device_status_response.xml",
			instance: &DeviceStatusResponse{},
		},
		{
			name:     "查询目录 CATALOG",
			filePath: "../../testdata/catalog.xml",
			instance: &Catalog{},
		},
		{
			name:     "查询目录-回复 CATALOG_RESPONSE",
			filePath: "../../testdata/catalog_response.xml",
			instance: &CatalogResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.filePath)
			if err != nil {
				t.Errorf("path[%s] err=[%v]", tt.filePath, err)
				return
			}
			switch v := tt.instance.(type) {
			case *Keepalive:
				verifyXML[*Keepalive](t, data, v)
			case *DeviceInfo:
				verifyXML[*DeviceInfo](t, data, v)
			case *DeviceInfoResponse:
				verifyXML[*DeviceInfoResponse](t, data, v)
			case *DeviceStatus:
				verifyXML[*DeviceStatus](t, data, v)
			case *DeviceStatusResponse:
				verifyXML[*DeviceStatusResponse](t, data, v)
			case *Catalog:
				verifyXML[*Catalog](t, data, v)
			case *CatalogResponse:
				verifyXML[*CatalogResponse](t, data, v)
			default:
				t.Errorf("不支持的类型: %T", v)
				return
			}
		})
	}
}

func verifyXML[T XMLTypes](t *testing.T, data []byte, instance T) {
	gbData := utf82gbk18030(data)
	if err := ParseXML(gbData, instance); err != nil {
		t.Errorf("type[%T] err=[%v]", instance, err)
		return
	}
	got := ToXML(instance)
	// 忽略格式 内容正确就ok
	expect := strings.ReplaceAll(string(gbData), " ", "")
	expect = strings.ReplaceAll(expect, "\n", "")
	actual := strings.ReplaceAll(string(got), " ", "")
	actual = strings.ReplaceAll(actual, "\n", "")
	if !strings.EqualFold(expect, actual) {
		t.Errorf("\nexpect:%s \n actual:%v", expect, actual)
		return
	}
}

func TestCommandConstructorsAndBoundary(t *testing.T) {
	ka := NewKeepalive("34020000001320000011", 7)
	if ka.CmdType != "Keepalive" || ka.Status != "OK" || ka.SN != 7 || ka.DeviceID != "34020000001320000011" {
		t.Fatalf("unexpected Keepalive: %+v", ka)
	}
	if xmlBytes := ToXML(ka); !strings.Contains(string(xmlBytes), "<CmdType>Keepalive</CmdType>") {
		t.Fatalf("unexpected Keepalive XML: %s", string(xmlBytes))
	}

	devInfo := NewDeviceInfoResponse("1001-sim", DeviceInfo{CmdType: "DeviceInfo", SN: 3, DeviceID: "34020000001320000011"})
	if devInfo.DeviceName != "1001-sim" || devInfo.Result != "OK" || devInfo.Channel != 4 {
		t.Fatalf("unexpected DeviceInfoResponse: %+v", devInfo)
	}

	devStatus := NewDeviceStatusResponse(DeviceStatus{CmdType: "DeviceStatus", SN: 4, DeviceID: "34020000001320000011"})
	if devStatus.Online != "ONLINE" || devStatus.Status != "OK" || devStatus.Result != "OK" {
		t.Fatalf("unexpected DeviceStatusResponse: %+v", devStatus)
	}

	// 空 DeviceID 与负数通道数不 panic
	emptyCatalog := NewCatalogResponse(Catalog{CmdType: "Catalog", SN: 5, DeviceID: ""}, 2)
	if emptyCatalog.SumNum != 2 || len(emptyCatalog.DeviceList.Items) != 2 || emptyCatalog.DeviceList.Items[0].DeviceID != "1" {
		t.Fatalf("unexpected empty DeviceID CatalogResponse: %+v", emptyCatalog)
	}
	negCatalog := NewCatalogResponse(Catalog{CmdType: "Catalog", SN: 6, DeviceID: "34020000001320000310"}, -1)
	if negCatalog.SumNum != 0 || len(negCatalog.DeviceList.Items) != 0 {
		t.Fatalf("unexpected negative channelNum CatalogResponse: %+v", negCatalog)
	}
}

func TestParseXMLCharsetsAndError(t *testing.T) {
	gbkXML := utf82gbk18030([]byte(`<?xml version="1.0" encoding="GBK"?><Query><CmdType>Catalog</CmdType><SN>1</SN><DeviceID>34020000001320000310</DeviceID></Query>`))
	var ct ConfirmType
	if err := ParseXML(gbkXML, &ct); err != nil || ct.CmdType != "Catalog" {
		t.Fatalf("GBK ParseXML failed: ct=%+v err=%v", ct, err)
	}

	utf8XML := []byte(`<?xml version="1.0" encoding="UTF-8"?><Query><CmdType>DeviceInfo</CmdType></Query>`)
	if err := ParseXML(utf8XML, &ct); err != nil || ct.CmdType != "DeviceInfo" {
		t.Fatalf("UTF-8 ParseXML failed: ct=%+v err=%v", ct, err)
	}

	if err := ParseXML([]byte("<broken"), &ct); err == nil {
		t.Fatal("expected error for broken XML")
	}
}
