package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
	ulid "github.com/oklog/ulid/v2"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const NumIDs = 10_000_000

func connect(ctx context.Context) (*mongo.Database, error) {
	client, err := mongo.Connect(ctx)
	if err != nil {
		logrus.WithError(err).Error("Unable to build mongo connection!")
		return nil, err
	}

	// check connection works as mongo-go lazily connects.
	err = client.Ping(ctx, nil)
	if err != nil {
		logrus.WithError(err).Error("Unable to connect to mongo! Have you configured your connection properly?")
		return nil, err
	}

	return client.Database("testIds"), nil
}

func genUUID() string {
	return uuid.New().String()
}

var entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)

func genULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

func configureIndex(ctx context.Context, collection *mongo.Collection) {
	index := mongo.IndexModel{
		Keys: bson.D{
			{
				Key:   "id",
				Value: int32(1),
			},
		},
		Options: options.Index().
			SetBackground(true).
			SetName(collection.Name() + "_id").
			SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, index)
	if err != nil {
		logrus.WithError(err).Error("error creating index")
		checkErr(err)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func insertRecords(ctx context.Context, collection *mongo.Collection, generator func() string) {
	cacheSize := NumIDs / 100
	for i := 0; i < 100; i++ {
		docs := make([]interface{}, cacheSize)
		for j := range docs {
			docs[j] = bson.D{
				{"id", generator()},
			}
		}
		_, err := collection.InsertMany(ctx, docs)
		checkErr(err)
		docs = nil
	}
}

func generateComparison(ctx context.Context, db *mongo.Database, collName string, generator func() string) {
	collection := db.Collection(collName)
	configureIndex(ctx, collection)
	insertRecords(ctx, collection, generator)
}

func main() {
	ctx := context.Background()
	db, err := connect(ctx)
	checkErr(err)

	generateComparison(ctx, db, "uuids", genUUID)
	generateComparison(ctx, db, "ulids", genULID)
}
