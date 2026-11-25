package jt808

import (
	"bytes"

	"github.com/cuteLittleDevil/go-jt808/protocol"
)

func unescape(data []byte) ([]byte, error) {
	const (
		beforeEscape = 0x7d
		afterRecover = 0x7e
	)
	if !(len(data) > 2 && data[0] == afterRecover && data[len(data)-1] == afterRecover) {
		return nil, protocol.ErrUnqualifiedData
	}
	inner := data[1 : len(data)-1]
	// 首尾标识位之间不允许出现未转义的裸 0x7e
	if bytes.IndexByte(inner, afterRecover) != -1 {
		return nil, protocol.ErrUnqualifiedData
	}
	// 快速路径：没有需要反转义的字节，直接零拷贝返回切片
	if bytes.IndexByte(inner, beforeEscape) == -1 {
		return inner, nil
	}

	buf := make([]byte, 0, len(inner))
	index := 1
	for i := 1; i < len(data)-1; i++ {
		if data[i] == beforeEscape {
			i++
			switch data[i] {
			case 0x01:
				buf = append(buf, data[index:i-1]...)
				buf = append(buf, beforeEscape)
			case 0x02:
				buf = append(buf, data[index:i-1]...)
				buf = append(buf, afterRecover)
			default:
				// 兼容一下设备校验码不转义的情况
				if i == len(data)-1 {
					return append(buf, data[index:len(data)-1]...), nil
				}
				return nil, protocol.ErrUnqualifiedData
			}
			index = i + 1
		}
	}
	if index != len(data)-1 {
		buf = append(buf, data[index:len(data)-1]...)
	}
	return buf, nil
}

func escape(data []byte) []byte {
	const (
		flag0x7d = 0x7d // 转义标志 0x7d -> 0x7d 0x01
		flag0x7e = 0x7e // 转义标志 0x7e -> 0x7d 0x02
	)
	// 快速路径：无 0x7d/0x7e 时直接一次分配并整块拷贝
	if bytes.IndexByte(data, flag0x7e) == -1 && bytes.IndexByte(data, flag0x7d) == -1 {
		out := make([]byte, len(data)+2)
		out[0] = flag0x7e
		copy(out[1:], data)
		out[len(out)-1] = flag0x7e
		return out
	}

	buf := make([]byte, 0, len(data)+4)
	buf = append(buf, flag0x7e)
	index := 0
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case flag0x7e:
			buf = append(buf, data[index:i]...)
			buf = append(buf, flag0x7d, 0x02)
			index = i + 1
		case flag0x7d:
			buf = append(buf, data[index:i]...)
			buf = append(buf, flag0x7d, 0x01)
			index = i + 1
		default:
		}
	}
	buf = append(buf, data[index:]...)
	buf = append(buf, flag0x7e)
	return buf
}
