package models

import "time"

// Media описывает модель, которая хранит в себе любой файл в двоичном виде, а так же некоторую мета-информацию.
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
