package crawler

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func findOne(postingId string) *posting {
	posting := posting{}
	err := connect().FindOne(context.TODO(), bson.M{"_id": postingId}).Decode(&posting)
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

func findAll() []posting {
	cur, err := connect().Find(context.TODO(), bson.M{})
	if err != nil {
		return nil // TODO
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
	return collection.FindOneAndReplace(
		context.TODO(),
		bson.M{"_id": posting.PostingId},
		posting,
		options.FindOneAndReplace().SetUpsert(true),
	)
}

func connect() *mongo.Collection {
	credential := options.Credential{
		Username: "root",
		Password: "example",
	}

	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017").
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

	return client.Database("fundgrube").Collection("postings")
}
