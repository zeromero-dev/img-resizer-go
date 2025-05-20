package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"img-resizer/internal/config"
	"img-resizer/internal/models"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ implements the Queue interface for RabbitMQ
type RabbitMQ struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	queueName    string
	exchangeName string
	routingKey   string
}

// NewRabbitMQ creates a new RabbitMQ instance
func NewRabbitMQ(cfg *config.Config) (*RabbitMQ, error) {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create a channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare an exchange
	err = channel.ExchangeDeclare(
		cfg.RabbitMQ.ExchangeName, // name
		"direct",                  // type
		true,                      // durable
		false,                     // auto-deleted
		false,                     // internal
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	// Declare a queue
	_, err = channel.QueueDeclare(
		cfg.RabbitMQ.QueueName, // name
		true,                   // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	// Bind the queue to the exchange
	err = channel.QueueBind(
		cfg.RabbitMQ.QueueName,    // queue name
		cfg.RabbitMQ.RoutingKey,   // routing key
		cfg.RabbitMQ.ExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind a queue: %w", err)
	}

	return &RabbitMQ{
		conn:         conn,
		channel:      channel,
		queueName:    cfg.RabbitMQ.QueueName,
		exchangeName: cfg.RabbitMQ.ExchangeName,
		routingKey:   cfg.RabbitMQ.RoutingKey,
	}, nil
}

// func (r *RabbitMQ) PublishTask(task *models.ImageProcessingTask) error {
// 	// Convert task to JSON
// 	body, err := json.Marshal(task)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal task: %w", err)
// 	}

// 	// Publish the message
// 	err = r.channel.Publish(
// 		r.exchangeName, // exchange
// 		r.routingKey,   // routing key
// 		false,          // mandatory
// 		false,          // immediate
// 		amqp.Publishing{
// 			DeliveryMode: amqp.Persistent,
// 			ContentType:  "application/json",
// 			Body:         body,
// 		},
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to publish a message: %w", err)
// 	}

// 	return nil
// }

// PublishTask publishes a task to the queue
func (r *RabbitMQ) PublishTask(task *models.ImageProcessingTask) error {
	// Convert task to JSON
	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Publish the message
	err = r.channel.PublishWithContext(
		ctx,
		r.exchangeName, // exchange
		r.routingKey,   // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

// ConsumeTask consumes a task from the queue
func (r *RabbitMQ) ConsumeTask(handler func(task *models.ImageProcessingTask) error) error {
	// Start consuming messages
	msgs, err := r.channel.Consume(
		r.queueName, // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	// Process messages
	for msg := range msgs {
		// Parse the message
		var task models.ImageProcessingTask
		err := json.Unmarshal(msg.Body, &task)
		if err != nil {
			// Reject the message
			msg.Reject(false)
			continue
		}

		// Process the task
		err = handler(&task)
		if err != nil {
			// Reject the message and requeue it
			msg.Reject(true)
			continue
		}

		// Acknowledge the message
		msg.Ack(false)
	}

	return nil
}

// Close closes the connection to RabbitMQ
func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
	return nil
}
