package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
)

const (
	QueueName     = "log"
	ConnectionURL = "amqp://guest:guest@localhost:5672/"
)

func main() {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(ConnectionURL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	// Declare the queue to consume
	q, err := ch.QueueDeclare(
		QueueName, // Queue name
		true,      // Durable
		false,     // Delete when unused
		false,     // Exclusive
		false,     // No-wait
		nil,       // Arguments
	)
	if err != nil {
		log.Fatal("Failed to declare a queue:", err)
	}

	// Consume messages from the queue
	msgs, err := ch.Consume(
		q.Name, // Queue
		"",     // Consumer
		true,   // Auto-acknowledge
		false,  // Exclusive
		false,  // No-local
		false,  // No-wait
		nil,    // Args
	)
	if err != nil {
		log.Fatal("Failed to register a consumer:", err)
	}

	// Process messages
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var message map[string]interface{}

			err := json.Unmarshal(d.Body, &message)
			if err != nil {
				log.Println("Error decoding message:", err)
				continue
			}

			processMessage(message)
		}
	}()

	log.Printf("Consumer started. To exit, press CTRL+C")
	<-forever
}

func processMessage(message map[string]interface{}) {
	log.Printf("Received message: %+v", message)

	// Add your processing logic here
	// For example, you can save the message to a database, trigger an action, etc.
	time.Sleep(2 * time.Second) // Simulate processing time
	log.Printf("Message processed")
}
