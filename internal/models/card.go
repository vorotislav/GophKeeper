package models

import "time"

// Card описывает модель банковской карты.
type Card struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Number    string    `json:"card_number"`
	CVC       string    `json:"cvc"`
	ExpMonth  int       `json:"exp_month_at"`
	ExpYear   int       `json:"exp_year_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
