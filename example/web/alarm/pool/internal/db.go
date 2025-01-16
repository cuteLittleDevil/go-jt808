package internal

import (
	"web/alarm/command"
	"web/alarm/conf"
	"web/internal/shared"
)

type DataInserter interface {
	InsertLocationBatch(data command.BatchLocation)
	InsertLocation(data command.Location)
	InsertFileLocation(data shared.T0x0801File)
}

type DB struct {
	inserts []DataInserter
}

func NewDB() *DB {
	tmp := &DB{
		inserts: make([]DataInserter, 0),
	}
	if tc := conf.GetData().TdengineConfig; tc.Enable {
		tmp.inserts = append(tmp.inserts, newTdengineDB(tc.Dsn, tc.Database, tc.SuperTable))
	}

	if mc := conf.GetData().MongodbConfig; mc.Enable {
		tmp.inserts = append(tmp.inserts, newMongodbDB(mc.Dsn, mc.Database, mc.Collection))
	}
	return tmp
}

func (d *DB) InsertLocationBatch(data command.BatchLocation) {
	for _, insert := range d.inserts {
		insert.InsertLocationBatch(data)
	}
}

func (d *DB) InsertLocation(data command.Location) {
	for _, insert := range d.inserts {
		insert.InsertLocation(data)
	}
}

func (d *DB) InsertFileLocation(data shared.T0x0801File) {
	for _, insert := range d.inserts {
		insert.InsertFileLocation(data)
	}
}
