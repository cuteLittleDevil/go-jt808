package jt1078

import "bytes"

func bcd2dec(data []byte) string {
	out := bcdConvert(data)
	if noZero := bytes.IndexFunc(out, func(r rune) bool {
		return r != '0'
	}); noZero != -1 {
		return string(out[noZero:])
	}
	return string(out)
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
