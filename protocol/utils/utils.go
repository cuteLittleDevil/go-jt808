package utils

import (
	"bytes"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"strings"
)

func Bcd2Dec(data []byte) string {
	out := bcdConvert(data)
	if noZero := bytes.IndexFunc(out, func(r rune) bool {
		return r != '0'
	}); noZero != -1 {
		return string(out[noZero:])
	}
	return string(out)
}

func Time2BCD(time string) []byte {
	if strings.Contains(time, ":") {
		time = strings.ReplaceAll(time, "-", "")
		time = strings.ReplaceAll(time, ":", "")
		time = strings.ReplaceAll(time, " ", "")
		if len(time) == 14 {
			time = time[2:]
		}
	}
	if len(time)%2 != 0 {
		time = "0" + time
	}
	bcd := make([]byte, len(time)/2)
	for i := 0; i < len(time); i += 2 {
		// 高4位是第一个字符，低4位是第二个字符
		bcd[i/2] = ((time[i] - '0') << 4) | (time[i+1] - '0')
	}
	return bcd
}

func CreateVerifyCode(data []byte) byte {
	var code byte
	for _, v := range data {
		code ^= v
	}
	return code
}

func String2FillingBytes(text string, size int) []byte {
	data := []byte(text)
	if len(data) < size {
		data = append(data, make([]byte, size-len(data))...)
	} else if len(data) > size {
		data = data[:size]
	}
	return data
}

func BCD2Time(bcd []byte) string {
	result := make([]byte, len(bcd)*2)
	for i, v := range bcd {
		result[2*i] = (v >> 4) + '0'
		result[2*i+1] = (v & 0x0F) + '0'
	}
	if len(bcd) == 6 {
		return fmt.Sprintf("20%s-%s-%s %s:%s:%s",
			result[0:2], result[2:4], result[4:6],
			result[6:8], result[8:10], result[10:12])
	}
	return string(result)
}

func GBK2UTF8(data []byte) []byte {
	utf8Data, _ := io.ReadAll(transform.NewReader(bytes.NewBuffer(data), simplifiedchinese.GBK.NewDecoder()))
	return utf8Data
}

func UTF82GBK(data []byte) []byte {
	gbkData, _ := io.ReadAll(transform.NewReader(bytes.NewBuffer(data), simplifiedchinese.GBK.NewEncoder()))
	return gbkData
}

func bcdConvert(data []byte) []byte {
	out := make([]byte, 2*len(data))
	index := 0
	for i := 0; i < len(data); i++ {
		out[index] = nibbleToHexChar(data[i] >> 4)
		index++

		out[index] = nibbleToHexChar(data[i] & 0x0f)
		index++
	}
	return out
}

func nibbleToHexChar(nibble byte) byte {
	if nibble <= 9 {
		return nibble + '0'
	}
	switch {
	case nibble == 0x0a:
		return 'a'
	case nibble == 0x0b:
		return 'b'
	case nibble == 0x0c:
		return 'c'
	case nibble == 0x0d:
		return 'd'
	case nibble == 0x0e:
		return 'e'
	default:
	}
	// 0x0f bcd编码16进制 只有0-9和a-f
	return 'f'
}
