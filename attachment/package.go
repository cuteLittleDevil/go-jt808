package attachment

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"maps"
	"sort"
)

type (
	PackageProgress struct {
		// ProgressStage 当前进度
		ProgressStage
		// Record 目前上传数据的记录 key=文件名 value=数据包上传情况
		Record map[string]*Package
		// ExtensionFields 扩展字段信息
		ExtensionFields ExtensionFields
		historyData     []byte
		handle          DataHandler
	}

	ExtensionFields struct {
		// CurrentPackage 当前完成的包情况 也在record中的
		CurrentPackage *Package
		// RecentTerminalMessage 终端主动上传的808数据
		RecentTerminalMessage *jt808.JTMessage
		// RecentPlatformData 平台下发的数据
		RecentPlatformData []byte `json:"-"`
		// ActiveSafetyType 主动安全标准
		consts.ActiveSafetyType
		// Err 异常情况
		Err error
		// TerminalPhoneNo 手机号 解析1210时记录
		TerminalPhoneNo string
		// Debug 记录tcp异常的详情
		Debug struct {
			RemoteAddr  string
			NetErr      error
			HistoryData []byte
		}
	}

	Package struct {
		// FileName 文件名称
		FileName string
		// FileSize 文件大小
		FileSize uint32
		// CurrentSize 已经上传的文件大小
		CurrentSize uint32
		// StreamHead 数据头部 当数据完成时候记录
		StreamHead []byte
		// StreamBody 数据体 当数据完成时候记录
		StreamBody []byte
		// Offset 偏移
		Offset int
		// OffsetDataRecord 偏移的数据记录 key=偏移 value=数据
		OffsetDataRecord map[int][]byte
		// OffsetRecord 偏移的记录 key=偏移 value=文件大小
		OffsetRecord map[int]int
	}
)

func (p *Package) StatisticalMissSegments() []model.P0x9212RetransmitPacket {
	// 不需要补包的情况 收到的文件大小=最终文件大小
	if p.CurrentSize == p.FileSize {
		return nil
	}
	var (
		missSegments  []model.P0x9212RetransmitPacket
		currentOffset = uint32(0)
	)

	segments := make([]model.P0x9212RetransmitPacket, 0, len(p.OffsetRecord))
	for offset, dataLen := range p.OffsetRecord {
		segments = append(segments, model.P0x9212RetransmitPacket{
			DataOffset: uint32(offset),
			DataLength: uint32(dataLen),
		})
	}
	// 看看漏掉了哪些包
	if len(segments) > 0 {
		sort.Slice(segments, func(i, j int) bool {
			return segments[i].DataOffset < segments[j].DataOffset
		})
		for _, segment := range segments {
			if currentOffset < segment.DataOffset {
				missSegments = append(missSegments, model.P0x9212RetransmitPacket{
					DataOffset: currentOffset,
					DataLength: segment.DataOffset - currentOffset,
				})
			}
			if end := segment.DataOffset + segment.DataLength; end > currentOffset {
				currentOffset = end
			}
		}
	}

	// 看看最后的包有没有漏掉
	if currentOffset < p.FileSize {
		missSegments = append(missSegments, model.P0x9212RetransmitPacket{
			DataOffset: currentOffset,
			DataLength: p.FileSize - currentOffset,
		})
	}

	return missSegments
}

func (p *PackageProgress) iter() func(func(err error) bool) {
	return func(yield func(err error) bool) {
		for len(p.historyData) > 0 {
			if err := p.stageStreamData(); err == nil {
				if !yield(nil) {
					return
				}
			} else if errors.Is(err, _errNotStreamData) { // 不是流数据格式的 换一个格式试一试
				if err := p.stageJT808Data(); err == nil {
					if !yield(nil) {
						return
					}
				} else {
					yield(err)
					return
				}
			} else {
				yield(err)
				return
			}
		}
	}
}

func (p *PackageProgress) stageStreamData() error {
	stream := p.handle.CreateStreamDataHandler()
	if !stream.HasStreamData(p.historyData) {
		return _errNotStreamData
	}
	if !stream.HasMinHeadLen(p.historyData) {
		return ErrInsufficientDataLen
	}
	headLen, bodyLen := stream.Parse(p.historyData)
	if headLen < 0 || bodyLen < 0 {
		return ErrDataInconsistency
	}
	if len(p.historyData) < headLen+bodyLen {
		return ErrInsufficientDataLen
	}
	stream.OnInitEvent(p)
	p.ProgressStage = ProgressStageStreamData
	name := stream.GetFileName()
	pack, ok := p.Record[name]
	if !ok {
		return errors.Join(fmt.Errorf("name[%s] is not exist record[%v]",
			name, p.Record), ErrDataInconsistency)
	}
	chunk := bytes.Clone(p.historyData[headLen : headLen+bodyLen])
	offset, dataLen := stream.GetDataOffsetAndLen()
	recordStreamChunk(pack, offset, dataLen, bodyLen, chunk)
	if pack.CurrentSize == pack.FileSize && len(pack.StreamBody) == 0 {
		assembleStreamBody(pack, bytes.Clone(p.historyData[:headLen]))
		p.ProgressStage = ProgressStageStreamDataComplete
	}
	p.ExtensionFields.CurrentPackage = pack
	p.Record[name] = pack
	p.consumeHistory(headLen + bodyLen)
	return nil
}

func (p *PackageProgress) consumeHistory(n int) {
	if n == len(p.historyData) {
		p.historyData = p.historyData[:0]
		return
	}
	p.historyData = p.historyData[n:]
}

func recordStreamChunk(pack *Package, offset, dataLen, bodyLen int, chunk []byte) {
	pack.Offset = offset
	if prevLen, exists := pack.OffsetRecord[offset]; exists {
		// 重复/补传同一偏移分片时只补差量，避免 CurrentSize 双计超过 FileSize
		if dataLen > prevLen {
			pack.CurrentSize += uint32(dataLen - prevLen)
			pack.OffsetRecord[offset] = dataLen
			pack.OffsetDataRecord[offset] = chunk
		}
		return
	}
	pack.OffsetRecord[offset] = dataLen
	pack.OffsetDataRecord[offset] = chunk
	pack.CurrentSize += uint32(bodyLen)
}

func assembleStreamBody(pack *Package, head []byte) {
	pack.StreamHead = head
	keys := make([]int, 0, len(pack.OffsetDataRecord))
	for key := range maps.Keys(pack.OffsetDataRecord) {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	pack.StreamBody = make([]byte, 0, pack.FileSize)
	for _, key := range keys {
		pack.StreamBody = append(pack.StreamBody, pack.OffsetDataRecord[key]...)
	}
}

func (p *PackageProgress) parseJT808Message() (*jt808.JTMessage, error) {
	const minHeadLen = 12 // jt808报文最小头部
	if len(p.historyData) < minHeadLen {
		return nil, ErrInsufficientDataLen
	}
	const sign = 0x7e
	index := bytes.IndexByte(p.historyData[1:], sign)
	if index == -1 {
		return nil, ErrInsufficientDataLen
	}
	index += 2
	jtMsg := jt808.NewJTMessage()
	if err := jtMsg.Decode(p.historyData[:index]); err != nil {
		return nil, fmt.Errorf("%w [%x]", err, p.historyData[:index])
	}
	if index == len(p.historyData) {
		p.historyData = p.historyData[:0]
	} else {
		p.historyData = p.historyData[index:]
	}
	return jtMsg, nil
}

func (p *PackageProgress) stageJT808Data() error {
	jtMsg, err := p.parseJT808Message()
	if err != nil {
		return err
	}
	if err := p.handle.Parse(jtMsg); err != nil {
		return err
	}
	p.ExtensionFields.RecentTerminalMessage = jtMsg
	p.ExtensionFields.TerminalPhoneNo = jtMsg.Header.TerminalPhoneNo
	p.handle.OnPackageProgressEvent(p)
	return nil
}

func (p *PackageProgress) hasJT808Reply() bool {
	switch p.ProgressStage {
	case ProgressStageInit, ProgressStageStart, ProgressStageComplete, ProgressStageSupplementary:
		return true
	default:
	}
	return false
}
