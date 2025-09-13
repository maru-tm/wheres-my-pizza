package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"restaurant-system/config"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  config.RabbitMQConfig
}

func New(cfg config.RabbitMQConfig) (*RabbitMQ, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Configure channel
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
		config:  cfg,
	}, nil
}

func (r *RabbitMQ) Channel() *amqp091.Channel {
	return r.channel
}

func (r *RabbitMQ) Reconnect() error {
	if r.conn != nil && !r.conn.IsClosed() {
		return nil
	}

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		r.config.User,
		r.config.Password,
		r.config.Host,
		r.config.Port,
	)

	conn, err := amqp091.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Close old connection
	if r.conn != nil {
		r.conn.Close()
	}

	r.conn = conn
	r.channel = ch

	return nil
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *RabbitMQ) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	if r.conn.IsClosed() {
		if err := r.Reconnect(); err != nil {
			return fmt.Errorf("failed to reconnect: %w", err)
		}
	}

	return r.channel.PublishWithContext(ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
			Timestamp:    time.Now(),
		})
}

func (r *RabbitMQ) Consume(ctx context.Context, queue, consumer string, autoAck bool) (<-chan amqp091.Delivery, error) {
	if r.conn.IsClosed() {
		if err := r.Reconnect(); err != nil {
			return nil, fmt.Errorf("failed to reconnect: %w", err)
		}
	}

	return r.channel.Consume(
		queue,
		consumer,
		autoAck,
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
}
