package models

import "time"

// Note описывает модель, которая содержит в себе текстовую заметку и мета-информацию.
type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Text      string    `json:"text"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
