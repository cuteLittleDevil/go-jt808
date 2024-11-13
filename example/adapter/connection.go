package main

import (
	"fmt"
	"log/slog"
	"net"
)

type connection struct {
	sim            string
	addr           string
	srcChan        chan<- []byte
	localWriteChan chan []byte
	quit           bool
	stopChan       <-chan struct{}
}

func newConnection(sim string, addr string, srcChan chan<- []byte, stopChan <-chan struct{}) *connection {
	return &connection{
		sim:            sim,
		addr:           addr,
		srcChan:        srcChan,
		localWriteChan: make(chan []byte, 10),
		stopChan:       stopChan,
	}
}

func (c *connection) run() {
	terminal, err := net.Dial("tcp", c.addr)
	if err != nil {
		slog.Warn("conn",
			slog.String("addr", c.addr),
			slog.Any("err", err))
		return
	}
	slog.Debug("加入",
		slog.String("sim", c.sim),
		slog.String("addr", terminal.LocalAddr().String()))
	defer func() {
		_ = terminal.Close()
		close(c.localWriteChan)
		c.quit = true
		slog.Debug("退出",
			slog.String("sim", c.sim),
			slog.String("addr", terminal.LocalAddr().String()))
	}()

	curData := make([]byte, 1023)
	for {
		select {
		case <-c.stopChan:
			return
		case data := <-c.localWriteChan:
			if n, err := terminal.Write(data); err != nil {
				slog.Warn("write",
					slog.String("sim", c.sim),
					slog.String("data", fmt.Sprintf("%x", data[:n])),
					slog.Any("err", err))
				return
			}
		}
		if n, err := terminal.Read(curData); err != nil {
			slog.Warn("conn",
				slog.String("sim", c.sim),
				slog.Any("err", err))
			return
		} else if n > 0 {
			c.srcChan <- curData[:n]
		}
	}
}
