package crawler

import (
	"context"
	log "github.com/sirupsen/logrus"
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

func FindOne(postingId string) *posting {
	posting := posting{}
	err := postingsCollection().FindOne(context.TODO(), bson.M{"_id": postingId}).Decode(&posting)
	if err != nil {
		return nil
	}
	return &posting
}

func FindAll(q query, afterTime *time.Time, limit int64, offset int64) []posting {
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
		filter["discount_in_percent"] = bson.M{"$gte": q.DiscountMin}
	}
	if q.OutletId != nil {
		filter["outlet.id"] = bson.M{"$eq": q.OutletId}
	}
	if q.Ids != nil {
		filter["_id"] = bson.M{"$in": q.Ids}
	} else if !q.FindInactive {
		filter["active"] = bson.M{"$eq": true}
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
	if priceMin != nil {
		return bson.M{"$gte": priceMin}
	}
	if priceMax != nil {
		return bson.M{"$lte": priceMax}
	}
	panic("priceFilter called without priceMin or priceMax set")
}

func SaveAllNewOrUpdated(postings []posting) (insertedCount int, updatedCount int, took time.Duration) {
	start := time.Now()
	loadedPostings := loadAll(postings)

	postingsToUpsert := []posting{}

	for _, posting := range postings {
		existing, ok := loadedPostings[posting.PostingId]
		if !ok {
			posting.CreDat = &start
			posting.ModDat = &start
			postingsToUpsert = append(postingsToUpsert, posting)
		} else {
			posting.CreDat = existing.CreDat
			posting.ModDat = existing.ModDat

			if !reflect.DeepEqual(existing, posting) {
				posting.ModDat = &start
				postingsToUpsert = append(postingsToUpsert, posting)
			}
		}
	}

	insertedCount, updatedCount = insertOrUpdateAll(postingsToUpsert)
	return insertedCount, updatedCount, time.Since(start)
}

func insertOrUpdateAll(postings []posting) (insertedCount int, updatedCount int) {
	if len(postings) == 0 {
		return 0, 0
	}

	var operations []mongo.WriteModel
	for _, p := range postings {
		update := mongo.NewReplaceOneModel()
		update.SetFilter(bson.M{"_id": p.PostingId})
		update.SetReplacement(p)
		update.SetUpsert(true)

		operations = append(operations, update)
	}

	write, err := postingsCollection().BulkWrite(context.TODO(), operations)
	if err != nil {
		panic(err)
	}
	return int(write.UpsertedCount), int(write.ModifiedCount)
}

func loadAll(postings []posting) map[string]posting {
	start := time.Now()

	loadedPostings := FindAll(query{Ids: toIds(postings)}, nil, int64(len(postings)), 0)

	ret := make(map[string]posting)
	for _, loadedPosting := range loadedPostings {
		ret[loadedPosting.PostingId] = loadedPosting
	}
	log.Debugf("Loaded %d existing postings for diff in %.2fs", len(ret), time.Since(start).Seconds())
	return ret
}

func SetRemainingPostingInactive(shop Shop, c category, outlets []outlet, postingIds []string) int {
	filter := bson.M{"shop": shop, "category.id": c.CategoryId, "_id": bson.M{"$nin": postingIds}}
	if outlets != nil && len(outlets) > 0 {
		filter["outlet.id"] = bson.M{"$in": outletIds(outlets)}
	}

	many, err := postingsCollection().UpdateMany(
		context.TODO(),
		filter,
		bson.M{"$set": bson.M{"active": false}},
	)
	if err != nil {
		panic(err)
	}
	return int(many.ModifiedCount)
}

func clearAll() {
	_, err := postingsCollection().DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
}

func Migrate(filterString string, updateString string) int {
	if !envBool("MIGRATE") {
		return dryRunFilter(filterString)
	}

	manyResponse, err := postingsCollection().UpdateMany(context.TODO(), toBson(filterString), toBson(updateString))
	if err != nil {
		panic(err)
	}
	migratedCount := int(manyResponse.ModifiedCount)
	log.Warnf("Migrated %d entries in posting collection filter: '%s' and update: '%s'", migratedCount, filterString, updateString)
	return migratedCount
}

func CleanUp(filterString string) int {
	if !envBool("CLEANUP") {
		return dryRunFilter(filterString)
	}

	deleteMany, err := postingsCollection().DeleteMany(context.TODO(), toBson(filterString))
	if err != nil {
		panic(err)
	}

	deletedCount := int(deleteMany.DeletedCount)
	log.Warnf("Deleted %d entries in posting collection with filter: '%s'", deletedCount, filterString)
	return deletedCount
}

func dryRunFilter(filterString string) int {
	cursor, err := postingsCollection().Find(context.TODO(), toBson(filterString))
	if err != nil {
		panic(err)
	}
	log.Warnf("[DRY_RUN] Filter '%s' matches %d elements.", filterString, cursor.RemainingBatchLength())
	return 0
}

func toBson(jsonString string) interface{} {
	var bsonFilter interface{}
	err := bson.UnmarshalExtJSON([]byte(jsonString), true, &bsonFilter)
	if err != nil {
		panic(err)
	}
	return bsonFilter
}

func updateSearchOperation(query query, now *time.Time) *mongo.SingleResult {
	md5Hex := hashQuery(query)
	op := operation{md5Hex, query.Desc, query, now}
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
		SetTimeout(15 * time.Second)

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
