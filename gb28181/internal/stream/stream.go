package stream

import (
	"github.com/cuteLittleDevil/go-jt808/gb28181/command"
	"sync"
)

type Manage struct {
	stopOnce      sync.Once
	stopChan      chan struct{}
	inviteChan    chan *command.InviteInfo
	ackChan       chan string
	byeChan       chan string
	OnInviteEvent func(*command.InviteInfo) *command.InviteInfo
}

func NewManage(onInviteEvent func(*command.InviteInfo) *command.InviteInfo) *Manage {
	return &Manage{
		OnInviteEvent: onInviteEvent,
		stopChan:      make(chan struct{}),
		inviteChan:    make(chan *command.InviteInfo, 10),
		ackChan:       make(chan string, 10),
		byeChan:       make(chan string, 10),
	}
}

func (s *Manage) Run() {
	var (
		record  = map[string]*command.InviteInfo{}
		servers = map[string]*jt1078Server{}
	)
	defer func() {
		clear(record)
		for _, v := range servers {
			v.stop()
		}
		clear(servers)
	}()
	for {
		select {
		case <-s.stopChan:
			return
		case v := <-s.inviteChan:
			if old, ok := servers[v.CallId]; ok {
				old.stop()
				delete(servers, v.CallId)
			}
			record[v.CallId] = v
		case callID := <-s.ackChan:
			if inviteInfo, ok := record[callID]; ok {
				// 触发回调 让jt808设备上传jt1078到端口A
				if s.OnInviteEvent != nil {
					inviteInfo = s.OnInviteEvent(inviteInfo)
				}
				// 监听端口A 有数据时 建立TCP连接 连接到端口B（收流端口)
				// 把端口A收到的jt1078数据 转ps流数据上传
				server := newJt1078Server(inviteInfo)
				go server.run()
				servers[callID] = server
			}
		case callID := <-s.byeChan:
			delete(record, callID)
			if v, ok := servers[callID]; ok {
				v.stop()
			}
			delete(servers, callID)
		}
	}
}

func (s *Manage) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

func (s *Manage) SubmitInvite(inviteInfo *command.InviteInfo) {
	select {
	case <-s.stopChan:
		return
	default:
		s.inviteChan <- inviteInfo
	}
}

func (s *Manage) SubmitAck(callID string) {
	select {
	case <-s.stopChan:
		return
	default:
		s.ackChan <- callID
	}
}

func (s *Manage) SubmitBye(callID string) {
	select {
	case <-s.stopChan:
		return
	default:
		s.byeChan <- callID
	}
}
