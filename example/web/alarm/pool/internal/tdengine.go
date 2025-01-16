package internal

import (
	"database/sql"
	"fmt"
	_ "github.com/taosdata/driver-go/v3/taosWS"
	"log/slog"
	"time"
	"web/alarm/command"
	"web/internal/shared"
)

type tdengineDB struct {
	*sql.DB
	database   string
	superTable string
}

func newTdengineDB(dsn string, database, superTable string) *tdengineDB {
	client, err := sql.Open("taosWS", dsn)
	if err != nil {
		panic(err)
	}

	if _, err := client.Exec("CREATE DATABASE IF NOT EXISTS " + database +
		" PRECISION 'us' buffer 30 duration 1h keep 72h "); err != nil {
		slog.Warn("create",
			slog.String("database", database),
			slog.String("super table", superTable),
			slog.Any("err", err))
	}
	if _, err := client.Exec("CREATE STABLE IF NOT EXISTS " + database + "." + superTable +
		"(ts TIMESTAMP,lat INT, lon INT, name BINARY(50)) TAGS (phone BINARY(24))"); err != nil {
		slog.Warn("create",
			slog.String("database", database),
			slog.String("super table", superTable),
			slog.Any("err", err))
	}
	return &tdengineDB{
		DB:         client,
		database:   database,
		superTable: superTable,
	}
}

func (t *tdengineDB) InsertLocationBatch(batch command.BatchLocation) {
	phone := batch.Phone
	str := "INSERT INTO " +
		fmt.Sprintf("power.'T%s' USING %s.%s TAGS('%s') ",
			phone, t.database, t.superTable, phone) +
		"VALUES "
	t0704 := batch.T0x0704
	if len(t0704.Items) > 0 {
		for _, v := range t0704.Items {
			str += fmt.Sprintf("(%d, '%d', '%d', '') ", time.Now().UnixMicro(), v.Latitude, v.Longitude)
		}
		if _, err := t.Exec(str); err != nil {
			slog.Warn("insert location batch fail",
				slog.String("sql", str),
				slog.Any("err", err))
		}
	}
}

func (t *tdengineDB) InsertLocation(data command.Location) {
	t0200 := data.T0x0200
	phone := data.Phone
	str := fmt.Sprintf("INSERT INTO %s.T%s USING %s.%s TAGS('%s') VALUES(%d, %d, %d, '') ",
		t.database, phone, t.database, t.superTable, phone,
		time.Now().UnixMicro(), t0200.Latitude, t0200.Longitude)
	if _, err := t.Exec(str); err != nil {
		slog.Warn("insert location fail",
			slog.String("sql", str),
			slog.Any("err", err))
	}
}

func (t *tdengineDB) InsertFileLocation(data shared.T0x0801File) {
	t0200 := data.T0x0200LocationItem
	phone := data.Phone
	str := fmt.Sprintf("INSERT INTO %s.T%s USING %s.%s TAGS('%s') VALUES(%d, %d, %d, '%s') ",
		t.database, phone, t.database, t.superTable, phone,
		time.Now().UnixMicro(), t0200.Latitude, t0200.Longitude, data.ObjectName)
	if _, err := t.Exec(str); err != nil {
		slog.Warn("insert location fail",
			slog.String("sql", str),
			slog.Any("err", err))
	}
}
