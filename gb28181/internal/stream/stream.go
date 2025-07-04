package stream

import (
	"gb28181/command"
	"sync"
)

type Stream struct {
	stopOnce      sync.Once
	stopChan      chan struct{}
	inviteChan    chan *command.InviteInfo
	ackChan       chan string
	byeChan       chan string
	OnInviteEvent func(command.InviteInfo)
}

func NewStream(onInviteEvent func(command.InviteInfo)) *Stream {
	return &Stream{
		OnInviteEvent: onInviteEvent,
		stopChan:      make(chan struct{}),
		inviteChan:    make(chan *command.InviteInfo, 10),
		ackChan:       make(chan string, 10),
		byeChan:       make(chan string, 10),
	}
}

func (s *Stream) Run() {
	var (
		record   = map[string]*command.InviteInfo{}
		managers = map[string]*jt1078Server{}
	)
	defer func() {
		clear(record)
		for _, v := range managers {
			v.stop()
		}
		clear(managers)
	}()
	for {
		select {
		case <-s.stopChan:
			return
		case v := <-s.inviteChan:
			if old, ok := managers[v.CallId]; ok {
				old.stop()
				delete(managers, v.CallId)
			}
			record[v.CallId] = v
		case callID := <-s.ackChan:
			if inviteInfo, ok := record[callID]; ok {
				// 监听端口A 有数据时 建立TCP连接 连接到端口B（收流端口)
				// 把端口A收到的jt1078数据 转ps流数据上传
				server := newJt1078Server(inviteInfo.SSRC)
				go server.run(inviteInfo.JT1078Info.Port, inviteInfo.IP, inviteInfo.Port)
				// 触发回调 让jt808设备上传jt1078到端口A
				if s.OnInviteEvent != nil {
					s.OnInviteEvent(*inviteInfo)
				}
				managers[callID] = server
			}
		case callID := <-s.byeChan:
			delete(record, callID)
			if v, ok := managers[callID]; ok {
				v.stop()
			}
			delete(managers, callID)
		}
	}
}

func (s *Stream) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

func (s *Stream) SubmitInvite(inviteInfo *command.InviteInfo) {
	select {
	case <-s.stopChan:
		return
	default:
		s.inviteChan <- inviteInfo
	}
}

func (s *Stream) SubmitAck(callID string) {
	select {
	case <-s.stopChan:
		return
	default:
		s.ackChan <- callID
	}
}

func (s *Stream) SubmitBye(callID string) {
	select {
	case <-s.stopChan:
		return
	default:
		s.byeChan <- callID
	}
}
