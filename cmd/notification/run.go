package notification

import (
	"context"
	"fmt"
	"time"

	"restaurant-system/internal/notification/infrastructure/rmq"
	"restaurant-system/pkg/logger"
	"restaurant-system/pkg/rabbitmq"
)

func Run(ctx context.Context, rabbitmqClient *rabbitmq.RabbitMQ, rid string) {
	// Create notification consumer
	consumer, err := rmq.NewNotificationConsumer(rabbitmqClient)
	if err != nil {
		logger.Log(logger.ERROR, "notification-subscriber", "consumer_init_failed", "failed to initialize consumer", rid, nil, err)
		return
	}

	logger.Log(logger.INFO, "notification-subscriber", "service_started", "Notification subscriber started", rid, nil, nil)

	// Consume messages
	messages, err := consumer.Consume(ctx)
	if err != nil {
		logger.Log(logger.ERROR, "notification-subscriber", "consume_failed", "failed to start consuming messages", rid, nil, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			logger.Log(logger.INFO, "notification-subscriber", "shutdown_initiated", "received termination signal", rid, nil, nil)
			return
		case msg, ok := <-messages:
			if !ok {
				return
			}

			requestID := fmt.Sprintf("notif-%d", time.Now().UnixNano())
			logger.Log(logger.DEBUG, "notification-subscriber", "notification_received", "received status update", requestID,
				map[string]interface{}{
					"order_number": msg.OrderNumber,
					"new_status":   msg.NewStatus,
				}, nil)

			// Print human-readable notification
			fmt.Printf("Notification for order %s: Status changed from '%s' to '%s' by %s. Estimated completion: %s\n",
				msg.OrderNumber, msg.OldStatus, msg.NewStatus, msg.ChangedBy, msg.EstimatedCompletion.Format(time.RFC3339))

			// Ack the message
			if err := consumer.Ack(msg.DeliveryTag); err != nil {
				logger.Log(logger.ERROR, "notification-subscriber", "ack_failed", "failed to ack message", requestID, nil, err)
			}
		}
	}
}
