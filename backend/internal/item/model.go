package item

import "time"

type Item struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
