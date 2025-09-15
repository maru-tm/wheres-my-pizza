package service

import (
	"context"
	"encoding/json"
	"fmt"

	"restaurant-system/internal/notification/infrastructure/rmq"
	"restaurant-system/pkg/logger"
)

type NotificationConsumer interface {
	ConsumeStatusUpdates(ctx context.Context, handler func(update *rmq.StatusUpdateMessage) error) error
}

type NotificationService struct {
	consumer NotificationConsumer
}

func (s *NotificationService) Start(ctx context.Context) error {
	rid := ""
	if v := ctx.Value("request_id"); v != nil {
		if str, ok := v.(string); ok {
			rid = str
		}
	}
	if rid == "" {
		rid = fmt.Sprintf("notif-%d", ctx.Value("startup_id"))
	}

	logger.Log(logger.INFO, "notification-subscriber", "service_started", "notification service started", rid, nil, nil)

	return s.consumer.ConsumeStatusUpdates(ctx, func(update *rmq.StatusUpdateMessage) error {
		logger.Log(logger.DEBUG, "notification-subscriber", "notification_received", "received status update", rid,
			map[string]interface{}{
				"order_number": update.OrderNumber,
				"old_status":   update.OldStatus,
				"new_status":   update.NewStatus,
				"changed_by":   update.ChangedBy,
			}, nil)

		estimated := ""
		if !update.EstimatedCompletion.IsZero() {
			estimated = fmt.Sprintf(", estimated completion: %s", update.EstimatedCompletion.Format("15:04:05"))
		}

		fmt.Printf("Notification for order %s: Status changed from '%s' to '%s' by %s%s\n",
			update.OrderNumber, update.OldStatus, update.NewStatus, update.ChangedBy, estimated)

		if jsonData, err := json.MarshalIndent(update, "", "  "); err == nil {
			fmt.Printf("Full message: %s\n", string(jsonData))
		}

		return nil
	})
}
