package models

type Transaction struct {
	ID        int64   `json:"id"`
	Number    string  `json:"no_trx"`
	Total     float64 `json:"total"`
	CreatedAt any     `json:"created_at"`
}
