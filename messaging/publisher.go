package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EmailQueueName = "emails"

	DefaultExchange = ""
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPublisher(rabbitmqURL string) (*Publisher, error) {
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

	return &Publisher{
		conn:    conn,
		channel: channel,
	}, nil
}

func (p *Publisher) PublishEmail(ctx context.Context, msg *EmailMessage) error {
	body, err := msg.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize email message: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		DefaultExchange, // exchange
		EmailQueueName,  // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	slog.Info("Published email to queue",
		"to", msg.To,
		"template", msg.Template,
	)

	return nil
}

func (p *Publisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			return err
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func PublishEmailSimple(rabbitmqURL, to, template string, data map[string]interface{}) error {
	publisher, err := NewPublisher(rabbitmqURL)
	if err != nil {
		return err
	}
	defer publisher.Close()

	msg := NewEmailMessage(to, template, data)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return publisher.PublishEmail(ctx, msg)
}
