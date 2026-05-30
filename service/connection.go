package service

import (
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

		// finallyCompleteChan 用于等待当前所有正在主动下发的指令完成或超时后才真正关闭连接.
		finallyCompleteChan chan struct{}
		// terminalUplinkMsgChan 终端上报的普通消息（上行消息）会放入此通道，
		// 随后由 write 协程统一处理回复逻辑.
		terminalUplinkMsgChan chan *Message
		// reissuePackChan 需要补发的分包消息通道.
		reissuePackChan chan *Message
		// platformSerialNumber 平台侧流水号（0~65535），自增后循环使用.
		platformSerialNumber uint16

		// activeMsgChan 平台主动下发给终端的指令会包装成 ActiveMessage 放入此通道.
		activeMsgChan chan *ActiveMessage
		// activeMsgCompleteChan 当平台主动下发的指令收到终端应答（或超时）后，
		// 会把完整的 Message 放入此通道，用于通知等待应答的协程.
		activeMsgCompleteChan chan *Message
		// activeUnfinishedSum 当前仍在等待终端应答的主动下发指令数量（原子操作）.
		activeUnfinishedSum int32
		// key 当前连接对应的终端唯一标识（默认是 SIM 卡号），
		// 由 onJoinEvent 回调函数返回后赋值.
		key string
	}

	// connectionParams 连接所需的配置参数集合.
	connectionParams struct {
		// conn 与终端建立的 TCP 连接
		conn *net.TCPConn
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

		finallyCompleteChan:   make(chan struct{}),           // 等待所有主动下发完成
		terminalUplinkMsgChan: make(chan *Message, 10),       // 上行普通消息队列
		activeMsgChan:         make(chan *ActiveMessage, 10), // 主动下发指令队列
		activeMsgCompleteChan: make(chan *Message, 10),       // 主动下发应答通知队列
		reissuePackChan:       make(chan *Message, 10),       // 补发包队列

		// 初始状态
		platformSerialNumber: 0,        // 平台流水号从0开始
		key:                  "",       // 终端唯一标识，注册成功后由 joinFunc 填充
		activeUnfinishedSum:  int32(0), // 当前未完成的下发指令计数
	}
}

func (c *connection) run() {
	go c.reader()
	go c.write()
}

func (c *connection) reader() {
	var (
		// 消息体长度最大为 10bit 也就是 1023 的字节
		// 实际上reader相当于同步执行的 因此无需拷贝一份新的
		// 1. onReadExecutionEvent 这里处理后在read读取因此不会污染 如果onRead事件需要的话 用户自行拷贝一份
		// 2. 历史的图片数据是make一份新的 因此一样不会有污染
		// 3. write异步的情况 即使污染了也不影响 因为实际用不到（仅需要已经序列化的头部)
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
					slog.Any("platform num", c.platformSerialNumber),
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
					slog.Any("platform num", c.platformSerialNumber),
					slog.Any("err", err))
			} else {
				slog.Error("read data",
					slog.Bool("join", join),
					slog.Any("platform num", c.platformSerialNumber),
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
		c.terminalUplinkMsgChan <- msg
	}
}

// write 是连接的核心处理协程，负责统一处理以下事件：
//  1. 平台主动下发指令
//  2. 主动下发指令的应答/超时
//  3. 分包补传
//  4. 终端上报消息的默认回复.
func (c *connection) write() {
	record := map[uint16]*ActiveMessage{}
	defer func() {
		close(c.activeMsgChan)
		close(c.activeMsgCompleteChan)
		close(c.reissuePackChan)
		close(c.terminalUplinkMsgChan)
		clear(record)
		clear(c.handles)
	}()

	for {
		select {
		case <-c.finallyCompleteChan: // 如果现在有平台主动下发的, 需要等待完成在退出
			return
		case activeMsg := <-c.activeMsgChan: // 平台主动下发的
			atomic.AddInt32(&c.activeUnfinishedSum, 1)
			c.onActiveSendEvent(activeMsg, record)

		case msg := <-c.activeMsgCompleteChan: // 平台主动下发的完成情况
			seq := msg.ExtensionFields.PlatformSeq
			// 超时的情况一定执行一次, 如果完成了,还可能在执行一次超时的回调
			if v, ok := record[seq]; ok {
				c.onActiveEventComplete(msg, v)
				delete(record, seq)
				atomic.AddInt32(&c.activeUnfinishedSum, -1)
			}

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

// stop 执行连接关闭流程：下线通知、关闭 TCP 连接、等待未完成指令.
func (c *connection) stop() {
	c.onLeaveEvent(c.key)
	c.terminalEvent.OnLeaveEvent(c.key)
	if err := c.conn.Close(); err != nil {
		slog.Warn("conn close fail",
			slog.String("key", c.key),
			slog.Any("err", err))
	}
	time.Sleep(time.Second)
	for atomic.LoadInt32(&c.activeUnfinishedSum) != 0 {
		time.Sleep(time.Second)
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
			slog.String("terminal data", fmt.Sprintf("%x", msg.ExtensionFields.TerminalData)),
			slog.Any("err", err))
		return
	}
	header := msg.JTMessage.Header
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

func (c *connection) onActiveSendEvent(activeMsg *ActiveMessage, record map[uint16]*ActiveMessage) {
	header := activeMsg.header
	platformSeq, _ := c.allocSeq(0)
	header.PlatformSerialNumber = platformSeq
	header.ReplyID = uint16(activeMsg.Command)
	packets := header.EncodePackets(activeMsg.Body)

	var (
		platformData = make([]byte, 0, len(activeMsg.Body)+10)
		writeErr     error
		jtMsg        = jt808.NewJTMessage()
	)

	for _, data := range packets {
		if _, err := c.conn.Write(data); err != nil {
			writeErr = errors.Join(ErrWriteDataFail, err)
		}
		_ = jtMsg.Decode(data)
		platformData = append(platformData, jtMsg.Body...)
	}
	if len(packets) > 0 {
		_ = jtMsg.Decode(platformData)
	}
	if _, seq := c.allocSeq(len(packets)); len(packets) > 1 {
		// 如果分包了，使用分包最后的一个流水号做标记, 当前终端回复使用的是分包的最后流水号的包
		platformSeq = seq - 1
	}

	replyMsg := newActiveSendMessage(jtMsg, activeMsg.Command, func(message *Message) {
		message.ExtensionFields.PlatformData = platformData
		message.ExtensionFields.Err = writeErr
		message.ExtensionFields.PlatformSeq = platformSeq
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
		Data:        platformData,
	}

	activeMsg.convertMessage = replyMsg
	record[platformSeq] = activeMsg
	if writeErr != nil {
		c.activeMsgCompleteChan <- replyMsg
	} else {
		duration := 3 * time.Second
		if activeMsg.OverTimeDuration > 0 {
			duration = activeMsg.OverTimeDuration
		}
		go func(overtimeMsg *Message, timeout time.Duration) {
			time.Sleep(timeout)
			select {
			case <-c.finallyCompleteChan:
				return
			default:
			}
			overtimeMsg.ExtensionFields.Err = errors.Join(ErrWriteDataOverTime,
				fmt.Errorf("overtime is [%.2f]second", timeout.Seconds()))
			c.activeMsgCompleteChan <- overtimeMsg
		}(replyMsg, duration)
	}
}

func (c *connection) onActiveRespondEvent(record map[uint16]*ActiveMessage, terminalMsg *Message) bool {
	matchFunc, ok := c.activeRespondHandles[terminalMsg.Command]
	if ok {
		for seq, platformMessage := range record {
			if matchFunc(platformMessage, terminalMsg) {
				terminalMsg.ExtensionFields.PlatformSeq = seq
				terminalMsg.ExtensionFields.TerminalCommand = terminalMsg.Protocol()
				c.activeMsgCompleteChan <- terminalMsg
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

func (c *connection) allocSeq(n int) (uint16, uint16) {
	original := c.platformSerialNumber
	c.platformSerialNumber += uint16(n)
	return original, c.platformSerialNumber
}
