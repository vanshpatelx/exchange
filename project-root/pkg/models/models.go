package models

type ExchangeMsg struct {
	Order Order `json:"order,omitempty"`
	Task  int   `json:"task"`
}

type Order struct {
	Id       int `json:"id"`
	Ticker   int `json:"ticker"`
	Type     int `json:"type"`
	Quantity int `json:"quantity"`
	Price    int `json:"price"`
	User     int `json:"user"`
}
