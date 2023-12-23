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

func init() {
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	mongoDB := mongodb.NewMongoDB()
	err := mongoDB.Connect("mongodb://root:root@localhost:27017", "logger")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.CloseMongoDB()

	// Start gRPC server
	go func() {
		defer wg.Done()

		err = api.StartGRPCServer(mongoDB)
		if err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// RabbitMQ consumer setup
	rabbitMQConsumer, err := rabbitmq.NewConsumer("amqp://guest:guest@localhost:5672/", "log")
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

	// Start RabbitMQ consumer
	go func() {
		defer wg.Done()
		rabbitMQConsumer.Start(func(message map[string]interface{}) bool {
			return processMessage(mongoDB, message)
		})
	}()

	log.Printf("Consumer and gRPC server started. To exit, press CTRL+C")
	wg.Wait()
}

func processMessage(mongoDB *mongodb.MongoDB, message map[string]interface{}) bool {
	err := mongoDB.InsertDocument("logs", message)
	if err != nil {
		log.Printf("Failed to insert document into MongoDB: %v", err)
		return false
	}

	log.Printf("Message processed and inserted into MongoDB: %+v", message)
	return true
}
