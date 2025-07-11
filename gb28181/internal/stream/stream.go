package stream

import (
	"fmt"
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
	ConvertFunc   func() command.JT1078ToGB28181er
}

func NewManage(onInviteEvent func(*command.InviteInfo) *command.InviteInfo,
	convertFunc func() command.JT1078ToGB28181er) *Manage {
	return &Manage{
		OnInviteEvent: onInviteEvent,
		ConvertFunc:   convertFunc,
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
			v.stop("模拟器退出")
		}
		clear(servers)
	}()
	for {
		select {
		case <-s.stopChan:
			return
		case v := <-s.inviteChan:
			if old, ok := servers[v.CallId]; ok {
				old.stop(fmt.Sprintf("之前的流还存在 callID[%s]", v.CallId))
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
				jt1078ToGB := s.ConvertFunc()
				jt1078ToGB.OnAck(inviteInfo)
				server := newJt1078Server(inviteInfo, jt1078ToGB)
				go server.run()
				servers[callID] = server
			}
		case callID := <-s.byeChan:
			delete(record, callID)
			if v, ok := servers[callID]; ok {
				v.stop("收到bye")
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
