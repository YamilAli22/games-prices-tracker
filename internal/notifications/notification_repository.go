package notifications

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AddNotification(ctx context.Context, pool *pgxpool.Pool, noti *Notification) error {
	query := `INSERT INTO notifications (user_id, game_id, message, read, created_at)
	          VALUES ($1, $2, $3, $4, $5)`
	_, err := pool.Exec(ctx, query, noti.UserID, noti.GameID, noti.Message, noti.Read, noti.CreatedAt)
	return err
}

func GetUserNotification(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]NotificationResponse, error) {
	//var msg NotificationResponse
	rows, err := pool.Query(ctx, `SELECT COALESCE(message, ' '), read, created_at FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[NotificationResponse])
}

func UpdateReadNoti(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	_, err := pool.Exec(ctx,`UPDATE notifications SET read = true WHERE user_id = $1 AND read = false`, userID)
	if err != nil {
		return err 
	}

	return nil 
}
