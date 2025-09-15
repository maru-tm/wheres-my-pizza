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
		err := conn.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		err := ch.Close()
		if err != nil {
			return nil, err
		}
		err = conn.Close()
		if err != nil {
			return nil, err
		}
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
		err := conn.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		err := ch.Close()
		if err != nil {
			return err
		}
		err = conn.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	if r.conn != nil {
		err := r.conn.Close()
		if err != nil {
			return err
		}
	}

	r.conn = conn
	r.channel = ch

	return nil
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		err := r.channel.Close()
		if err != nil {
			return
		}
	}
	if r.conn != nil {
		err := r.conn.Close()
		if err != nil {
			return
		}
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
		false,
		false,
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
		false,
		false,
		false,
		nil,
	)
}
