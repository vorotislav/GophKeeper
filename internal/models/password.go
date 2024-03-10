package models

import "time"

type Password struct {
	ID             int
	Title          string
	Login          string
	Password       string
	URL            string
	Note           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExpirationDate time.Time
}
