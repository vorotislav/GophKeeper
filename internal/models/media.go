package models

import "time"

// Media описывает модель, которая хранит в себе любой файл в двоичном виде, а так же некоторую мета-информацию.
type Media struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      []byte    `json:"body"`
	MediaType string    `json:"media_type"`
	Note      string    `json:"note"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
