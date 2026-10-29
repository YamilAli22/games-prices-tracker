package worker

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/YamilAli22/content_tracker/internal/games"
	"github.com/YamilAli22/content_tracker/internal/notifications"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Worker struct {
	Pool *pgxpool.Pool
}

func NewWorker(db *pgxpool.Pool) *Worker {
	return &Worker{
		Pool: db,
	}
}

func (w *Worker) StartWorker(ctx context.Context) {
	ticker := time.NewTicker(time.Hour) // change later
	defer ticker.Stop()
	
	log.Println("Worker started!!")
	
	for {
		select {
		case <-ticker.C:
			log.Println("Ticker - worker checking prices ... ")
			// check the prices in steam and then compare to the games in db 
			if err := w.CheckAllPrices(ctx); err != nil {
				log.Printf("Error en en checkAllPrices: %v", err)
			} 
		case <-ctx.Done():
			log.Println("nada para chequear")
			return
		}
	}
}

func (w *Worker) CheckAllPrices(ctx context.Context) (error) {
	query := "SELECT DISTINCT steam_app_id FROM games"
	rows, err := w.Pool.Query(ctx, query)
	if err != nil {
		return err 
	}
	appIDs, err := pgx.CollectRows(rows, pgx.RowTo[int])
	log.Printf("appIDs encontrados: %v\n", appIDs)
	if err != nil {
		return err
	}
	
	var steam_games_data []games.SteamAppData
	for _, id := range appIDs {
		steam_resp, err := games.CallSteamSearchByID(ctx, id)
		if err != nil {
			log.Printf("Error llamando a la api para app_id %d: %v", id, err) 
			continue
		}
		if !steam_resp.Success || steam_resp.Data == nil {
			log.Printf("Steam no devolvió datos para app_id: %d", id)
			continue
		}
		if steam_resp.Data.IsFree || steam_resp.Data.PriceOverview == nil {
			log.Printf("EL juego con app_id %d es gratis, no hay price overview\n", id)
			continue
		}
		log.Printf("Steam respondio para %d: precio=%f", id, steam_resp.Data.PriceOverview.Final)
		steam_games_data = append(steam_games_data, *steam_resp.Data)
	}
	log.Println("HOLA ????????")
	// var user_games_data []games.GameResponse
	query2 := "SELECT id, user_id, steam_app_id, name, COALESCE(current_price, 0), COALESCE(target_price, 0), is_free, created_at FROM games"
	g_rows, err := w.Pool.Query(ctx, query2)
	if err != nil {
		log.Printf("ERROR QUERY: %v", err)
		return err
	}
	user_games_data, err := pgx.CollectRows(g_rows, pgx.RowToStructByPos[games.Game])
	if err != nil {
		log.Printf("ERROR COLLECT ROW: %v", err)
		return err
	}
	
	log.Printf("STEAM GAMES: %v \n\n USER GAMES: %v", steam_games_data, user_games_data)
	for _, steam_game := range steam_games_data {
		if steam_game.PriceOverview == nil {
			continue
		}
		price := steam_game.PriceOverview.Final

		for _, user_game := range user_games_data {
			if steam_game.ID != user_game.SteamAppID {
				continue
			}
			if math.Abs(price - user_game.CurrentPrice) < 0.005 {
				continue // el precio no bajo, nada que avisar
			}
			log.Println("Entrando al chequeo de precio...")
			if price <= user_game.TargetPrice {
				log.Println("El precio bajó del target...")
				msg := fmt.Sprintf("%s bajó a %.2f (tu objetivo: %.2f)", user_game.Name, price, user_game.TargetPrice)
				if err := w.createNotification(ctx, user_game.UserID, user_game.ID, msg); err != nil {
					log.Printf("Error creando notificación para user %s: %v", user_game.UserID, err)
				} else {
					log.Printf("Notificación creada para user %s: %s", user_game.UserID, msg)
				}
			}
		}
		// después de procesar a todos los usuarios de ese juego, guardar el precio nuevo
		tag, err := w.Pool.Exec(ctx,
			`UPDATE games SET current_price = $1 WHERE steam_app_id = $2`,
			price, steam_game.ID)
		if err != nil {
			log.Printf("Error actualizando current_price de %d: %v", steam_game.ID, err)
		} else {
			log.Printf("UPDATE %d (%s) -> %.2f, filas: %d", steam_game.ID, steam_game.Name, price, tag.RowsAffected())
		}
	}
	return nil
}

func (w *Worker) createNotification(ctx context.Context, userID, gameID uuid.UUID, message string) error {
	noti := notifications.Notification{
		UserID: userID,
		GameID: gameID,
		Message: message,
		Read: false,
		CreatedAt: time.Now(),
	}

	return notifications.AddNotification(ctx, w.Pool, &noti)
}

