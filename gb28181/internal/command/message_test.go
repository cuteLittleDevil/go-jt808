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
