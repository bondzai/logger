package mongodb

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Context = context.Background()
var Database *mongo.Database

func Initial() {
	connection, err := connectToMongoDB()
	if err != nil {
		log.Fatal(err)
	} else {
		Database = connection.Database("logger")
	}
}

func connectToMongoDB() (*mongo.Client, error) {
	url := "mongodb://root:root@localhost:27017"

	clientOptions := options.Client().ApplyURI(url)

	client, err := mongo.Connect(Context, clientOptions)
	if err != nil {
		log.Printf("Failed to connect to MongoDB server: %v", err)
		return nil, err
	}

	err = client.Ping(Context, nil)
	if err != nil {
		log.Printf("Failed to ping MongoDB server: %v", err)
		client.Disconnect(Context)
		return nil, err
	}

	log.Println("Connected to MongoDB!")
	return client, nil
}
