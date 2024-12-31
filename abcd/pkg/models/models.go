package models

type Order struct {
	Stock     string  `json:"stock"`
	Type      string  `json:"type"`
	Quantity int     `json:"quantity"`
	Price     float64 `json:"price"`
}
