package models

import "time"

type Note struct {
	ID        int
	Title     string
	Text      string
	ExpiredAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
