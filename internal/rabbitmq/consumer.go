// rabbitmq_consumer.go
package rabbitmq

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/streadway/amqp"
)

type MessageHandler func(message map[string]interface{}) bool

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewConsumer(amqpURI, queueName string) (*Consumer, error) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn:    conn,
		channel: channel,
		queue:   queue,
	}, nil
}

func (c *Consumer) Start(handler MessageHandler) {
	msgs, err := c.channel.Consume(
		c.queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	var wg sync.WaitGroup

	go func() {
		defer wg.Done()
		for msg := range msgs {
			message := make(map[string]interface{})

			err := json.Unmarshal(msg.Body, &message)
			if err != nil {
				log.Printf("Failed to unmarshal message body: %v", err)
				continue
			}

			if !handler(message) {
				log.Println("Message processing failed. Stopping consumer...")
				c.Stop()
				break
			}
		}
	}()

	wg.Add(1)
}

func (c *Consumer) Stop() {
	log.Println("Closing RabbitMQ channel and connection...")
	err := c.channel.Close()
	if err != nil {
		log.Printf("Error closing channel: %v", err)
	}

	err = c.conn.Close()
	if err != nil {
		log.Printf("Error closing connection: %v", err)
	}
}
