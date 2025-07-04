package command

import (
	"bytes"
	"encoding/xml"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"strings"
)

type XMLTypes interface {
	*ConfirmType | *Keepalive | *DeviceInfo | *DeviceInfoResponse |
		*DeviceStatus | *DeviceStatusResponse | *Catalog | *CatalogResponse
}

func ToXML[T XMLTypes](v T) []byte {
	output, _ := xml.MarshalIndent(v, "", "  ")
	result := append([]byte("<?xml version=\"1.0\" encoding=\"GB2312\"?>\n"), output...)
	return utf82gbk18030(result)
}

func ParseXML[T XMLTypes](data []byte, v T) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch strings.ToUpper(charset) {
		case "GBK":
			return transform.NewReader(input, simplifiedchinese.GBK.NewDecoder()), nil
		case "GB2312":
			return transform.NewReader(input, simplifiedchinese.GB18030.NewDecoder()), nil
		default:
			// 对于其他编码，直接使用原始输入（默认处理UTF-8）
			return input, nil
		}
	}
	return decoder.Decode(v)
}

func utf82gbk18030(data []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GB18030.NewEncoder())
	b, _ := io.ReadAll(reader)
	return b
}
