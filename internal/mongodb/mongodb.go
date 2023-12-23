package mongodb

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

var mongoDB *MongoDB

func InitMongoDB() {
	connectionURL := "mongodb://root:root@localhost:27017"
	clientOptions := options.Client().ApplyURI(connectionURL)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB server: %v", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB server: %v", err)
		client.Disconnect(context.Background())
	}

	log.Println("Connected to MongoDB!")

	mongoDB = &MongoDB{
		client:   client,
		database: client.Database("logger"),
	}
}

func CloseMongoDB() {
	if mongoDB != nil {
		mongoDB.client.Disconnect(context.Background())
		log.Println("Disconnected from MongoDB")
	}
}

func InsertDocument(collectionName string, document interface{}) error {
	collection := mongoDB.database.Collection(collectionName)
	_, err := collection.InsertOne(context.Background(), document)
	return err
}

func InsertDocuments(collectionName string, documents []interface{}) error {
	collection := mongoDB.database.Collection(collectionName)

	opts := options.BulkWrite().SetOrdered(false)
	opts.SetBypassDocumentValidation(true)

	bulkModels := make([]mongo.WriteModel, len(documents))
	for i, doc := range documents {
		bulkModels[i] = mongo.NewInsertOneModel().SetDocument(doc)
	}

	result, err := collection.BulkWrite(context.Background(), bulkModels, opts)
	if err != nil {
		log.Printf("Failed to perform bulk write: %v", err)
		return err
	}

	log.Printf("Inserted %d documents", result.InsertedCount)

	return nil
}

func FindLatestDocuments(collectionName string) ([]interface{}, error) {
	collection := mongoDB.database.Collection(collectionName)

	findOptions := options.Find().SetSort(
		bson.D{{Key: "timestamp", Value: -1}}).SetLimit(5)

	cursor, err := collection.Find(context.Background(), bson.D{}, findOptions)
	if err != nil {
		log.Printf("Failed to execute find operation: %v", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []interface{}
	for cursor.Next(context.Background()) {
		var result interface{}
		if err := cursor.Decode(&result); err != nil {
			log.Printf("Failed to decode document: %v", err)
			return nil, err
		}
		results = append(results, result)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor iteration error: %v", err)
		return nil, err
	}

	return results, nil
}
