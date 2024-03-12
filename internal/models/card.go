package models

import "time"

// Card описывает модель банковской карты.
type Card struct {
	ID        int
	Name      string
	Number    string
	CVC       string
	ExpMonth  int
	ExpYear   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
