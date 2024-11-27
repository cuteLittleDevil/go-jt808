package attachment

import (
	"bytes"
	"encoding/binary"
)

type suBiaoStreamDataHandle struct {
	// FrameSign 帧标识 固定0x30 0x31 0x63 0x64
	FrameSign uint32
	// FileName 文件名 [50]byte
	FileName string
	// DataOffset 数据偏移量
	DataOffset uint32
	// DataLen 数据长度
	DataLen uint32
	// Data 数据体 默认长度64k 文件小于64k则为实际长度
	Data []byte
}

func newSuBiaoStreamDataHandle() *suBiaoStreamDataHandle {
	return &suBiaoStreamDataHandle{}
}

func (s *suBiaoStreamDataHandle) HasMinHeadLen(data []byte) bool {
	return len(data) >= 62
}

func (s *suBiaoStreamDataHandle) HasStreamData(data []byte) bool {
	return bytes.Contains(data, []byte{0x30, 0x31, 0x63, 0x64}) // 808543076 = 0x30 0x31 0x63 0x64
}

func (s *suBiaoStreamDataHandle) GetLen(data []byte) (headLen int, bodyLen int) {
	s.FrameSign = binary.BigEndian.Uint32(data[0:4])
	s.FileName = string(bytes.Trim(data[4:54], "\x00"))
	s.DataOffset = binary.BigEndian.Uint32(data[54:58])
	s.DataLen = binary.BigEndian.Uint32(data[58:62])
	s.Data = data[62:]
	return 62, int(s.DataLen)
}

func (s *suBiaoStreamDataHandle) GetFileName() string {
	return s.FileName
}

func (s *suBiaoStreamDataHandle) GetDataOffsetAndLen() (offset int, dataLen int) {
	return int(s.DataOffset), int(s.DataLen)
}
