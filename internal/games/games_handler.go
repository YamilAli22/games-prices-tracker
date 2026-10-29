package games

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GameHandler struct {
	Pool *pgxpool.Pool
	SteamData SteamSearchResponse
}

func (h *GameHandler) HandleGameRequest(w http.ResponseWriter, r *http.Request) {
	game := r.URL.Query().Get("term")
	steamResp, err := CallSteamSearch(r.Context(), game)
	if err != nil {
		log.Printf("CallSteamSerch: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return
	}
	json.NewEncoder(w).Encode(steamResp)
}

func (h *GameHandler) HandleAddGame(w http.ResponseWriter, r *http.Request) {
	userID, err := GetTokenClaims(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return 
	}

	gameReq := new(GameRequest)
	err = json.NewDecoder(r.Body).Decode(gameReq)
	
	if gameReq.TargetPrice < 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
		return 
	}
	
	if gameReq.SteamAppID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
		return 
	}
	
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid Request"))
		return 
	}

	steamResp, err := CallSteamSearchByID(r.Context(), gameReq.SteamAppID)
	if err != nil {
    	log.Printf("CallSteamSearchByID: %v", err)
    	w.WriteHeader(http.StatusInternalServerError)
    	w.Write([]byte("Something went wrong"))
    	return
	}

	gameDetails, err := GameDetailsLogic(steamResp, gameReq.SteamAppID, gameReq.TargetPrice)
	if err != nil {
 	   	log.Printf("CallSteamEndpointByID: %v", err)
    	w.WriteHeader(http.StatusInternalServerError)
    	w.Write([]byte("Something went wrong"))
   	 	return
	}

	if gameDetails.CurrentPrice <= gameDetails.TargetPrice {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("El juego ya está por debajo de tu precio elegido!"))
		return 
	}
		
	// store in DB
	// TODO: MODULARIZAR ESTO (QUIZAS?)
	inputGame := Game{
		UserID: userID,
		SteamAppID: gameReq.SteamAppID,
		Name: gameDetails.Name,
		CurrentPrice: gameDetails.CurrentPrice,
		TargetPrice: gameDetails.TargetPrice,
		IsFree: gameDetails.IsFree,
	}
	
	DBResponse, err := StoreGameInDB(r.Context(), h.Pool, inputGame)  
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Ya tenés este juego en tu lista"))
			return
		}
		log.Printf("StoreGameInDB: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(DBResponse)
}

func (h *GameHandler) HandleGetUserGames(w http.ResponseWriter, r *http.Request) {
	userID, err := GetTokenClaims(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return 
	}
	response, err := GetUserGames(r.Context(), h.Pool, userID)
	if err != nil {
		log.Printf("GetUserGames: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return 
	}
	json.NewEncoder(w).Encode(response)
}

func (h *GameHandler) HandleUpdateGame(w http.ResponseWriter, r *http.Request) {
	gameReq := new(GameRequest)
	err := json.NewDecoder(r.Body).Decode(gameReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid Request"))
		return 
	}
	log.Println("antes del token claims")
	userID, err := GetTokenClaims(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return 
	}
	log.Println("pase el token claims")
	response, err := UpdateGameData(r.Context(), h.Pool, gameReq.TargetPrice, gameReq.SteamAppID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return 
	}
	json.NewEncoder(w).Encode(response)
}

func (h *GameHandler) HandleDeleteGame(w http.ResponseWriter, r *http.Request) {
	userID, err := GetTokenClaims(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Token Invalido"))
		return 
	}

	steamAppID, err := strconv.Atoi(chi.URLParam(r, "steamAppID"))
	if err != nil || steamAppID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid game id"))
		return
	}

	err = DeleteGame(r.Context(), h.Pool, userID, steamAppID)
	if errors.Is(err, ErrGameNotFound) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Ese juego no está en tu lista"))
		return
	}
	if err != nil {
		log.Printf("DeleteGame: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


