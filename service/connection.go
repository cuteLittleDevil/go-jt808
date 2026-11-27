package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

const (
	defaultActiveTimeout = 3 * time.Second
	// defaultWriteTimeout 单次写超时 超时说明终端网络异常 直接断开连接
	defaultWriteTimeout = 10 * time.Second
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

		// finallyCompleteChan 连接关闭时通知 write 协程排空并等待在途主动消息完成.
		finallyCompleteChan chan struct{}
		// terminalUplinkMsgChan 终端上报的普通消息（上行消息）会放入此通道，
		// 随后由 write 协程统一处理回复逻辑.
		terminalUplinkMsgChan chan *Message
		// reissuePackChan 需要补发的分包消息通道.
		reissuePackChan chan *Message
		// platformSerialNumber 平台侧流水号（低 16 位 0~65535 循环使用）.
		platformSerialNumber atomic.Uint32

		// activeMsgChan 平台主动下发给终端的指令会包装成 ActiveMessage 放入此通道.
		activeMsgChan chan *ActiveMessage
		// activeTimeoutChan 主动下发超时通知通道（不主动关闭，避免关闭竞态 panic）.
		activeTimeoutChan chan *activeRecord
		// key 当前连接对应的终端唯一标识（默认是 SIM 卡号），
		// 由 onJoinEvent 回调函数返回后赋值.
		key string
	}

	activeRecord struct {
		seq     uint16
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
		// maxFrameBufferSize 粘包缓冲区最大字节数 达到后断开连接 <=0不限制.
		maxFrameBufferSize int
		// writeTimeout 单次写超时 <=0时使用默认defaultWriteTimeout.
		writeTimeout time.Duration
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
	if params.writeTimeout <= 0 {
		params.writeTimeout = defaultWriteTimeout
	}
	return &connection{
		connectionParams: params,

		finallyCompleteChan:   make(chan struct{}),           // 通知 write 协程进入关闭收尾
		terminalUplinkMsgChan: make(chan *Message, 10),       // 上行普通消息队列
		activeMsgChan:         make(chan *ActiveMessage, 10), // 主动下发指令队列
		activeTimeoutChan:     make(chan *activeRecord, 10),  // 主动下发超时通知队列
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
		// 复用 1023B 读缓冲（JT808 单包最大长度），热路径零拷贝。
		data = make([]byte, 1023)
		pack = newPackageParse()
		join = false
		once sync.Once
	)

	defer func() {
		pack.clear()
		c.stop(join)
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
					if err := c.joinHandle(msgs[0]); err != nil {
						slog.Warn("key",
							slog.String("effective data", fmt.Sprintf("%x", effectiveData)),
							slog.Any("err", err))
						return
					}
					join = true
					c.timeout.Key = c.key
				}

				once.Do(func() {
					c.timeout.FirstPacketTime = time.Now()
				})
				c.timeout.LastPacketTime = time.Now()
				c.handleMessages(msgs)
			}
			// 粘包缓冲区未组成完整帧的数据过多 说明数据异常 断开连接防止内存无限增长
			if pending := pack.frameReader.Pending(); c.maxFrameBufferSize > 0 && pending >= c.maxFrameBufferSize {
				slog.Warn("frame buffer overflow",
					slog.String("key", c.key),
					slog.Int("pending", pending),
					slog.Int("max", c.maxFrameBufferSize))
				return
			}
		}

		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) { // 读取超时,仅当设置 idleTimeout 时生效
				if c.onTerminalTimeoutEvent != nil {
					c.onTerminalTimeoutEvent(c.timeout)
				}
			} else if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
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
	key, err := c.onJoinEvent(msg, c.activeMsgChan)
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
		if !msg.ExtensionFields.SubcontractComplete && !isHighFreqCommand(msg.Command) {
			msg.Body = bytes.Clone(msg.Body)
			msg.ExtensionFields.TerminalData = bytes.Clone(msg.ExtensionFields.TerminalData)
		}
		c.terminalUplinkMsgChan <- msg
	}
}

func isHighFreqCommand(cmd consts.JT808CommandType) bool {
	switch cmd {
	case consts.T0002HeartBeat,
		consts.T0200LocationReport,
		consts.T0704LocationBatchUpload:
		return true
	}
	return false
}

// write 是连接的核心处理协程，负责统一处理以下事件：
//  1. 平台主动下发指令
//  2. 主动下发指令的应答/超时
//  3. 分包补传
//  4. 终端上报消息的默认回复.
func (c *connection) write() {
	record := map[uint16]*activeRecord{}
	done := make(chan struct{})
	defer func() {
		close(done)
		for _, v := range record {
			v.timer.Stop()
		}
		clear(record)
		clear(c.handles)
	}()

	for {
		select {
		case <-c.finallyCompleteChan: // 连接关闭后排空队列并等待在途主动消息完成
			c.drainBeforeExit(record, done)
			return
		case activeMsg := <-c.activeMsgChan: // 平台主动下发的
			c.onActiveSendEvent(activeMsg, record, done)

		case rec := <-c.activeTimeoutChan: // 平台主动下发超时
			c.onActiveTimeoutEvent(rec, record)

		case subPackMsg := <-c.reissuePackChan: // 分包补传的
			c.onReissueSubcontractingEvent(subPackMsg)

		case msg := <-c.terminalUplinkMsgChan: // 终端上传的
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

func (c *connection) drainBeforeExit(record map[uint16]*activeRecord, done <-chan struct{}) {
	for len(c.terminalUplinkMsgChan) > 0 {
		msg := <-c.terminalUplinkMsgChan
		if len(record) > 0 && msg.hasComplete() {
			c.onActiveRespondEvent(record, msg)
		}
	}
	for len(c.activeMsgChan) > 0 {
		c.onActiveSendEvent(<-c.activeMsgChan, record, done)
	}
	for len(record) > 0 {
		c.onActiveTimeoutEvent(<-c.activeTimeoutChan, record)
	}
}

// stop 执行连接关闭流程：下线通知、关闭 TCP 连接、通知写协程收尾.
func (c *connection) stop(join bool) {
	if join {
		c.onLeaveEvent(c.key)
		c.terminalEvent.OnLeaveEvent(c.key)
	}
	if err := c.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		slog.Warn("conn close fail",
			slog.String("key", c.key),
			slog.Any("err", err))
	}
	close(c.finallyCompleteChan)
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
	header := msg.JTMessage.Header
	header.ReplyID = uint16(msg.ReplyProtocol())
	platformSeq, _ := c.allocSeq(0)
	header.PlatformSerialNumber = platformSeq

	packets := header.EncodePackets(body)
	for _, data := range packets {
		if err = c.writeData(data); err != nil {
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
	header := msg.JTMessage.Header
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
	if err := c.writeData(data); err != nil {
		slog.Warn("write fail",
			slog.String("data", fmt.Sprintf("%x", data)),
			slog.Any("err", err))
		msg.ExtensionFields.Err = errors.Join(ErrWriteDataFail, err)
	}

	msg.ExtensionFields.PlatformSeq = original
	msg.ExtensionFields.PlatformData = data
	c.onWriteExecutionEvent(msg)
}

func (c *connection) onActiveSendEvent(activeMsg *ActiveMessage, record map[uint16]*activeRecord, done <-chan struct{}) {
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
	duration := defaultActiveTimeout
	if activeMsg.OverTimeDuration > 0 {
		duration = activeMsg.OverTimeDuration
	}
	if old, ok := record[platformSeq]; ok {
		old.timer.Stop()
		c.onActiveTimeoutEvent(old, record)
	}
	rec := &activeRecord{
		seq:     platformSeq,
		msg:     activeMsg,
		timeout: duration,
	}
	rec.timer = time.AfterFunc(duration, func() {
		select {
		case <-done:
		case c.activeTimeoutChan <- rec:
		}
	})
	record[platformSeq] = rec
}

func (c *connection) onActiveTimeoutEvent(rec *activeRecord, record map[uint16]*activeRecord) {
	if record[rec.seq] != rec {
		return
	}
	delete(record, rec.seq)
	overtimeMsg := rec.msg.convertMessage
	overtimeMsg.ExtensionFields.Err = errors.Join(ErrWriteDataOverTime,
		fmt.Errorf("overtime is [%.2f]second", rec.timeout.Seconds()))
	c.onActiveEventComplete(overtimeMsg, rec.msg)
}

func (c *connection) activeWritePackets(activeMsg *ActiveMessage) (*jt808.JTMessage, []byte, error) {
	var (
		platformData = make([]byte, 0, len(activeMsg.Body)+10)
		writeErr     error
		jtMsg        = jt808.NewJTMessage()
	)
	header := activeMsg.header
	platformSeq, _ := c.allocSeq(0)
	header.PlatformSerialNumber = platformSeq
	header.ReplyID = uint16(activeMsg.Command)
	packets := header.EncodePackets(activeMsg.Body)

	for _, data := range packets {
		if err := c.writeData(data); err != nil {
			writeErr = errors.Join(ErrWriteDataFail, err)
		}
		platformData = append(platformData, data...)
	}

	if len(packets) > 0 {
		_ = jtMsg.Decode(packets[len(packets)-1])
		if len(packets) > 1 {
			jtMsg.Body = activeMsg.Body
		}
	}
	_, _ = c.allocSeq(len(packets))
	return jtMsg, platformData, writeErr
}

func (c *connection) onActiveRespondEvent(record map[uint16]*activeRecord, terminalMsg *Message) bool {
	matchFunc, ok := c.activeRespondHandles[terminalMsg.Command]
	if ok {
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

// writeData 带超时的写 超时说明终端网络异常 关闭连接让读协程退出走正常收尾.
func (c *connection) writeData(data []byte) error {
	_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	_, err := c.conn.Write(data)
	if errors.Is(err, os.ErrDeadlineExceeded) {
		slog.Warn("write timeout, close connection",
			slog.String("key", c.key),
			slog.Any("err", err))
		_ = c.conn.Close()
	}
	return err
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
