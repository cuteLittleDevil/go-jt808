package pkg

import (
	"fmt"
	"net"
)

type audioOperationFunc func(record map[int]*session)

type (
	AudioManager struct {
		operationFuncChan chan audioOperationFunc
		audioPorts        [2]int
		audios            map[int]*session
	}

	session struct {
		// 是否使用
		use bool
		// 收到音频数据
		audioChan chan []byte
	}
)

func NewAudioManager(audioPorts [2]int) *AudioManager {
	return &AudioManager{
		operationFuncChan: make(chan audioOperationFunc, 10),
		audioPorts:        audioPorts,
	}
}

func (s *AudioManager) Init() error {
	audios := make(map[int]*session, 10)
	for i := s.audioPorts[0]; i <= s.audioPorts[1]; i++ {
		ch := make(chan []byte, 10)
		listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", i))
		if err != nil {
			return err
		}
		go func(readChan chan<- []byte) {
			for {
				conn, err := listen.Accept()
				if err == nil {
					go func() {
						buf := make([]byte, 10*1024)
						defer clear(buf)
						for {
							if n, err := conn.Read(buf); err != nil {
								return
							} else if n > 0 {
								readChan <- buf[:n]
							}
						}
					}()
				}
			}
		}(ch)
		audios[i] = &session{
			use:       false,
			audioChan: ch,
		}
	}
	s.audios = audios
	return nil
}

func (s *AudioManager) Run() {
	for {
		select {
		case opFunc := <-s.operationFuncChan:
			opFunc(s.audios)
		}
	}
}

func (s *AudioManager) allocate() (<-chan []byte, int, error) {
	type Message struct {
		audioPort int
		audioChan chan []byte
		Err       error
	}
	ch := make(chan *Message)
	defer close(ch)
	s.operationFuncChan <- func(record map[int]*session) {
		msg := &Message{
			audioPort: -1,
			audioChan: nil,
			Err:       fmt.Errorf("音频端口都被使用了"),
		}
		defer func() {
			ch <- msg
		}()
		for k, v := range record {
			if !v.use {
				v.use = true
				msg.audioPort = k
				msg.audioChan = v.audioChan
				msg.Err = nil
				return
			}
		}
	}
	msg := <-ch
	return msg.audioChan, msg.audioPort, msg.Err
}

func (s *AudioManager) recycle(audioPort int) {
	ch := make(chan struct{})
	s.operationFuncChan <- func(record map[int]*session) {
		defer close(ch)
		if _, ok := record[audioPort]; ok {
			record[audioPort].use = false
		}
	}
	<-ch
}
