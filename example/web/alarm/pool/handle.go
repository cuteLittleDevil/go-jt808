package pool

import (
	"log/slog"
	"runtime"
	"web/alarm/command"
	"web/alarm/pool/internal"
	"web/internal/shared"
)

type save interface {
	command.Location | command.BatchLocation | shared.T0x0801File
}

type Handle[T save] struct {
	queue   []chan T
	next    int
	cap     int
	failSum int
	db      *internal.DB
}

func NewHandle[T save]() *Handle[T] {
	num := runtime.NumCPU() + 1
	return &Handle[T]{
		queue: make([]chan T, num),
		next:  0,
		cap:   num,
		db:    internal.NewDB(),
	}
}

func (p *Handle[T]) Do(data T) {
	p.next++
	if p.next >= p.cap {
		p.next = 0
	}
	select {
	case p.queue[p.next] <- data:
	default:
		p.failSum++
		slog.Warn("queue full",
			slog.Int("fail sum", p.failSum),
			slog.Any("data", data))
	}
}

func (p *Handle[T]) Run() {
	p.queue = make([]chan T, p.cap)
	for i := 0; i < p.cap; i++ {
		p.queue[i] = make(chan T, 100)
		go func(taskCh <-chan T) {
			for task := range taskCh {
				switch v := any(task).(type) {
				case command.Location:
					p.db.InsertLocation(v)
				case command.BatchLocation:
					p.db.InsertLocationBatch(v)
				case shared.T0x0801File:
					p.db.InsertFileLocation(v)
				}
			}
		}(p.queue[i])
	}
}
