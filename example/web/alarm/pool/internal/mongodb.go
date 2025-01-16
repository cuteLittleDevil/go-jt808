package internal

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"time"
	"web/alarm/command"
	"web/internal/shared"
)

type mongodbDB struct {
	*mongo.Database
	collection string
}

func newMongodbDB(dsn string, database, collection string) *mongodbDB {
	// 创建客户端选项
	clientOptions := options.Client().ApplyURI(dsn)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	// 选择数据库和集合
	db := client.Database(database)
	// 创建时序集合
	var (
		metaField = "sim"
		//granularity    = "seconds"
		bucketMaxSpan  = 3600 * time.Second
		bucketRounding = 3600 * time.Second
	)
	opts := options.CreateCollection().SetTimeSeriesOptions(&options.TimeSeriesOptions{
		TimeField: "timestamp",
		MetaField: &metaField,
		//Granularity:    &granularity,
		BucketMaxSpan:  &bucketMaxSpan,
		BucketRounding: &bucketRounding,
	})
	if err := db.CreateCollection(context.Background(), collection, opts); err != nil {
		slog.Warn("create",
			slog.String("collection", collection),
			slog.Any("err", err))
	}
	return &mongodbDB{
		Database:   db,
		collection: collection,
	}
}

func (m *mongodbDB) InsertLocationBatch(data command.BatchLocation) {
	coll := m.Collection(m.collection)
	t0704 := data.T0x0704
	batchs := make([]any, 0, len(t0704.Items))
	if len(data.Items) > 0 {
		for _, v := range data.Items {
			batchs = append(batchs, bson.M{
				"timestamp": time.Now(),
				"sim":       data.Phone,
				"latitude":  v.Latitude,
				"longitude": v.Longitude,
				"dateTime":  v.DateTime,
			})
		}

		if _, err := coll.InsertMany(context.Background(), batchs); err != nil {
			slog.Warn("insert",
				slog.String("collection", m.collection),
				slog.Any("err", err))
		}
	}
}

func (m *mongodbDB) InsertLocation(data command.Location) {
	coll := m.Collection(m.collection)
	if _, err := coll.InsertOne(context.Background(), bson.M{
		"timestamp": time.Now(),
		"sim":       data.Phone,
		"latitude":  data.T0x0200.Latitude,
		"longitude": data.T0x0200.Longitude,
		"dateTime":  data.T0x0200.DateTime,
	}); err != nil {
		slog.Warn("insert",
			slog.String("collection", m.collection),
			slog.Any("err", err))
	}
}

func (m *mongodbDB) InsertFileLocation(data shared.T0x0801File) {
	coll := m.Collection(m.collection)
	if _, err := coll.InsertOne(context.Background(), bson.M{
		"timestamp":  time.Now(),
		"sim":        data.Phone,
		"latitude":   data.T0x0200LocationItem.Latitude,
		"longitude":  data.T0x0200LocationItem.Longitude,
		"dateTime":   data.T0x0200LocationItem.DateTime,
		"name":       data.Name,
		"objectName": data.ObjectName,
	}); err != nil {
		slog.Warn("insert",
			slog.String("collection", m.collection),
			slog.Any("err", err))
	}
}
