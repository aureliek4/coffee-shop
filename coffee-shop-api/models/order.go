package models

import "time"

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusPreparing OrderStatus = "preparing"
	StatusReady     OrderStatus = "ready"
	StatusPickedUp  OrderStatus = "picked-up"
	StatusCancelled OrderStatus = "cancelled"
)

// Order représente une commande
type Order struct {
	ID           string      `json:"id"`
	DrinkID      string      `json:"drink_id"`
	DrinkName    string      `json:"drink_name"`
	Size         string      `json:"size"`   // small, medium, large
	Extras       []string    `json:"extras"` // milk, sugar, cream, caramel
	CustomerName string      `json:"customer_name"`
	Status       OrderStatus `json:"status"`
	TotalPrice   float64     `json:"total_price"`
	OrderedAt    time.Time   `json:"ordered_at"`
}
