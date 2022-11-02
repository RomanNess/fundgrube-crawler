package crawler

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"reflect"
	"time"
)

var collectionPostings *mongo.Collection
var collectionOperations *mongo.Collection

func findOne(postingId string) *posting {
	return _findOne(postingId)
}

func _findOne(postingId string) *posting {
	posting := posting{}
	err := postingsCollection().FindOne(context.TODO(), bson.M{"_id": postingId}).Decode(&posting)
	if err != nil {
		return nil
	}
	return &posting
}

func findAll(q query, afterTime *time.Time, limit int64, offset int64) []posting {
	filter := bson.M{}
	if afterTime != nil {
		filter["mod_dat"] = bson.M{"$gte": primitive.NewDateTimeFromTime(*afterTime)}
	}
	if q.NameRegex != nil {
		filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: *q.NameRegex, Options: "i"}}
	}
	if q.BrandRegex != nil {
		filter["brand.name"] = bson.M{"$regex": primitive.Regex{Pattern: *q.BrandRegex, Options: "i"}}
	}
	if q.PriceMin != nil || q.PriceMax != nil {
		filter["price"] = priceFilter(q.PriceMin, q.PriceMax)
	}
	if q.DiscountMin != nil {
		filter["discount"] = bson.M{"$gte": q.DiscountMin}
	}
	if q.OutletId != nil {
		filter["outlet.id"] = bson.M{"$eq": q.OutletId}
	}

	findOptions := options.Find().SetLimit(limit).SetSkip(offset).SetSort(bson.M{"price": 1})
	cur, err := postingsCollection().Find(context.TODO(), filter, findOptions)
	if err != nil {
		panic(err)
	}
	postings := []posting{}
	for cur.Next(context.TODO()) {
		var elem posting
		err := cur.Decode(&elem)
		if err != nil {
			panic(err)
		}
		postings = append(postings, elem)
	}
	return postings
}

func priceFilter(priceMin *float64, priceMax *float64) bson.M {
	if priceMin != nil && priceMax != nil {
		return bson.M{"$gte": priceMin, "$lte": priceMax}
	}
	if priceMin == nil {
		return bson.M{"$gte": priceMin}
	}
	if priceMax == nil {
		return bson.M{"$lte": priceMax}
	}
	panic("priceFilter called without priceMin or priceMax set")
}

func saveOne(posting posting) (inserted int, updated int) {
	existing := _findOne(posting.PostingId)
	now := time.Now()

	if existing == nil {
		posting.CreDat = &now
		posting.ModDat = &now

		_, err := postingsCollection().InsertOne(
			context.TODO(),
			posting,
		)
		if err != nil {
			panic(err)
		}
		return 1, 0
	}

	// set cre_dat & mod_dat so we can use equals
	posting.CreDat = existing.CreDat
	posting.ModDat = existing.ModDat

	if !reflect.DeepEqual(*existing, posting) {
		posting.ModDat = &now

		postingsCollection().FindOneAndReplace(
			context.TODO(),
			bson.M{"_id": posting.PostingId},
			posting,
			options.FindOneAndReplace().SetUpsert(true),
		)
		return 0, 1
	}
	return 0, 0
}

func saveAll(postings []posting) (int, int) {
	overallInserted, overallUpdated := 0, 0
	for _, posting := range postings {
		inserted, updated := saveOne(posting)
		overallInserted = overallInserted + inserted
		overallUpdated = overallUpdated + updated
	}
	return overallInserted, overallUpdated
}

func clearAll() {
	_, err := postingsCollection().DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
}

func updateSearchOperation(query query, now *time.Time) *mongo.SingleResult {
	md5Hex := hashQuery(query)
	op := operation{md5Hex, query.String(), query, now}
	return operationsCollection().FindOneAndReplace(
		context.TODO(),
		bson.M{"_id": md5Hex},
		op,
		options.FindOneAndReplace().SetUpsert(true),
	)
}

func findSearchOperation(id string) *operation {
	op := operation{}
	err := operationsCollection().FindOne(context.TODO(), bson.M{"_id": id}).Decode(&op)
	if err != nil {
		return nil
	}
	return &op
}

func postingsCollection() *mongo.Collection {
	if collectionPostings == nil {
		collectionPostings = connect(env("MONGODB_COLLECTION_POSTINGS", "postings"))
	}
	return collectionPostings
}

func operationsCollection() *mongo.Collection {
	if collectionOperations == nil {
		collectionOperations = connect(env("MONGODB_COLLECTION_OPERATIONS", "operations"))
	}
	return collectionOperations
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
		panic(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
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
