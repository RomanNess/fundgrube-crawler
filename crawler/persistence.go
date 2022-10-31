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
	return _findOne(connectPostings(), postingId)
}

func _findOne(collection *mongo.Collection, postingId string) *posting {
	posting := posting{}
	err := collection.FindOne(context.TODO(), bson.M{"_id": postingId}).Decode(&posting)
	if err != nil {
		return nil
	}
	return &posting
}

func findAll(regex *string, afterTime *time.Time, limit int64, offset int64) []posting {
	filter := bson.M{}
	if afterTime != nil {
		filter["mod_dat"] = bson.M{"$gte": primitive.NewDateTimeFromTime(*afterTime)}
	}
	if regex != nil {
		filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: *regex, Options: "i"}}
	}
	findOptions := options.Find().SetLimit(limit).SetSkip(offset).SetSort(bson.M{"price": 1})
	cur, err := connectPostings().Find(context.TODO(), filter, findOptions)
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

func saveOne(posting posting) {
	_saveOne(connectPostings(), posting)
}

func saveAll(postings []posting) {
	collection := connectPostings()
	for _, posting := range postings {
		_saveOne(collection, posting)
	}
}

func clearAll() {
	_, err := connectPostings().DeleteMany(context.TODO(), bson.M{})
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

func updateSearchOperation(query query, now *time.Time) *mongo.SingleResult {
	md5Hex := hashQuery(query)
	op := operation{md5Hex, query.Regex, now}
	return connectOperations().FindOneAndReplace(
		context.TODO(),
		bson.M{"_id": md5Hex},
		op,
		options.FindOneAndReplace().SetUpsert(true),
	)
}

func findSearchOperation(id string) *operation {
	op := operation{}
	err := connectOperations().FindOne(context.TODO(), bson.M{"_id": id}).Decode(&op)
	if err != nil {
		return nil
	}
	return &op
}

func connectPostings() *mongo.Collection {
	return connect(env("MONGODB_COLLECTION_POSTINGS", "postings"))
}

func connectOperations() *mongo.Collection {
	return connect(env("MONGODB_COLLECTION_OPERATIONS", "operations"))
}

func connect(collectionName string) *mongo.Collection {
	credential := options.Credential{
		Username: env("MONGODB_USERNAME", "root"),
		Password: env("MONGODB_PASSWORD", "example"),
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(env("MONGODB_URI", "mongodb://localhost:27017")).
		SetAuth(credential).
		SetTimeout(5 * time.Second)

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

	return client.Database(env("MONGODB_DB", "fundgrube")).Collection(collectionName)
}

func env(key string, defaultValue string) string {
	value, present := os.LookupEnv(key)
	if present {
		return value
	}
	return defaultValue
}
