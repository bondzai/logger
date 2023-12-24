package rabbitmq

import (
	"context"
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

func (c *Consumer) Start(ctx context.Context, handler MessageHandler, wg *sync.WaitGroup) error {
	defer wg.Done()

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
		return err
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Received cancellation signal. Stopping consumer...")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Channel closed. Stopping consumer...")
				return nil
			}

			message := make(map[string]interface{})
			err := json.Unmarshal(msg.Body, &message)
			if err != nil {
				log.Printf("Failed to unmarshal message body: %v", err)
				continue
			}

			if !handler(message) {
				log.Println("Message processing failed. Stopping consumer...")
				return nil
			}
		}
	}
}

func (c *Consumer) Stop() {
	log.Println("Closing RabbitMQ channel and connection...")
	_ = c.channel.Close()
	_ = c.conn.Close()
}
