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

type Stock struct {
	TickerID int
	Quantity int
	LQ       int
	Price    int
}

type Config struct {
    REDIS_URL1   string
    REDIS_URL2   string
    RABBITMQ_URL string
	SUBSCRIBER_QUEUE string
	PUBLISHER_EXCHANGE string
	PUBLISHER_ROUTING_KEY string
}


