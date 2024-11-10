package service

import (
	"errors"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"time"
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
	for {
		select {
		case opFunc := <-s.operationFuncChan:
			opFunc(record)
		}
	}
}

func (s *sessionManager) join(message *Message, activeChan chan<- *ActiveMessage) (string, error) {
	key, ok := s.keyFunc(message)
	if !ok {
		return "", _errKeyInvalid
	}
	ch := make(chan error)
	defer close(ch)
	s.operationFuncChan <- func(record map[string]*session) {
		if v, ok := record[key]; ok {
			ch <- errors.Join(fmt.Errorf("exist key join time[%s]",
				v.joinTime.Format(time.RFC3339)), _errKeyExist)
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

func (s *sessionManager) leave(key string) {
	ch := make(chan struct{})
	s.operationFuncChan <- func(record map[string]*session) {
		defer close(ch)
		if _, ok := record[key]; ok {
			delete(record, key)
		}
	}
	<-ch
	return
}

func (s *sessionManager) write(activeMsg *ActiveMessage) *Message {
	replyChan := make(chan *Message)
	defer close(replyChan)
	s.operationFuncChan <- func(record map[string]*session) {
		key := activeMsg.Key
		if v, ok := record[key]; ok {
			activeMsg.header = v.header
			activeMsg.completeChan = make(chan struct{})
			activeMsg.replyChan = replyChan
			v.activeMsgChan <- activeMsg
			return
		}
		replyChan <- newErrMessage(errors.Join(ErrNotExistKey,
			fmt.Errorf("key=[%s] sum=[%d] ", key, len(record))))
	}
	return <-replyChan
}
