package crawler

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"reflect"
	"time"
)

func findOne(postingId string) *posting {
	return _findOne(connect(), postingId)
}

func _findOne(collection *mongo.Collection, postingId string) *posting {
	posting := posting{}
	err := collection.FindOne(context.TODO(), bson.M{"_id": postingId}).Decode(&posting)
	if err != nil {
		return nil
	}
	return &posting
}

func saveOne(posting posting) {
	_saveOne(connect(), posting)
}

func saveAll(postings []posting) {
	collection := connect()
	for _, posting := range postings {
		_saveOne(collection, posting)
	}
}

func findAll(afterTime time.Time, limit int64, offset int64) []posting {
	filter := bson.M{"mod_dat": bson.M{"$gte": primitive.NewDateTimeFromTime(afterTime)}}
	findOptions := options.Find().SetLimit(limit).SetSkip(offset).SetSort(bson.M{"price": 1})
	cur, err := connect().Find(context.TODO(), filter, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	postings := []posting{}
	for cur.Next(context.TODO()) {
		var elem posting
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		postings = append(postings, elem)
	}
	return postings
}

func clearAll() {
	_, err := connect().DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
}

func _saveOne(collection *mongo.Collection, posting posting) *mongo.SingleResult {
	existing := _findOne(collection, posting.PostingId)
	now := time.Now()

	if existing == nil {
		posting.CreDat = &now
		posting.ModDat = &now

		return _save(collection, posting)
	}

	// set cre_dat & mod_dat so we can use equals
	posting.CreDat = existing.CreDat
	posting.ModDat = existing.ModDat

	if !reflect.DeepEqual(*existing, posting) {
		posting.ModDat = &now

		return _save(collection, posting)
	}
	return nil // no update
}

func _save(collection *mongo.Collection, posting posting) *mongo.SingleResult {
	return collection.FindOneAndReplace(
		context.TODO(),
		bson.M{"_id": posting.PostingId},
		posting,
		options.FindOneAndReplace().SetUpsert(true),
	)
}

func connect() *mongo.Collection {
	credential := options.Credential{
		Username: env("MONGODB_USERNAME", "root"),
		Password: env("MONGODB_PASSWORD", "example"),
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(env("MONGODB_URI", "mongodb://localhost:27017")).
		SetAuth(credential)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err) // TODO: ugly
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	return client.Database(env("MONGODB_DB", "fundgrube")).Collection(env("MONGODB_COLLECTION", "postings"))
}

func env(key string, defaultValue string) string {
	value, present := os.LookupEnv(key)
	if present {
		return value
	}
	return defaultValue
}
