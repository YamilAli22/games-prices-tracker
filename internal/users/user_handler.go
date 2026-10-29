package users

/*
handler.go for handles the http requests and serving the responses, interacting with the database if necessary
*/

import (
	"encoding/json"
	"net/http"

	"github.com/YamilAli22/content_tracker/internal/auth"
	"github.com/go-chi/chi/v5"

	//	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	Pool *pgxpool.Pool
}

func (h *Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	newUser := new(UserRequestBody)
	err := json.NewDecoder(r.Body).Decode(newUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please enter a valid input"))
		return
	}
	newUser.Password, err = auth.HashPassword(newUser.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong in the server during the user creation"))
		return
	}
	response, err := CreateUser(r.Context(), h.Pool, newUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong in the server"))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	user := new(UserRequestBody)
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please enter a valid input"))
		return
	}
	user_check, err := GetUserByEmailWithHash(r.Context(), h.Pool, user.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid credentials"))
		return
	}
	if !auth.CheckPasswordHash(user.Password, user_check.Hash) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid credentials"))
		return 
	}
	JWT, err := auth.CreateJWT(user_check.Id, user_check.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong in the server"))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(JWT)
}

func (h *Handler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := GetUsers(r.Context(), h.Pool)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func (h *Handler) HandleGetUserByEmail(w http.ResponseWriter, r *http.Request) {
	user_email := r.URL.Query().Get("email")
	user, err := GetUserByEmail(r.Context(), h.Pool, user_email)
	if err != nil {
 		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) HandleGetUserByID(w http.ResponseWriter, r *http.Request) {	
	id := chi.URLParam(r, "id")
	user, err := GetUserById(r.Context(), h.Pool, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return 
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}


