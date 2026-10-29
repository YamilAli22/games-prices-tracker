package users

import "github.com/google/uuid"

type User struct {
	// lo que está entre ` ` son structs tags, le dicen a go como serializar/deserializar el struct, tambien para validar
	Id uuid.UUID `json:"id"`
	Email string `json:"email"`
	Hash string `json:"hash"` 
}

type UserRequestBody struct {
	Email string  `json:"email"`
	Password string  `json:"password"`
}

type UserResponse struct {
	Id uuid.UUID  `json:"id"`
	Email string  `json:"email"`
}



