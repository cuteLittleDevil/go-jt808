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
			currentOffset = segment.DataOffset + segment.DataLength
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
				yield(nil)
			} else if errors.Is(err, _errNotStreamData) { // 不是流数据格式的 换一个格式试一试
				if err := p.stageJT808Data(); err == nil {
					yield(nil)
				} else if !errors.Is(err, ErrDataInconsistency) {
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
	if len(p.historyData) >= headLen+bodyLen {
		stream.OnInitEvent(p)
		p.ProgressStage = ProgressStageStreamData
		name := stream.GetFileName()
		pack, ok := p.Record[name]
		if !ok {
			return errors.Join(fmt.Errorf("name[%s] is not exist record[%v]",
				name, p.Record), ErrDataInconsistency)
		}
		defer func() {
			p.ExtensionFields.CurrentPackage = pack
			p.Record[name] = pack
			p.historyData = p.historyData[headLen+bodyLen:]
		}()
		offset, dataLen := stream.GetDataOffsetAndLen()
		pack.Offset = offset
		pack.OffsetRecord[offset] = dataLen
		pack.OffsetDataRecord[offset] = p.historyData[headLen : headLen+bodyLen]
		pack.CurrentSize += uint32(bodyLen)
		if pack.CurrentSize == pack.FileSize {
			pack.StreamHead = p.historyData[:headLen]
			keys := make([]int, 0)
			for key := range maps.Keys(pack.OffsetDataRecord) {
				keys = append(keys, key)
			}
			sort.Ints(keys)
			for _, key := range keys {
				pack.StreamBody = append(pack.StreamBody, pack.OffsetDataRecord[key]...)
			}
			p.ProgressStage = ProgressStageStreamDataComplete
		}
		return nil
	}
	return ErrInsufficientDataLen
}

func (p *PackageProgress) parseJT808Message() (*jt808.JTMessage, error) {
	if len(p.historyData) < 10 {
		return nil, ErrInsufficientDataLen
	}
	const sign = 0x7e
	index := bytes.IndexFunc(p.historyData[1:], func(r rune) bool {
		return r == sign
	})
	if index == -1 {
		return nil, ErrInsufficientDataLen
	}
	index += 2
	jtMsg := jt808.NewJTMessage()
	if err := jtMsg.Decode(p.historyData[:index]); err != nil {
		return nil, fmt.Errorf("%w [%x]", err, p.historyData[:index])
	}
	p.historyData = p.historyData[index:]
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
