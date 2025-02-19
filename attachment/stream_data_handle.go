package attachment

import (
	"bytes"
	"encoding/binary"
)

type baseStreamDataHandle struct {
	// FrameSign 帧标识 固定0x30 0x31 0x63 0x64
	FrameSign uint32
	// FileName 文件名 [50]byte
	// 苏标 文件名50
	// 黑标 文件名长度byte 文件名名称
	// 广东标 文件名50
	// 湖南标 文件名50
	// 四川标 文件名50
	FileName string
	// DataOffset 数据偏移量
	DataOffset uint32
	// DataLen 数据长度
	DataLen uint32
	// Data 数据体 默认长度64k 文件小于64k则为实际长度
	Data []byte
}

func newBaseStreamDataHandle() *baseStreamDataHandle {
	return &baseStreamDataHandle{}
}

func (s *baseStreamDataHandle) HasStreamData(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0x30, 0x31, 0x63, 0x64}) // 808543076 = 0x30 0x31 0x63 0x64
}

func (s *baseStreamDataHandle) HasMinHeadLen(data []byte) bool {
	return len(data) >= 62
}

func (s *baseStreamDataHandle) Parse(data []byte) (headLen int, bodyLen int) {
	s.FrameSign = binary.BigEndian.Uint32(data[0:4])
	s.FileName = string(bytes.Trim(data[4:54], "\x00"))
	s.DataOffset = binary.BigEndian.Uint32(data[54:58])
	s.DataLen = binary.BigEndian.Uint32(data[58:62])
	s.Data = data[62:]
	return 62, int(s.DataLen)
}

func (s *baseStreamDataHandle) OnInitEvent(_ *PackageProgress) {

}

func (s *baseStreamDataHandle) GetFileName() string {
	return s.FileName
}

func (s *baseStreamDataHandle) GetDataOffsetAndLen() (offset int, dataLen int) {
	return int(s.DataOffset), int(s.DataLen)
}

type heiBiaoStreamDataHandle struct {
	baseStreamDataHandle
	// FileNameLen 文件名长度
	FileNameLen byte
	// FileName 文件名
	FileName string
}

func newHeiBiaoStreamDataHandle() *heiBiaoStreamDataHandle {
	return &heiBiaoStreamDataHandle{}
}

func (h *heiBiaoStreamDataHandle) HasMinHeadLen(data []byte) bool {
	const minLen = 4 + 1 // 固定校验码和读取头部的长度
	if len(data) < minLen {
		return false
	}
	h.FileNameLen = data[4]
	return len(data) >= 4+1+int(h.FileNameLen)+4+4
}

func (h *heiBiaoStreamDataHandle) Parse(data []byte) (headLen int, bodyLen int) {
	h.FrameSign = binary.BigEndian.Uint32(data[0:4])
	h.FileNameLen = data[4]
	start, end := 5, 5+int(h.FileNameLen)
	h.FileName = string(bytes.Trim(data[start:end], "\x00"))
	start = end
	end += 4
	const sign = 808543076 // 0x30 0x31 0x63 0x64 固定标识
	h.baseStreamDataHandle = baseStreamDataHandle{
		FrameSign:  sign,
		FileName:   h.FileName,
		DataOffset: binary.BigEndian.Uint32(data[start:end]),
		DataLen:    binary.BigEndian.Uint32(data[start+4 : end+4]),
		Data:       data[end:],
	}
	return 4 + 1 + int(h.FileNameLen) + 4 + 4, int(h.DataLen)
}
