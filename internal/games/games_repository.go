package games

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func StoreGameInDB(ctx context.Context, conn *pgxpool.Pool, input Game) (GameResponse, error) {
	var SteamAppId int 
	var Name string
	var CurrentPrice float64
	var TargetPrice float64
	var IsFree bool

	query := `INSERT INTO games (user_id, steam_app_id, name, current_price, target_price, is_free) 
			  VALUES ($1, $2, $3, $4, $5, $6) 
			  RETURNING steam_app_id, name, current_price, target_price, is_free`
	
	err := conn.QueryRow(ctx, query, input.UserID, input.SteamAppID, input.Name, input.CurrentPrice, input.TargetPrice, input.IsFree).Scan(&SteamAppId, &Name, &CurrentPrice, &TargetPrice, &IsFree)
	if err != nil {
		return GameResponse{}, fmt.Errorf("error during db query: %w", err)
	}
	response := GameResponse{
		SteamAppID: SteamAppId,
		Name: Name,
		CurrentPrice: CurrentPrice,
		TargetPrice: TargetPrice,
		IsFree: IsFree,
	}
	return response, nil
}

func GetAllGames(ctx context.Context, conn *pgxpool.Pool) ([]*GameResponse, error) {
	rows, _ := conn.Query(ctx, `SELECT steam_app_id, name, current_price, target_price, is_free FROM games`)
	games, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*GameResponse, error) {
		game := new(GameResponse)
		err := row.Scan(&game.SteamAppID, &game.Name, &game.CurrentPrice, &game.TargetPrice, &game.IsFree)
		return game, err
	})
	return games, err
}

func GetUserGames(ctx context.Context, conn *pgxpool.Pool, userID uuid.UUID) ([]*GameResponse, error) {
	rows, err := conn.Query(ctx, `SELECT steam_app_id, name, COALESCE(current_price, 0), COALESCE(target_price, 0), is_free FROM games WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	games, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*GameResponse, error) {
		game := new(GameResponse)
		err := row.Scan(&game.SteamAppID, &game.Name, &game.CurrentPrice, &game.TargetPrice, &game.IsFree)
		return game, err
	})
	return games, err
}

func GetGameByID(ctx context.Context, conn *pgxpool.Pool, gameID string) (GameResponse, error) {
	var game GameResponse
	query := `SELECT * FROM games WHERE steam_app_id=$1`
	err := conn.QueryRow(
		ctx, query, gameID).
		Scan(&game.SteamAppID, &game.Name, &game.CurrentPrice, 
		&game.TargetPrice, &game.IsFree)

	if err != nil {
		return GameResponse{}, nil
	}
	return game, nil
}

func UpdateGameData(ctx context.Context, conn *pgxpool.Pool, targetPrice float64, gameID int, userID uuid.UUID) (float64, error) {
	var targetP float64
	
	query := `UPDATE games SET target_price = $1 WHERE steam_app_id = $2 AND user_id = $3 RETURNING target_price`
	err := conn.QueryRow(ctx, query, targetPrice, gameID, userID).Scan(&targetP)
	if err != nil {
		return 0, err
	}
	return targetP, nil 
}

var ErrGameNotFound = errors.New("game not found")

func DeleteGame(ctx context.Context, conn *pgxpool.Pool, userID uuid.UUID, steamAppID int) (error) {
	tag, err := conn.Exec(ctx,
		`DELETE FROM games WHERE user_id = $1 AND steam_app_id = $2`,
		userID, steamAppID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrGameNotFound
	}
	return nil
}
