package messaging

import (
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageHandler func(*EmailMessage) error

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	handler MessageHandler
}

func NewConsumer(rabbitmqURL string, handler MessageHandler) (*Consumer, error) {
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	_, err = channel.QueueDeclare(
		EmailQueueName, // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &Consumer{
		conn:    conn,
		channel: channel,
		handler: handler,
	}, nil
}

func (c *Consumer) Start() error {
	msgs, err := c.channel.Consume(
		EmailQueueName, // queue
		"",             // consumer tag
		false,          // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	slog.Info("Consumer started, waiting for messages...")

	go func() {
		for msg := range msgs {
			c.processMessage(msg)
		}
	}()

	return nil
}

func (c *Consumer) processMessage(delivery amqp.Delivery) {
	emailMsg, err := FromJSON(delivery.Body)
	if err != nil {
		slog.Error("Failed to deserialize message",
			"error", err,
			"body", string(delivery.Body),
		)

		delivery.Nack(false, false)
		return
	}

	slog.Info("Processing email message",
		"to", emailMsg.To,
		"subject", emailMsg.Subject,
	)


	err = c.handler(emailMsg)
	if err != nil {
		slog.Error("Failed to process email",
			"to", emailMsg.To,
			"error", err,
		)

		delivery.Nack(false, false)
		return
	}


	err = delivery.Ack(false)
	if err != nil {
		slog.Error("Failed to acknowledge message",
			"error", err,
		)
	}

	slog.Info("Email processed successfully",
		"to", emailMsg.To,
	)
}


func (c *Consumer) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return err
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}


func (c *Consumer) WaitForever() {
	select {}
}
