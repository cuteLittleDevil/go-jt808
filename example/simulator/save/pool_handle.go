package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"runtime"
	"simulator/internal/shared"
	"time"
)

type task struct {
	data []byte
	sub  string
}

type poolHandle struct {
	dsn     string
	queue   []chan *task
	next    int
	cap     int
	failSum int
}

func newPoolHandle(dsn string) *poolHandle {
	sum := 2*runtime.NumCPU() + 1
	return &poolHandle{
		dsn:  dsn,
		cap:  sum,
		next: 0,
	}
}

func (p *poolHandle) Do(task *task) {
	p.next++
	if p.next >= p.cap {
		p.next = 0
	}
	select {
	case p.queue[p.next] <- task:
	default:
		p.failSum++
		slog.Warn("queue full",
			slog.String("sub", task.sub),
			slog.Int("fail sum", p.failSum),
			slog.Any("data", task.data))
	}
}

func (p *poolHandle) run() {
	p.queue = make([]chan *task, p.cap)
	for i := 0; i < p.cap; i++ {
		p.queue[i] = make(chan *task, 100)
		go func(taskCh <-chan *task) {
			db, err := sql.Open("taosWS", p.dsn)
			if err != nil {
				panic(err)
			}
			var (
				location      shared.Location
				locationBatch shared.LocationBatch
			)
			outTimeSecond := 3 * time.Second
			timer := time.NewTimer(outTimeSecond)
			var data []shared.Location
			for {
				select {
				case <-timer.C:
					if len(data) > 0 {
						insertLocation(db, data)
						data = data[0:0]
					}
					timer.Reset(outTimeSecond)
				case v := <-taskCh:
					switch v.sub {
					case shared.SubLocation:
						if err := location.Decode(v.data); err == nil {
							data = append(data, location)
							if len(data) >= 10 {
								insertLocation(db, data)
								data = data[0:0]
							}
							timer.Reset(outTimeSecond)
						}
					case shared.SubLocationBatch:
						if err := locationBatch.Decode(v.data); err == nil {
							insertLocationBatch(db, locationBatch)
						}
					}
				}
			}
		}(p.queue[i])
	}
}

func insertLocation(db *sql.DB, locations []shared.Location) {
	str := "INSERT INTO "
	for _, v := range locations {
		str += fmt.Sprintf("power.T%s USING power.meters TAGS('%s') VALUES(%d, %d, %d) ",
			v.Phone, v.Phone, v.TimestampMicro, v.Latitude, v.Longitude)
	}
	if _, err := db.Exec(str); err != nil {
		slog.Warn("insert location fail",
			slog.String("sql", str),
			slog.Any("err", err))
	}
}

func insertLocationBatch(db *sql.DB, batch shared.LocationBatch) {
	phone := batch.Locations[0].Phone
	str := "INSERT INTO " +
		fmt.Sprintf("power.'T%s' USING power.meters TAGS('%s')  ", phone, phone) +
		"VALUES "
	for _, v := range batch.Locations {
		str += fmt.Sprintf("(%d, '%d', '%d') ", v.TimestampMicro, v.Latitude, v.Longitude)
	}
	if _, err := db.Exec(str); err != nil {
		slog.Warn("insert location batch fail",
			slog.String("sql", str),
			slog.Any("err", err))
	}
}
