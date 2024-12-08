package attachment

import (
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"sync"
)

type BaseJT808DataHandler[T1210, T1211, T1212 JT808Handler] struct {
	T0x1210 T1210
	T0x1211 T1211
	T0x1212 T1212
	Command consts.JT808CommandType
	jtMsg   *jt808.JTMessage
	once    sync.Once
	head    *jt808.Header
	seq     uint16
}

func (d *BaseJT808DataHandler[T1210, T1211, T1212]) Parse(jtMsg *jt808.JTMessage) error {
	d.once.Do(func() {
		d.head = jtMsg.Header
	})
	d.jtMsg = jtMsg
	d.Command = consts.JT808CommandType(jtMsg.Header.ID)
	switch d.Command {
	case consts.T1210AlarmAttachInfoMessage:
		d.once.Do(func() {
			d.head = jtMsg.Header
		})
		return d.T0x1210.Parse(jtMsg)
	case consts.T1211FileInfoUpload:
		return d.T0x1211.Parse(jtMsg)
	case consts.T1212FileUploadComplete:
		return d.T0x1212.Parse(jtMsg)
	default:
	}
	return errors.Join(fmt.Errorf("%s", d.Command), ErrUnknownCommand)
}

func (d *BaseJT808DataHandler[T1210, T1211, T1212]) ReplyData() ([]byte, error) {
	type Handle struct {
		ReplyBody    func(jtMsg *jt808.JTMessage) ([]byte, error)
		replyCommand consts.JT808CommandType
	}
	handle := Handle{}
	switch d.Command {
	case consts.T1210AlarmAttachInfoMessage:
		handle.ReplyBody = d.T0x1210.ReplyBody
		handle.replyCommand = d.T0x1210.ReplyProtocol()
	case consts.T1211FileInfoUpload:
		handle.ReplyBody = d.T0x1211.ReplyBody
		handle.replyCommand = d.T0x1211.ReplyProtocol()
	case consts.T1212FileUploadComplete:
		handle.ReplyBody = d.T0x1212.ReplyBody
		handle.replyCommand = d.T0x1212.ReplyProtocol()
	default:
		return nil, ErrUnknownCommand
	}
	d.head.ReplyID = uint16(handle.replyCommand)
	d.head.PlatformSerialNumber = d.seq
	d.seq++
	body, err := handle.ReplyBody(d.jtMsg)
	if err != nil {
		return nil, err
	}
	return d.head.Encode(body), nil
}

func (d *BaseJT808DataHandler[T1210, T1211, T1212]) OnPackageProgressEvent(progress *PackageProgress) {
	switch d.Command {
	case consts.T1210AlarmAttachInfoMessage:
		progress.ProgressStage = ProgressStageInit
	case consts.T1211FileInfoUpload:
		progress.ProgressStage = ProgressStageStart
	case consts.T1212FileUploadComplete:
		progress.ProgressStage = ProgressStageComplete
	}
}

func (d *BaseJT808DataHandler[T1210, T1211, T1212]) CreateStreamDataHandler() StreamDataHandler {
	return newBaseStreamDataHandle()
}

type standardJT808DataHandle struct {
	BaseJT808DataHandler[*model.T0x1210, *model.T0x1211, *model.T0x1212]
}

func newStandardJT808DataHandle(asType consts.ActiveSafetyType) *standardJT808DataHandle {
	return &standardJT808DataHandle{
		BaseJT808DataHandler: BaseJT808DataHandler[*model.T0x1210, *model.T0x1211, *model.T0x1212]{
			T0x1210: &model.T0x1210{
				P9208AlarmSign: model.P9208AlarmSign{
					ActiveSafetyType: asType,
				},
			},
			T0x1211: &model.T0x1211{},
			T0x1212: &model.T0x1212{},
		}}
}

func (s *standardJT808DataHandle) OnPackageProgressEvent(progress *PackageProgress) {
	s.BaseJT808DataHandler.OnPackageProgressEvent(progress)
	switch s.Command {
	case consts.T1210AlarmAttachInfoMessage:
		for _, v := range s.T0x1210.T0x1210AlarmItemList {
			progress.Record[v.FileName] = &Package{
				FileName:         v.FileName,
				FileSize:         v.FileSize,
				CurrentSize:      0,
				StreamHead:       nil,
				StreamBody:       nil,
				OffsetDataRecord: map[int][]byte{},
				OffsetRecord:     map[int]int{},
			}
		}
	case consts.T1212FileUploadComplete:
		name := s.T0x1212.FileName
		if v, ok := progress.Record[name]; ok {
			s.T0x1212.P0x9212RetransmitPacketList = v.StatisticalMissSegments()
			if len(s.T0x1212.P0x9212RetransmitPacketList) > 0 {
				progress.ProgressStage = ProgressStageSupplementary
			}
			return
		}
	}
}

func (s *standardJT808DataHandle) CreateStreamDataHandler() StreamDataHandler {
	if s.T0x1210.P9208AlarmSign.ActiveSafetyType == consts.ActiveSafetyHLJ {
		return newHeiBiaoStreamDataHandle()
	}
	return newBaseStreamDataHandle()
}
