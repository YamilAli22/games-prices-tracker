package users

/*
repository.go is responsible for managing database operations, implementing necessary operations, and interacting with the database.
*/

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUser(ctx context.Context, conn *pgxpool.Pool, user *UserRequestBody) (UserResponse, error) {
	var id uuid.UUID
	var email string
	query := `INSERT INTO users (email, hash) VALUES ($1, $2) RETURNING id, email`
	err := conn.QueryRow(ctx, query, user.Email, user.Password).Scan(&id, &email)
	if err != nil {
		response := UserResponse{
			Id: id,
		}
		return response, err 
	}
	response := UserResponse{
		Id: id,
		Email: email,
	}
	return response, err
}


// esta funcion devuelve los usuarios traidos de la db en un formato json 
func GetUsers(ctx context.Context, conn *pgxpool.Pool) ([]*UserResponse, error) {
    rows, _ := conn.Query(ctx, "SELECT id, email FROM users")
    users, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*UserResponse, error) {
        user := new(UserResponse)
        err := row.Scan(&user.Id, &user.Email)
        return user, err
    })
    return users, err
}

// GetUserByEmail returns User (internal use, contains hash, this function is used by the login handler)
func GetUserByEmailWithHash(ctx context.Context, conn *pgxpool.Pool, email string) (User, error) {
    var user User
    query := `SELECT id, email, hash FROM users WHERE email = $1`
    err := conn.QueryRow(ctx, query, email).Scan(&user.Id, &user.Email, &user.Hash)
	if err != nil {
		return User{}, err
	}
    return user, err
}

func GetUserByEmail(ctx context.Context, conn *pgxpool.Pool, email string) (UserResponse, error) {
    var user UserResponse
    query := `SELECT id, email FROM users WHERE email = $1`
    err := conn.QueryRow(ctx, query, email).Scan(&user.Id, &user.Email)
	if err != nil {
		return UserResponse{}, err
	}
    return user, err
}


func GetUserById(ctx context.Context, conn *pgxpool.Pool, id string) (UserResponse, error) {
	var user UserResponse
	query := `SELECT id, email FROM users WHERE id = $1`
	err := conn.QueryRow(ctx, query, id).Scan(&user.Id, &user.Email)
	if err != nil {
 		return UserResponse{}, err
	}
	return user, err
}



