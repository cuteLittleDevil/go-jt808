package service

import (
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type connection struct {
	conn     *net.TCPConn
	handles  map[consts.JT808CommandType]Handler
	stopOnce sync.Once
	// stopChan 读终端数据异常时退出
	stopChan chan struct{}
	// finallyCompleteChan 读终端数据异常后 等待主动下发的全部指令完成或者超时
	finallyCompleteChan   chan struct{}
	msgChan               chan *Message
	activeMsgChan         chan *ActiveMessage
	activeMsgCompleteChan chan *Message
	reissuePackChan       chan *Message
	// platformSerialNumber 平台流水号 到了math.MaxUint16后+1重新变成0
	platformSerialNumber uint16
	joinFunc             func(message *Message, activeChan chan<- *ActiveMessage) (string, error)
	leaveFunc            func(key string)
	key                  string
	filter               bool
	// activeUnfinishedSum 当前主动下发未完成的指令数量
	activeUnfinishedSum int32
	terminalEvent       TerminalEventer
}

func newConnection(conn *net.TCPConn, handles map[consts.JT808CommandType]Handler, terminalEvent TerminalEventer, filter bool,
	join func(message *Message, activeChan chan<- *ActiveMessage) (string, error), leave func(key string)) *connection {
	return &connection{
		conn:                  conn,
		handles:               handles,
		stopOnce:              sync.Once{},
		stopChan:              make(chan struct{}),
		finallyCompleteChan:   make(chan struct{}),
		msgChan:               make(chan *Message, 10),
		activeMsgChan:         make(chan *ActiveMessage, 10),
		activeMsgCompleteChan: make(chan *Message, 10),
		reissuePackChan:       make(chan *Message, 10),
		platformSerialNumber:  uint16(0),
		joinFunc:              join,
		leaveFunc:             leave,
		filter:                filter,
		activeUnfinishedSum:   int32(0),
		terminalEvent:         terminalEvent,
	}
}

func (c *connection) run() {
	go c.reader()
	go c.write()
}

func (c *connection) reader() {
	var (
		// 消息体长度最大为 10bit 也就是 1023 的字节
		curData = make([]byte, 1023)
		pack    = newPackageParse()
		join    = false
	)

	defer func() {
		clear(curData)
		pack.clear()
		c.stop()
	}()

	for {
		select {
		case <-c.stopChan:
			return
		default:
			if n, err := c.conn.Read(curData); err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
					slog.Debug("connection close",
						slog.Bool("join", join),
						slog.Any("platform num", c.platformSerialNumber),
						slog.Any("err", err))
					return
				}
				slog.Error("read data",
					slog.Bool("join", join),
					slog.Any("platform num", c.platformSerialNumber),
					slog.Any("err", err))
				return
			} else if n > 0 {
				effectiveData := curData[:n]
				msgs, err := pack.parse(effectiveData)
				if err != nil {
					slog.Error("parse data",
						slog.Bool("join", join),
						slog.Any("platform num", c.platformSerialNumber),
						slog.String("effective data", fmt.Sprintf("%x", effectiveData)),
						slog.Any("err", err))
					return
				}
				if len(msgs) > 0 {
					if !join {
						if err := c.joinHandle(msgs[0]); err == nil {
							join = true
						} else if errors.Is(err, _errKeyExist) {
							slog.Warn("key",
								slog.String("effective data", fmt.Sprintf("%x", effectiveData)),
								slog.Any("err", err))
							return
						}
					}

					c.handleMessages(msgs)
				}
			}
		}
	}
}

func (c *connection) joinHandle(msg *Message) error {
	key, err := c.joinFunc(msg, c.activeMsgChan)
	if err == nil {
		c.key = key
	}

	c.terminalEvent.OnJoinEvent(msg, key, err)
	return err
}

func (c *connection) handleMessages(msgs []*Message) {
	for _, msg := range msgs {
		msg.Key = c.key
		if handler, ok := c.handles[msg.Command]; ok {
			msg.Handler = handler
			if msg.Command == consts.P8003ReissueSubcontractingRequest {
				c.reissuePackChan <- msg
				continue
			}
			c.onReadExecutionEvent(msg)
		} else {
			c.terminalEvent.OnNotSupportedEvent(msg)
			continue
		}
		c.msgChan <- msg
	}
}

func (c *connection) write() {
	record := map[uint16]*ActiveMessage{}
	defer func() {
		close(c.activeMsgChan)
		close(c.activeMsgCompleteChan)
		close(c.reissuePackChan)
		close(c.msgChan)
		clear(record)
		clear(c.handles)
	}()

	for {
		select {
		case <-c.finallyCompleteChan: // 如果现在有平台主动下发的 需要等待完成在退出
			return
		case activeMsg := <-c.activeMsgChan: // 平台主动下发的
			atomic.AddInt32(&c.activeUnfinishedSum, 1)
			c.onActiveEvent(activeMsg, record)

		case msg := <-c.activeMsgCompleteChan: // 平台主动下发的完成情况
			seq := msg.ExtensionFields.PlatformSeq
			// 超时的情况一定执行一次 如果完成了 还可能执行一次
			if v, ok := record[seq]; ok {
				c.onActiveEventComplete(msg, v)
				delete(record, seq)
				atomic.AddInt32(&c.activeUnfinishedSum, -1)
			}

		case subPackMsg := <-c.reissuePackChan: // 分包补传的
			c.onSubPackReplyEvent(subPackMsg)

		case msg := <-c.msgChan: // 终端上传的
			if len(record) > 0 && msg.hasComplete() { // 说明现在有主动的请求 等待回复中
				if c.onActiveRespondEvent(record, msg) {
					continue
				}
			}
			if msg.hasComplete() || !c.filter { // 默认完整包才触发回复
				c.defaultReplyEvent(msg)
			}
		}
	}
}

func (c *connection) stop() {
	c.stopOnce.Do(func() {
		c.leaveFunc(c.key)
		c.terminalEvent.OnLeaveEvent(c.key)
		close(c.stopChan)
		_ = c.conn.Close()
		for atomic.LoadInt32(&c.activeUnfinishedSum) != 0 {
			time.Sleep(time.Second)
		}
		close(c.finallyCompleteChan)
	})
}

func (c *connection) defaultReplyEvent(msg *Message) {
	if has := msg.HasReply(); !has {
		return
	}
	body, err := msg.ReplyBody(msg.JTMessage)
	if err != nil {
		slog.Warn("reply body fail",
			slog.String("terminal data", fmt.Sprintf("%x", msg.ExtensionFields.TerminalData)),
			slog.Any("err", err))
		return
	}
	header := msg.JTMessage.Header
	header.ReplyID = uint16(msg.ReplyProtocol())
	seq := c.curSeq()
	header.PlatformSerialNumber = seq
	data := header.Encode(body)

	if _, err = c.conn.Write(data); err != nil {
		slog.Warn("write fail",
			slog.String("data", fmt.Sprintf("%x", data)),
			slog.Any("err", err))
		msg.ExtensionFields.Err = errors.Join(ErrWriteDataFail, err)
	}
	msg.ExtensionFields.PlatformCommand = msg.ReplyProtocol()
	msg.ExtensionFields.PlatformSeq = seq
	msg.ExtensionFields.PlatformData = data
	c.onWriteExecutionEvent(msg)
}

func (c *connection) onSubPackReplyEvent(msg *Message) {
	header := msg.JTMessage.Header
	seq := c.curSeq()
	header.PlatformSerialNumber = seq
	header.ReplyID = uint16(consts.P8003ReissueSubcontractingRequest)
	data := header.Encode(msg.JTMessage.Body)

	if _, err := c.conn.Write(data); err != nil {
		slog.Warn("write fail",
			slog.String("data", fmt.Sprintf("%x", data)),
			slog.Any("err", err))
		msg.ExtensionFields.Err = errors.Join(ErrWriteDataFail, err)
	}
	msg.ExtensionFields.PlatformSeq = seq
	msg.ExtensionFields.PlatformData = data
	c.onWriteExecutionEvent(msg)
}

func (c *connection) onActiveEvent(activeMsg *ActiveMessage, record map[uint16]*ActiveMessage) {
	header := activeMsg.header
	seq := c.curSeq()
	header.PlatformSerialNumber = seq
	header.ReplyID = uint16(activeMsg.Command)
	data := header.Encode(activeMsg.Body)
	activeMsg.ExtensionFields = struct {
		PlatformSeq uint16 `json:"platformSeq,omitempty"`
		Data        []byte `json:"data,omitempty"`
	}{
		PlatformSeq: seq,
		Data:        data,
	}
	_, err := c.conn.Write(data)
	replyMsg := newActiveMessage(seq, activeMsg.Command, data, err)
	if v, ok := c.handles[activeMsg.Command]; ok {
		replyMsg.Handler = v
	}
	c.onReadExecutionEvent(replyMsg)

	activeMsg.convertMessage = replyMsg
	record[seq] = activeMsg
	if err != nil {
		replyMsg.ExtensionFields.Err = errors.Join(ErrWriteDataFail, err)
		c.activeMsgCompleteChan <- replyMsg
	} else {
		duration := 3 * time.Second
		if activeMsg.OverTimeDuration > 0 {
			duration = activeMsg.OverTimeDuration
		}
		go func(overtimeMsg *Message, timeout time.Duration) {
			time.Sleep(timeout)
			overtimeMsg.ExtensionFields.Err = errors.Join(ErrWriteDataOverTime,
				fmt.Errorf("overtime is [%.2f]second", duration.Seconds()))
			c.activeMsgCompleteChan <- overtimeMsg
		}(replyMsg, duration)
	}
}

func (c *connection) onActiveRespondEvent(record map[uint16]*ActiveMessage, msg *Message) bool {
	type respond struct {
		JT808Handler
		HasRespondFunc func(seq uint16) bool
	}
	tmp := respond{
		JT808Handler:   nil,
		HasRespondFunc: nil,
	}
	switch msg.Command {
	case consts.T0001GeneralRespond:
		t0x0001 := &model.T0x0001{}
		tmp.JT808Handler = t0x0001
		tmp.HasRespondFunc = func(seq uint16) bool {
			// 如果是这些命令的话 等待后续应答 如 8801 -> 8805
			switch record[seq].Command {
			case consts.P8801CameraShootImmediateCommand, consts.P9003QueryTerminalAudioVideoProperties,
				consts.P9205QueryResourceList, consts.P9206FileUploadInstructions:
				return false
			default:
				return seq == t0x0001.SerialNumber
			}
		}
	case consts.T0104QueryParameter:
		t0x0104 := &model.T0x0104{}
		tmp.JT808Handler = t0x0104
		tmp.HasRespondFunc = func(seq uint16) bool {
			return seq == t0x0104.RespondSerialNumber
		}
	case consts.T1003UploadAudioVideoAttr:
		t0x1003 := &model.T0x1003{}
		tmp.JT808Handler = t0x1003
		tmp.HasRespondFunc = func(_ uint16) bool {
			return true
		}
	case consts.T0201QueryLocation:
		t0x0201 := &model.T0x0201{}
		tmp.JT808Handler = t0x0201
		tmp.HasRespondFunc = func(seq uint16) bool {
			return seq == t0x0201.RespondSerialNumber
		}
	case consts.T0302QuestionAnswer:
		t0x0302 := &model.T0x0302{}
		tmp.JT808Handler = t0x0302
		tmp.HasRespondFunc = func(seq uint16) bool {
			return seq == t0x0302.SerialNumber
		}
	case consts.T1205UploadAudioVideoResourceList:
		t0x1205 := &model.T0x1205{}
		tmp.JT808Handler = t0x1205
		tmp.HasRespondFunc = func(seq uint16) bool {
			return seq == t0x1205.SerialNumber
		}
	case consts.T1206FileUploadCompleteNotice:
		t0x1206 := &model.T0x1206{}
		tmp.JT808Handler = t0x1206
		tmp.HasRespondFunc = func(seq uint16) bool {
			return seq == t0x1206.RespondSerialNumber
		}
	case consts.T0805CameraShootImmediately:
		t0x0805 := &model.T0x0805{}
		tmp.JT808Handler = t0x0805
		tmp.HasRespondFunc = func(seq uint16) bool {
			return seq == t0x0805.RespondSerialNumber
		}
	}
	if tmp.HasRespondFunc != nil {
		if err := tmp.Parse(msg.JTMessage); err != nil {
			slog.Warn("parse fail",
				slog.String("terminal data", fmt.Sprintf("%x", msg.ExtensionFields.TerminalData)),
				slog.Any("err", err))
			return true
		}
		for k := range record {
			if tmp.HasRespondFunc(k) {
				msg.ExtensionFields.PlatformSeq = k
				msg.ExtensionFields.TerminalCommand = tmp.Protocol()
				c.activeMsgCompleteChan <- msg
				return true
			}
		}
	}

	return false
}

func (c *connection) onActiveEventComplete(msg *Message, activeMsg *ActiveMessage) {
	msg.ExtensionFields.PlatformData = activeMsg.ExtensionFields.Data
	msg.ExtensionFields.PlatformCommand = activeMsg.Command
	msg.ExtensionFields.ActiveSend = true
	// 如0x8300 -> 0x0001
	// 超时的情况 msg就是平台下发的指令 如0x8300
	// 完成的情况 msg就是平台需要回复的 如0x0001
	c.onWriteExecutionEvent(msg)
	if msg.ExtensionFields.Err == nil {
		// 完成的情况 前面已经触发了如0x0001的回调 在触发0x8300的回调
		c.onWriteExecutionEvent(activeMsg.convertMessage)
	}
	activeMsg.replyChan <- msg
}

func (c *connection) onReadExecutionEvent(msg *Message) {
	if c.filter && !msg.hasComplete() {
		return
	}
	if msg.Handler == nil {
		slog.Warn("Handler is nil",
			slog.String("read", msg.Header.String()))
		return
	}
	msg.Handler.OnReadExecutionEvent(msg)

	c.terminalEvent.OnReadExecutionEvent(msg)
}

func (c *connection) onWriteExecutionEvent(msg *Message) {
	if c.filter && !msg.hasComplete() {
		return
	}
	if msg.Handler == nil {
		slog.Warn("Handler is nil",
			slog.String("write", msg.Header.String()))
		return
	}
	msg.Handler.OnWriteExecutionEvent(*msg)

	c.terminalEvent.OnWriteExecutionEvent(*msg)
}

func (c *connection) curSeq() uint16 {
	defer func() {
		c.platformSerialNumber++
	}()
	return c.platformSerialNumber
}
