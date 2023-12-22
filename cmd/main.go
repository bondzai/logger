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

const (
	maxBufferSize = 1000
)

var (
	messageBuffer       []interface{}
	shutdownGracePeriod = 10 * time.Second
)

func init() {
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

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

		// Allow a grace period for clean shutdown
		shutdownDeadline := time.Now().Add(shutdownGracePeriod)
		rabbitMQConsumer.Stop()

		// Process any remaining messages in the buffer before exiting
		processAndInsertBufferedMessages()

		select {
		case <-time.After(time.Until(shutdownDeadline)):
			log.Println("Shutdown grace period expired. Exiting.")
		}
	}()

	wg.Add(1)

	rabbitMQConsumer.Start(processMessage)

	log.Printf("Consumer started. To exit, press CTRL+C")
	wg.Wait()
}

func processMessage(message map[string]interface{}) bool {
	log.Printf("Received message: %+v", message)

	// Append the message to the buffer
	messageBuffer = append(messageBuffer, message)

	// Check if the buffer size has reached the maximum
	if len(messageBuffer) >= maxBufferSize {
		// If the buffer is full, process and insert the messages
		processAndInsertBufferedMessages()
	}

	return true
}

func processAndInsertBufferedMessages() {
	// Check if there are any messages in the buffer
	if len(messageBuffer) > 0 {
		// Perform the bulk write operation
		err := mongodb.InsertDocuments("logs", messageBuffer)
		if err != nil {
			log.Printf("Failed to insert documents into MongoDB: %v", err)
		} else {
			log.Printf("Bulk write successful for %d documents", len(messageBuffer))
		}

		// Clear the buffer after processing
		messageBuffer = messageBuffer[:0]
	}
}
