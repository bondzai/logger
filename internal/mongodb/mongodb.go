package mongodb

import (
	"context"
	"log"

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
