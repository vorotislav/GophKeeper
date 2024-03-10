package models

import "time"

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
