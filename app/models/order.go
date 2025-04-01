package models

import "time"

type Order struct {
	ID        string    `json:"id"`
	Product   string    `json:"product"`
	Quantity  int       `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
}
