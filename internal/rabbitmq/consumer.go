package rabbitmq

import (
	"encoding/json"
	"log"
	"math/rand"

	"github.com/streadway/amqp"
)

type Consumer struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	queue         *amqp.Queue
	errorQueue    *amqp.Queue
	consumer      <-chan amqp.Delivery
	errorProducer *amqp.Channel
	stop          chan struct{} // Added a channel to signal consumer stop
}

func NewConsumer(connectionURL, queueName, errorQueueName string) (*Consumer, error) {
	conn, err := amqp.Dial(connectionURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queueName, // Queue name
		true,      // Durable
		false,     // Delete when unused
		false,     // Exclusive
		false,     // No-wait
		nil,       // Arguments
	)
	if err != nil {
		return nil, err
	}

	errorQueue, err := ch.QueueDeclare(
		errorQueueName, // Error queue name
		true,           // Durable
		false,          // Delete when unused
		false,          // Exclusive
		false,          // No-wait
		nil,            // Arguments
	)
	if err != nil {
		return nil, err
	}

	errorProducer, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn:          conn,
		channel:       ch,
		queue:         &q,
		errorQueue:    &errorQueue,
		consumer:      nil,
		errorProducer: errorProducer,
		stop:          make(chan struct{}),
	}, nil
}

func (c *Consumer) Start(processMessage func(message map[string]interface{}) bool) {
	consumer, err := c.channel.Consume(
		c.queue.Name, // Queue
		"",           // Consumer
		true,         // Auto-acknowledge
		false,        // Exclusive
		false,        // No-local
		false,        // No-wait
		nil,          // Args
	)
	if err != nil {
		log.Fatal("Failed to register a consumer:", err)
		return
	}

	c.consumer = consumer

	go func() {
		for {
			select {
			case d, ok := <-c.consumer:
				if !ok {
					// Channel closed, exit the goroutine
					return
				}

				var message map[string]interface{}

				err := json.Unmarshal(d.Body, &message)
				if err != nil {
					log.Println("Error decoding message:", err)
					continue
				}

				success := !simulateTaskFailure()

				if processMessage(message) && success {
					d.Ack(false)
				} else {
					// Handle the case where processing failed
					// You may choose to nack, requeue, or handle the error accordingly
					log.Println("Processing failed. Moving message to the error queue.")
					// c.moveToErrorQueue(d.Body)
					d.Ack(false)
				}
			case <-c.stop:
				// Signal to stop processing messages
				return
			}
		}
	}()
}

func (c *Consumer) Stop() {
	close(c.stop)
	c.channel.Close()
	c.conn.Close()
	c.errorProducer.Close()
}

func (c *Consumer) moveToErrorQueue(messageBody []byte) {
	err := c.errorProducer.Publish(
		"",                // Exchange
		c.errorQueue.Name, // Routing key (queue name)
		false,             // Mandatory
		false,             // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		},
	)
	if err != nil {
		log.Println("Error moving message to the error queue:", err)
	}
}

// Function to simulate occasional task failure
func simulateTaskFailure() bool {
	// Simulate failure 20% of the time
	return rand.Float64() < 0.2
}
