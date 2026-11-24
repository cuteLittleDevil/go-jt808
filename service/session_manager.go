package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
)

type sessionOperationFunc func(record map[string]*session)

type (
	sessionManager struct {
		operationFuncChan chan sessionOperationFunc
		keyFunc           func(message *Message) (string, bool)
	}

	session struct {
		header *jt808.Header
		// 加入时间
		joinTime time.Time
		// 数据发送到终端
		activeMsgChan chan<- *ActiveMessage
	}
)

func newSessionManager(keyFunc func(message *Message) (string, bool)) *sessionManager {
	return &sessionManager{
		operationFuncChan: make(chan sessionOperationFunc, 10),
		keyFunc:           keyFunc,
	}
}

func (s *sessionManager) run() {
	record := make(map[string]*session, 1000)
	for opFunc := range s.operationFuncChan {
		opFunc(record)
	}
}

// join 将终端会话注册到会话管理器中.
func (s *sessionManager) join(message *Message, activeChan chan<- *ActiveMessage) (string, error) {
	key, ok := s.keyFunc(message)
	if !ok {
		return "", _errKeyInvalid
	}
	ch := make(chan error)
	defer close(ch)
	s.operationFuncChan <- func(record map[string]*session) {
		if v, ok := record[key]; ok {
			ch <- errors.Join(fmt.Errorf("key[%s] join time[%s]",
				key, v.joinTime.Format(time.RFC3339)), _errKeyExist)
			return
		}
		record[key] = &session{
			header:        message.Header,
			joinTime:      time.Now(),
			activeMsgChan: activeChan,
		}
		ch <- nil
	}
	return key, <-ch
}

// leave 将指定 key 的终端会话从会话管理器中移除.
func (s *sessionManager) leave(key string) {
	ch := make(chan struct{})
	s.operationFuncChan <- func(record map[string]*session) {
		defer close(ch)
		delete(record, key)
	}
	<-ch
}

// write 将主动消息分配到目标终端会话并返回应答结果.
// 会话表只在 run() 中访问。write 必须在本闭包内完成
// activeMsgChan 发送，以便与 leave 串行，避免发到已关闭通道。
// 缓冲 10 是余量：业务上 SendActiveMessage 一问一答，
// 同一连接同时通常最多一条在途，不会靠堆积连续下发。
func (s *sessionManager) write(activeMsg *ActiveMessage) *Message {
	replyChan := make(chan *Message)
	defer close(replyChan)
	s.operationFuncChan <- func(record map[string]*session) {
		key := activeMsg.Key
		if v, ok := record[key]; ok {
			activeMsg.header = v.header
			activeMsg.replyChan = replyChan
			v.activeMsgChan <- activeMsg
			return
		}
		replyChan <- newErrMessage(errors.Join(ErrNotExistKey,
			fmt.Errorf("key=[%s] sum=[%d] ", key, len(record))))
	}
	return <-replyChan
}
