package notifications

import (
	"time" 
	"github.com/google/uuid"
)

type Notification struct {
	UserID uuid.UUID `json:"user_id"`
	GameID uuid.UUID `json:"game_id"`
	Message string `json:"message"` 
	Read bool `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

// para la api
type NotificationResponse struct {
	Message string `json:"message"`
	Read bool `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}
