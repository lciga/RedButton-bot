package bot

import (
	"context"
	"time"
)

func (b *Bot) runNotifications(ctx context.Context) {
	b.sendPendingNotifications(ctx)

	ticker := time.NewTicker(b.notificationInterval)
	defer ticker.Stop()

	for {
		activation := b.nextTaskActivation(ctx)
		var activationTimer *time.Timer
		var activationChannel <-chan time.Time
		if activation != nil {
			activationTimer = time.NewTimer(notificationDelay(time.Now(), *activation))
			activationChannel = activationTimer.C
		}

		select {
		case <-ctx.Done():
			stopTimer(activationTimer)
			return
		case <-ticker.C:
			stopTimer(activationTimer)
			b.sendPendingNotifications(ctx)
		case <-activationChannel:
			b.sendPendingNotifications(ctx)
		}
	}
}

func (b *Bot) nextTaskActivation(ctx context.Context) *time.Time {
	activation, err := b.services.Notifications.GetNextStartsAt(ctx)
	if err != nil {
		b.logError(ctx, "Failed to get next task activation time", err)
		return nil
	}

	return activation
}

func notificationDelay(now, activation time.Time) time.Duration {
	delay := activation.Sub(now)
	if delay < 0 {
		return 0
	}

	return delay
}

func stopTimer(timer *time.Timer) {
	if timer == nil {
		return
	}
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

func (b *Bot) sendPendingNotifications(ctx context.Context) {
	if !b.isAvailable(time.Now()) {
		return
	}

	const batchSize = 100
	for {
		notifications, err := b.services.Notifications.GetPending(ctx, batchSize)
		if err != nil {
			b.logError(ctx, "Failed to get pending notifications", err)
			return
		}
		if len(notifications) == 0 {
			return
		}

		sent := 0
		for _, notification := range notifications {
			if err := b.sendTask(ctx, notification.TelegramUserID, notification.Task, true); err != nil {
				b.logError(
					ctx,
					"Failed to send task notification",
					err,
					"telegram_user_id", notification.TelegramUserID,
					"task_id", notification.Task.ID,
				)
				continue
			}
			if err := b.services.Notifications.MarkSent(
				ctx,
				notification.UserID,
				notification.Task.ID,
			); err != nil {
				b.logError(
					ctx,
					"Failed to mark task notification as sent",
					err,
					"telegram_user_id", notification.TelegramUserID,
					"task_id", notification.Task.ID,
				)
				continue
			}
			sent++
		}

		if len(notifications) < batchSize || sent == 0 {
			return
		}
	}
}
