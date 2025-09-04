package rabbitmq

import (
	"fmt"
	"log"
	"restaurant-system/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func New(cfg config.RabbitMQConfig) (*RabbitMQ, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.User, cfg.Password, cfg.Host, cfg.Port,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("не удалось открыть канал: %w", err)
	}

	log.Println("✅ Подключение к RabbitMQ успешно")

	return &RabbitMQ{Conn: conn, Channel: ch}, nil
}

func (r *RabbitMQ) DeclareQueue(queueName string) (amqp.Queue, error) {
	q, err := r.Channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return q, fmt.Errorf("ошибка создания очереди: %w", err)
	}
	return q, nil
}

func (r *RabbitMQ) Publish(queueName string, body []byte) error {
	return r.Channel.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Conn != nil {
		r.Conn.Close()
	}
}
