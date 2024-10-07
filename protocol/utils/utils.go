package utils

import (
	"bytes"
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
	//if len(time)%2 != 0 {
	//	time = "0" + time
	//}
	bcd := make([]byte, len(time)/2)
	for i := 0; i < len(time); i += 2 {
		// 高4位是第一个字符，低4位是第二个字符
		bcd[i/2] = ((time[i] - '0') << 4) | (time[i+1] - '0')
	}
	return bcd
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
