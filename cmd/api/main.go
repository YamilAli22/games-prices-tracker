package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/YamilAli22/content_tracker/internal/db"
	"github.com/YamilAli22/content_tracker/internal/debug"
	"github.com/YamilAli22/content_tracker/internal/games"
	"github.com/YamilAli22/content_tracker/internal/middleware"
	"github.com/YamilAli22/content_tracker/internal/notifications"
	"github.com/YamilAli22/content_tracker/internal/users"
	"github.com/YamilAli22/content_tracker/internal/worker"
)

var PORT string
var IS_DEV bool

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Couldn't load .env file!!")
	}
	PORT = os.Getenv("PORT")
	IS_DEV = os.Getenv("ENV") == "dev"
}

type Server struct {
	Router *chi.Mux
	DB *pgxpool.Pool
}

func newServer(db *pgxpool.Pool, w *worker.Worker) *Server {
	server := &Server{
		Router: chi.NewRouter(), // NewRouter() returns a new Mux
		DB: db,	
	}
	
	if IS_DEV {
		server.Router.Use(cors.Handler(cors.Options{
       	AllowedOrigins:   []string{"http://localhost:5500"},
       	AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
       	AllowedHeaders:   []string{"Authorization", "Content-Type"},
		}))
	}

	usersHandler := &users.Handler{Pool: server.DB}
	server.Router.Post("/users", usersHandler.HandleCreateUser)
	server.Router.Post("/users/login", usersHandler.HandleUserLogin)
	server.Router.Get("/users", usersHandler.HandleGetUsers)
	server.Router.Get("/user/search", usersHandler.HandleGetUserByEmail) // now only searchs by email, but it can be generalized and add more filters
	server.Router.Get("/user/{id}", usersHandler.HandleGetUserByID)
	// GAMES ROUTES AND NOTI ROUTES
	gamesHandler := &games.GameHandler{Pool: server.DB}
	notiHandler := &notifications.NotiHandler{Pool: server.DB}
	server.Router.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Post("/games", gamesHandler.HandleAddGame)
		r.Get("/games/user-games", gamesHandler.HandleGetUserGames)
		r.Get("/games/search", gamesHandler.HandleGameRequest)
		r.Patch("/games/update", gamesHandler.HandleUpdateGame)
		r.Delete("/games/{steamAppID}", gamesHandler.HandleDeleteGame)
		// notification route
		r.Get("/notifications", notiHandler.HandleGetNotification)
		r.Patch("/notification/read", notiHandler.HandleMarkRead)
	})

	debugHandler := &debug.Handler{Worker: w}
	if IS_DEV {
		server.Router.Group(func(r chi.Router) {
		r.Post("/debug/check-prices", debugHandler.HandleCheckPrices)
		})
	}
	return server
}

func main() {
	conn, err := db.Connect()
	if err != nil {
		log.Fatalln(err)
	}

	defer conn.Close()

	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Successfully Connected to Postgresql")
	
	priceWorker := worker.NewWorker(conn)
	go priceWorker.StartWorker(context.Background())
	
	server := newServer(conn, priceWorker)
	
	fmt.Println("server running on port", PORT)
	log.Fatalln(http.ListenAndServe(PORT, server.Router))
}
