package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bondzai/logger/internal/mongodb"
	"github.com/bondzai/logger/internal/rabbitmq"
)

func init() {
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("initial started")

	mongodb.Initial()
}

func main() {
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
		defer wg.Done() // Decrement the WaitGroup counter when done
		<-signals
		log.Println("Received termination signal. Stopping consumer...")
		consumer.Stop()
	}()

	wg.Add(1) // Increment the WaitGroup counter

	consumer.Start(processMessage)

	log.Printf("Consumer started. To exit, press CTRL+C")
	wg.Wait() // Wait for all goroutines to finish before exiting
}

func processMessage(message map[string]interface{}) bool {
	log.Printf("Received message: %+v", message)
	time.Sleep(1 * time.Second)
	log.Printf("Message processed")
	return true
}
