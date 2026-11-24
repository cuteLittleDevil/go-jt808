package service

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type (
	// connection 表示与一个终端（车载设备）保持的长连接会话对象。
	// 它是 JT808 协议服务端的核心会话管理结构，负责：
	//   - 终端上下行消息的处理
	//   - 平台主动下发指令的管理与超时控制
	//   - 流水号维护
	//   - 终端上下线事件回调
	//   - 分包补传处理等
	connection struct {
		connectionParams

		// stopWriteChan 连接关闭时由 stop() 关闭，通知 write 协程排空并等待在途主动消息完成.
		stopWriteChan chan struct{}
		// writeDoneChan write 协程完全退出时关闭，通知 stop() 与超时定时器停止发送.
		writeDoneChan chan struct{}
		// terminalUplinkMsgChan 终端上报的普通消息（上行消息）会放入此通道，
		// 随后由 write 协程统一处理回复逻辑.
		terminalUplinkMsgChan chan *Message
		// reissuePackChan 需要补发的分包消息通道.
		reissuePackChan chan *Message
		// platformSerialNumber 平台侧流水号（低 16 位 0~65535 循环使用，原子读写避免日志竞态）.
		platformSerialNumber atomic.Uint32

		// activeMsgChan 平台主动下发给终端的指令会包装成 ActiveMessage 放入此通道.
		activeMsgChan chan *ActiveMessage
		// activeTimeoutChan 主动下发超时定时器通知通道（只传流水号且永不关闭，避免关闭竞态 panic）.
		activeTimeoutChan chan uint16
		// activePending 当前在途主动下发数量（仅用于告知 reader() 按需拷贝低频应答，热路径不拷贝）.
		activePending atomic.Int32
		// sessionHeader 首包深拷贝得到的独立 Header（含独立 bcdTerminalPhoneNo），供会话与下行回复零拷贝复用.
		sessionHeader atomic.Pointer[jt808.Header]
		// key 当前连接对应的终端唯一标识（默认是 SIM 卡号），
		// 由 onJoinEvent 回调函数返回后赋值.
		key string
	}

	// activeRecord 记录一条在途主动下发指令及其可取消的超时定时器，仅由 write 协程访问.
	activeRecord struct {
		msg     *ActiveMessage
		timer   *time.Timer
		timeout time.Duration
	}

	// connectionParams 连接所需的配置参数集合.
	connectionParams struct {
		// conn 与终端建立的连接
		conn net.Conn
		// handles 信令处理.每一个连接都是独立的map, 不会影响其他的.
		handles map[consts.JT808CommandType]Handler
		// activeRespondHandles 主动下发，自定义处理回复相关的指令
		activeRespondHandles map[consts.JT808CommandType]func(platformMsg *ActiveMessage, terminalMsg *Message) bool
		// 终端事件处理器，用于推送终端上下线、心跳、位置上报等事件给上层业务.
		terminalEvent TerminalEventer
		// filter 是否启用消息过滤,过滤分包情况的事件，默认过滤.
		filter bool
		// timeout 超时相关配置(空闲超时、首次包时间、最后包时间等).
		timeout TerminalTimeout
		// onTerminalTimeoutEvent 空闲超时事件，设置IdleTimeout时触发.
		onTerminalTimeoutEvent func(timeout TerminalTimeout)
		// 终端上线时的回调函数.
		onJoinEvent func(message *Message, activeChan chan<- *ActiveMessage) (string, error)
		// 终端下线时的回调函数.
		onLeaveEvent func(key string)
	}
)

func newConnection(params connectionParams) *connection {
	return &connection{
		connectionParams: params,

		stopWriteChan:         make(chan struct{}),           // 通知 write 协程进入收尾阶段
		writeDoneChan:         make(chan struct{}),           // 等待 write 协程退出
		terminalUplinkMsgChan: make(chan *Message, 10),       // 上行普通消息队列
		activeMsgChan:         make(chan *ActiveMessage, 10), // 主动下发指令队列
		activeTimeoutChan:     make(chan uint16, 10),         // 主动下发超时通知队列
		reissuePackChan:       make(chan *Message, 10),       // 补发包队列

		// 初始状态
		key: "", // 终端唯一标识，注册成功后由 joinFunc 填充
	}
}

func (c *connection) run() {
	go c.reader()
	go c.write()
}

func (c *connection) reader() {
	var (
		// 复用 1023B 读缓冲（JT808 单包最大长度），热路径零拷贝；详见 doc.go。
		data = make([]byte, 1023)
		pack = newPackageParse()
		join = false
		once sync.Once
	)

	defer func() {
		pack.clear()
		c.stop()
	}()

	for {
		// 设置读取超时（仅当配置了 IdleTimeout 时生效）
		if c.timeout.IdleTimeout > 0 {
			_ = c.conn.SetReadDeadline(time.Now().Add(c.timeout.IdleTimeout))
		}
		// https://pkg.go.dev/io#Reader
		n, err := c.conn.Read(data)
		if n > 0 {
			effectiveData := data[:n]
			msgs, err := pack.parse(effectiveData) // 解析 JT808 协议包（支持分包重组）
			if err != nil {
				slog.Error("parse data",
					slog.Bool("join", join),
					slog.Any("platform num", uint16(c.platformSerialNumber.Load())),
					slog.String("effective data", fmt.Sprintf("%x", effectiveData)),
					slog.Any("err", err))
				return
			}
			if len(msgs) > 0 {
				if !join {
					if err := c.joinHandle(msgs[0]); err == nil {
						join = true
						c.timeout.Key = c.key
					} else if errors.Is(err, _errKeyExist) {
						slog.Warn("key",
							slog.String("effective data", fmt.Sprintf("%x", effectiveData)),
							slog.Any("err", err))
						return
					}
				}

				once.Do(func() {
					c.timeout.FirstPacketTime = time.Now()
				})
				c.timeout.LastPacketTime = time.Now()
				c.handleMessages(msgs)
			}
		}

		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) { // 读取超时,仅当设置 idleTimeout 时生效
				if c.onTerminalTimeoutEvent != nil {
					c.onTerminalTimeoutEvent(c.timeout)
				}
			} else if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				slog.Debug("connection close",
					slog.Bool("join", join),
					slog.Any("platform num", uint16(c.platformSerialNumber.Load())),
					slog.Any("err", err))
			} else {
				slog.Error("read data",
					slog.Bool("join", join),
					slog.Any("platform num", uint16(c.platformSerialNumber.Load())),
					slog.Any("err", err))
			}
			return
		}
	}
}

func (c *connection) joinHandle(msg *Message) error {
	// 首包深拷贝独立 Header（脱离 reader 复用缓冲），供 sessionManager 与后续下行回复复用
	independentHeader := cloneHeader(msg.JTMessage.Header)
	if frame := bytes.Clone(msg.ExtensionFields.TerminalData); len(frame) > 0 {
		jtMsg := jt808.NewJTMessage()
		if err := jtMsg.Decode(frame); err == nil {
			independentHeader = jtMsg.Header
		}
	}
	msg.JTMessage.Header = independentHeader
	key, err := c.onJoinEvent(msg, c.activeMsgChan)
	if err == nil {
		c.key = key
		c.sessionHeader.Store(independentHeader)
	}

	c.terminalEvent.OnJoinEvent(msg, key, err)
	return err
}

func cloneHeader(src *jt808.Header) *jt808.Header {
	h := *src
	if h.Property != nil {
		prop := *h.Property
		h.Property = &prop
	}
	return &h
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
		c.copyOnDemandForWrite(msg)
		c.terminalUplinkMsgChan <- msg
	}
}

// copyOnDemandForWrite 仅对低频、且会在 write() 内读取 Body/TerminalData 的消息按需拷贝；
// 高频上行（0x0200/0x0002/0x0704 等）保持零拷贝。
func (c *connection) copyOnDemandForWrite(msg *Message) {
	if msg.ExtensionFields.SubcontractComplete {
		return
	}
	if c.activePending.Load() > 0 {
		if _, ok := c.activeRespondHandles[msg.Command]; ok {
			msg.Body = bytes.Clone(msg.Body)
			msg.ExtensionFields.TerminalData = bytes.Clone(msg.ExtensionFields.TerminalData)
			return
		}
	}
	switch msg.Command {
	case consts.T0102RegisterAuth,
		consts.T0801MultimediaDataUpload,
		consts.T1212FileUploadComplete:
		msg.Body = bytes.Clone(msg.Body)
	}
}

// write 是连接的核心处理协程，负责统一处理主动下发、超时、分包补传与默认回复.
func (c *connection) write() {
	record := map[uint16]*activeRecord{}
	defer func() {
		for _, entry := range record {
			entry.timer.Stop()
		}
		clear(record)
		clear(c.handles)
		close(c.writeDoneChan)
	}()

	for {
		select {
		case <-c.stopWriteChan:
			c.shutdownWrite(record)
			return

		case activeMsg := <-c.activeMsgChan: // 平台主动下发的
			c.onActiveSendEvent(activeMsg, record)

		case seq := <-c.activeTimeoutChan: // 平台主动下发超时
			c.onActiveTimeoutEvent(seq, record)

		case subPackMsg := <-c.reissuePackChan: // 分包补传的
			c.onReissueSubcontractingEvent(subPackMsg)

		case msg := <-c.terminalUplinkMsgChan: // 终端上传的
			if len(record) > 0 && msg.hasComplete() && c.onActiveRespondEvent(record, msg) {
				continue
			}
			if msg.hasComplete() || !c.filter { // 默认完整包才触发回复
				c.defaultReplyEvent(msg)
			}
		}
	}
}

// shutdownWrite 执行断线收尾：此时 reader 已退出且 sessionManager 已移除该会话，
// 不会再有新数据写入 terminalUplinkMsgChan 和 activeMsgChan.
func (c *connection) shutdownWrite(record map[uint16]*activeRecord) {
	// 1. 先匹配断线前已入队的终端应答，避免误判超时
	for len(c.terminalUplinkMsgChan) > 0 {
		msg := <-c.terminalUplinkMsgChan
		if len(record) > 0 && msg.hasComplete() {
			c.onActiveRespondEvent(record, msg)
		}
	}
	// 2. 排空未发出的主动请求（conn 已关闭，Write 失败会直接返回 ErrWriteDataFail）
	for len(c.activeMsgChan) > 0 {
		c.onActiveSendEvent(<-c.activeMsgChan, record)
	}
	// 3. 方案 B：等待断线前已发出的在途主动请求依次超时完成
	for len(record) > 0 {
		c.onActiveTimeoutEvent(<-c.activeTimeoutChan, record)
	}
}

// stop 执行连接关闭流程：下线通知、关闭连接、通知 write 收尾并等待退出.
func (c *connection) stop() {
	c.onLeaveEvent(c.key)
	c.terminalEvent.OnLeaveEvent(c.key)
	if err := c.conn.Close(); err != nil {
		slog.Warn("conn close fail",
			slog.String("key", c.key),
			slog.Any("err", err))
	}
	close(c.stopWriteChan)
	<-c.writeDoneChan
}

func (c *connection) replyHeader(msgHeader *jt808.Header) *jt808.Header {
	if base := c.sessionHeader.Load(); base != nil {
		header := cloneHeader(base)
		header.ID = msgHeader.ID
		header.SerialNumber = msgHeader.SerialNumber
		return header
	}
	return cloneHeader(msgHeader)
}

func (c *connection) defaultReplyEvent(msg *Message) {
	if has := msg.HasReply(); !has {
		return
	}
	body, err := msg.ReplyBody(msg.JTMessage)
	if err != nil {
		slog.Warn("reply body fail",
			slog.String("key", c.key),
			slog.Any("command", msg.Command),
			slog.Any("seq", msg.ExtensionFields.TerminalSeq),
			slog.Any("err", err))
		return
	}
	header := c.replyHeader(msg.JTMessage.Header)
	header.ReplyID = uint16(msg.ReplyProtocol())
	platformSeq, _ := c.allocSeq(0)
	header.PlatformSerialNumber = platformSeq

	packets := header.EncodePackets(body)
	for _, data := range packets {
		if _, err = c.conn.Write(data); err != nil {
			slog.Warn("write fail",
				slog.String("data", fmt.Sprintf("%x", data)),
				slog.Any("err", err))
			msg.ExtensionFields.Err = errors.Join(ErrWriteDataFail, err)
		}
		msg.ExtensionFields.PlatformData = append(msg.ExtensionFields.PlatformData, data...)
	}
	if _, seq := c.allocSeq(len(packets)); len(packets) > 1 {
		platformSeq = seq - 1
	}

	msg.ExtensionFields.PlatformCommand = msg.ReplyProtocol()
	msg.ExtensionFields.PlatformSeq = platformSeq

	c.onWriteExecutionEvent(msg)
}

func (c *connection) onReissueSubcontractingEvent(msg *Message) {
	header := c.replyHeader(msg.JTMessage.Header)
	original, _ := c.allocSeq(1)
	header.PlatformSerialNumber = original
	header.ReplyID = uint16(consts.P8003ReissueSubcontractingRequest)
	packets := header.EncodePackets(msg.JTMessage.Body)

	if len(packets) != 1 { // 分包补传的报文固定大小，不会分包
		slog.Warn("onReissueSubcontractingEvent",
			slog.String("key", c.key),
			slog.Int("packets", len(packets)),
			slog.Any("data", fmt.Sprintf("%x", msg.JTMessage.Body)))
		return
	}

	data := packets[0]
	if _, err := c.conn.Write(data); err != nil {
		slog.Warn("write fail",
			slog.String("data", fmt.Sprintf("%x", data)),
			slog.Any("err", err))
		msg.ExtensionFields.Err = errors.Join(ErrWriteDataFail, err)
	}

	msg.ExtensionFields.PlatformSeq = original
	msg.ExtensionFields.PlatformData = data
	c.onWriteExecutionEvent(msg)
}

func (c *connection) onActiveSendEvent(activeMsg *ActiveMessage, record map[uint16]*activeRecord) {
	// 必须在 conn.Write 之前递增 activePending，保证 happens-before 终端回包到达 reader()
	c.activePending.Add(1)
	jtMsg, data, err := c.activeWritePackets(activeMsg)
	platformSeq := jtMsg.Header.SerialNumber
	replyMsg := newActiveSendMessage(jtMsg, activeMsg.Command, func(message *Message) {
		message.ExtensionFields.PlatformData = data
		message.ExtensionFields.Err = err
	})
	if v, ok := c.handles[activeMsg.Command]; ok {
		replyMsg.Handler = v
		c.onReadExecutionEvent(replyMsg)
	} else {
		c.terminalEvent.OnNotSupportedEvent(replyMsg)
	}
	activeMsg.ExtensionFields = struct {
		PlatformSeq uint16 `json:"platformSeq,omitempty"`
		Data        []byte `json:"data,omitempty"`
	}{
		PlatformSeq: platformSeq,
		Data:        data,
	}

	activeMsg.convertMessage = replyMsg
	if err != nil {
		c.onActiveEventComplete(replyMsg, activeMsg)
		return
	}

	duration := 3 * time.Second
	if activeMsg.OverTimeDuration > 0 {
		duration = activeMsg.OverTimeDuration
	}
	if old, ok := record[platformSeq]; ok {
		old.timer.Stop()
		c.activePending.Add(-1)
	}
	timer := time.AfterFunc(duration, func() {
		select {
		case <-c.writeDoneChan:
		case c.activeTimeoutChan <- platformSeq:
		}
	})
	record[platformSeq] = &activeRecord{
		msg:     activeMsg,
		timer:   timer,
		timeout: duration,
	}
}

func (c *connection) onActiveTimeoutEvent(seq uint16, record map[uint16]*activeRecord) {
	entry, ok := record[seq]
	if !ok {
		return
	}
	delete(record, seq)
	overtimeMsg := entry.msg.convertMessage
	overtimeMsg.ExtensionFields.Err = errors.Join(ErrWriteDataOverTime,
		fmt.Errorf("overtime is [%.2f]second", entry.timeout.Seconds()))
	c.onActiveEventComplete(overtimeMsg, entry.msg)
}

func (c *connection) activeWritePackets(activeMsg *ActiveMessage) (*jt808.JTMessage, []byte, error) {
	var (
		platformData = make([]byte, 0, len(activeMsg.Body)+25)
		writeErr     error
		jtMsg        = jt808.NewJTMessage()
	)
	header := cloneHeader(activeMsg.header)
	platformSeq, _ := c.allocSeq(0)
	header.PlatformSerialNumber = platformSeq
	header.ReplyID = uint16(activeMsg.Command)
	packets := header.EncodePackets(activeMsg.Body)

	for _, data := range packets {
		if _, err := c.conn.Write(data); err != nil {
			writeErr = errors.Join(ErrWriteDataFail, err)
		}
		platformData = append(platformData, data...)
	}

	if len(packets) > 0 {
		_ = jtMsg.Decode(packets[0])
		if len(packets) > 1 {
			jtMsg.Body = activeMsg.Body
		}
	}
	_, _ = c.allocSeq(len(packets))
	return jtMsg, platformData, writeErr
}

func (c *connection) onActiveRespondEvent(record map[uint16]*activeRecord, terminalMsg *Message) bool {
	matchFunc, ok := c.activeRespondHandles[terminalMsg.Command]
	if !ok {
		return false
	}
	for seq, entry := range record {
		if matchFunc(entry.msg, terminalMsg) {
			entry.timer.Stop()
			delete(record, seq)
			terminalMsg.ExtensionFields.PlatformSeq = seq
			terminalMsg.ExtensionFields.TerminalCommand = terminalMsg.Protocol()
			c.onActiveEventComplete(terminalMsg, entry.msg)
			return true
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
	c.activePending.Add(-1)
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

func (c *connection) allocSeq(n int) (uint16, uint16) {
	original := uint16(c.platformSerialNumber.Load())
	if n == 0 {
		return original, original
	}
	next := original + uint16(n)
	c.platformSerialNumber.Store(uint32(next))
	return original, next
}
