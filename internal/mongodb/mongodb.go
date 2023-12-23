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

// NewMongoDB creates a new instance of MongoDB.
func NewMongoDB() *MongoDB {
	return &MongoDB{}
}

func (m *MongoDB) Connect(connectionURL, dbName string) error {
	clientOptions := options.Client().ApplyURI(connectionURL)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB server: %v", err)
		return err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB server: %v", err)
		client.Disconnect(context.Background())
		return err
	}

	log.Println("Connected to MongoDB!")

	m.client = client
	m.database = client.Database(dbName)
	return nil
}

func (m *MongoDB) CloseMongoDB() {
	if m.client != nil {
		m.client.Disconnect(context.Background())
		log.Println("Disconnected from MongoDB")
	}
}

func (m *MongoDB) InsertDocument(collectionName string, document interface{}) error {
	collection := m.database.Collection(collectionName)
	_, err := collection.InsertOne(context.Background(), document)
	return err
}

func (m *MongoDB) InsertDocuments(collectionName string, documents []interface{}) error {
	collection := m.database.Collection(collectionName)

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

func (m *MongoDB) FindDocuments(collectionName string, query bson.D, findOptions *options.FindOptions) ([]interface{}, error) {
	collection := m.database.Collection(collectionName)

	cursor, err := collection.Find(context.Background(), query, findOptions)
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
