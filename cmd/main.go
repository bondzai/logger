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

	connectionURL := "amqp://guest:guest@localhost:5672/"
	logQueueName := "log"
	failedQueueName := "failed"

	consumer, err := rabbitmq.NewConsumer(connectionURL, logQueueName, failedQueueName)
	if err != nil {
		log.Fatal("Failed to create RabbitMQ consumer:", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	go func() {
		defer wg.Done()
		<-signals
		log.Println("Received termination signal. Stopping consumer...")
		consumer.Stop()
	}()

	wg.Add(1)

	consumer.Start(processMessage)

	log.Printf("Consumer started. To exit, press CTRL+C")
	wg.Wait()
}

// In your main.go file

func processMessage(message map[string]interface{}) bool {
	log.Printf("Received message: %+v", message)
	// Process your message...

	// Example: Inserting a document into a collection named "logs"
	err := mongodb.InsertDocument("logs", message)
	if err != nil {
		log.Printf("Failed to insert document into MongoDB: %v", err)
		return false
	}

	log.Printf("Message processed and inserted into MongoDB")
	return true
}
