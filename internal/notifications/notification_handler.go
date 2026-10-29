package notifications

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/YamilAli22/content_tracker/internal/games"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotiHandler struct {
	Pool *pgxpool.Pool
}

func (h *NotiHandler) HandleGetNotification(w http.ResponseWriter, r *http.Request) {
	userID, err := games.GetTokenClaims(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid Token"))
		return 
	}

	msgs, err := GetUserNotification(r.Context(), h.Pool, userID)
	if err != nil {
		log.Printf("GET NOTIFICATION ERROR: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error geting notification"))
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}

func (h *NotiHandler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	userID, err := games.GetTokenClaims(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = UpdateReadNoti(r.Context(), h.Pool, userID)
	if err != nil {
		log.Printf("MarkRead: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
