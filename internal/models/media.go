package models

import "time"

type Media struct {
	ID        int
	Title     string
	Body      []byte
	MediaType string
	Note      string
	ExpiredAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
