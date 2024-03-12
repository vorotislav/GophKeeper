package models

import "time"

// Note описывает модель, которая содержит в себе текстовую заметку и мета-информацию.
type Note struct {
	ID        int
	Title     string
	Text      string
	ExpiredAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
