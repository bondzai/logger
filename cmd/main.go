package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bondzai/logger/internal/mongodb"
	"github.com/bondzai/logger/internal/rabbitmq"
)

func init() {
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("initial started")

	mongodb.InitMongoDB()
}

func main() {
	defer mongodb.CloseMongoDB()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

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

	wg.Add(1)

	rabbitMQConsumer.Start(processMessage)

	log.Printf("Consumer started. To exit, press CTRL+C")
	wg.Wait()
}

func processMessage(message map[string]interface{}) bool {
	log.Printf("Received message: %+v", message)

	err := mongodb.InsertDocument("logs", message)
	if err != nil {
		log.Printf("Failed to insert document into MongoDB: %v", err)
		return false
	}

	log.Printf("Message processed and inserted into MongoDB")
	return true
}
