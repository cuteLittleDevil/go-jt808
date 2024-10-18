package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/nats-io/nats.go"
	_ "github.com/taosdata/driver-go/v3/taosWS"
	"log/slog"
	"os"
	"os/signal"
	"simulator/internal/mq"
	"simulator/internal/shared"
	"strconv"
	"syscall"
)

func main() {
	var natsAddr string
	var dsn string
	flag.StringVar(&natsAddr, "nats", "127.0.0.1:4222", "nats address")
	flag.StringVar(&dsn, "dsn", "root:taosdata@ws(127.0.0.1:6041)/information_schema?tz=Shanghai&parseTime=true&loc=Local", "dsn")
	flag.Parse()

	if err := mq.Init(natsAddr); err != nil {
		slog.Error("nats init fail",
			slog.String("nats", natsAddr),
			slog.Any("err", err))
		return
	}

	if err := createTable(dsn); err != nil {
		slog.Error("create table fail",
			slog.String("dsn", dsn),
			slog.Any("err", err))
		return
	}

	pool := newPoolHandle(dsn)
	go pool.run()

	mq.Default().Run(map[string]nats.MsgHandler{
		shared.SubLocation: func(msg *nats.Msg) {
			pool.Do(&task{
				data: msg.Data,
				sub:  shared.SubLocation,
			})
		},
		shared.SubLocationBatch: func(msg *nats.Msg) {
			pool.Do(&task{
				data: msg.Data,
				sub:  shared.SubLocationBatch,
			})
		},
	})

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT) // kill -2
	<-quit
}

func createTable(dsn string) error {
	// https://docs.taosdata.com/reference/taos-sql/database/ keep 保存的有效期
	// 删库命令 drop database power
	db, err := sql.Open("taosWS", dsn)
	if err != nil {
		return err
	}
	_, _ = db.Exec("drop database if exists power")
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS power PRECISION 'us' buffer 30 duration 1h keep 72h "); err != nil {
		return err
	}
	if _, err := db.Exec("CREATE STABLE IF NOT EXISTS power.meters " +
		"(ts TIMESTAMP,lat INT, lon INT) TAGS (phone BINARY(24))"); err != nil {
		return err
	}
	// 提前创建1w个表
	// CREATE TABLE IF NOT EXISTS power.T1 USING power.meters TAGS 1
	for i := 0; i < 10000; i++ {
		if _, err := db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS power.T%d "+
			"USING power.meters TAGS (%s)", i+1, strconv.Itoa(i+1))); err != nil {
			return err
		}
	}
	_ = db.Close()
	return nil
}
