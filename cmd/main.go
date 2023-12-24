package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bondzai/logger/internal/api"
	"github.com/bondzai/logger/internal/mongodb"
	"github.com/bondzai/logger/internal/rabbitmq"
)

const (
	mongoURL  = "mongodb://root:root@localhost:27017"
	mongoDB   = "logger"
	mongoCol  = "logs"
	rabbitURL = "amqp://guest:guest@localhost:5672/"
	rabbitKey = "logs"
)

func init() {
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	mongo := mongodb.NewMongoDB()
	err := mongo.Connect(mongoURL, mongoDB)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongo.CloseMongoDB()

	go func() {
		defer wg.Done()

		err = api.StartGRPCServer(mongo)
		if err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	rabbitMQConsumer, err := rabbitmq.NewConsumer(rabbitURL, rabbitKey)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ consumer: %v", err)
	}

	go func() {
		defer wg.Done()
		<-signals
		log.Println("Received termination signal. Stopping consumer...")
		rabbitMQConsumer.Stop()
	}()

	wg.Add(2)

	go func() {
		defer wg.Done()
		rabbitMQConsumer.Start(func(message map[string]interface{}) bool {
			return processMessage(mongo, message)
		})
	}()

	log.Printf("Consumer and gRPC server started. To exit, press CTRL+C")
	wg.Wait()
}

func processMessage(mongo *mongodb.MongoDB, message map[string]interface{}) bool {
	err := mongo.InsertDocument(mongoCol, message)
	if err != nil {
		log.Printf("Failed to insert document into MongoDB: %v", err)
		return false
	}

	log.Printf("Message processed and inserted into MongoDB: %+v", message)
	return true
}
