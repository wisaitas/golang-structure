package getusers

import "github.com/google/uuid"

type Request struct{}

type Response struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Age  int       `json:"age"`
}
