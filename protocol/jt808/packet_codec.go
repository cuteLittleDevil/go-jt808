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
	// 快速路径 没有需要转义的
	if !bytes.ContainsRune(data, beforeEscape) {
		return data[1 : len(data)-1], nil
	}

	buf := new(bytes.Buffer)
	index := 1
	for i := 1; i < len(data)-1; i++ {
		if v := data[i]; v == beforeEscape {
			i++
			switch data[i] {
			case 0x01:
				buf.Write(data[index : i-1])
				buf.WriteByte(beforeEscape)
			case 0x02:
				buf.Write(data[index : i-1])
				buf.WriteByte(afterRecover)
			default:
				// 兼容一下设备校验码不转义的情况
				if i == len(data)-1 {
					buf.Write(data[index : len(data)-1])
					return buf.Bytes(), nil
				}
				return nil, protocol.ErrUnqualifiedData
			}
			index = i + 1
		}
	}
	if index != len(data)-1 {
		buf.Write(data[index : len(data)-1])
	}
	return buf.Bytes(), nil
}

func escape(data []byte) []byte {
	const (
		flag0x7d = 0x7d // 转义标志 0x7d -> 0x7d 0x01
		flag0x7e = 0x7e // 转义标志 0x7e -> 0x7d 0x02
	)
	buf := new(bytes.Buffer)
	buf.WriteByte(flag0x7e)
	index := 0
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case flag0x7e:
			buf.Write(data[index:i])
			buf.Write([]byte{0x7d, 0x02})
			index = i + 1
		case flag0x7d:
			buf.Write(data[index:i])
			buf.Write([]byte{0x7d, 0x01})
			index = i + 1
		default:
		}
	}
	buf.Write(data[index:])
	buf.WriteByte(flag0x7e)
	return buf.Bytes()
}
