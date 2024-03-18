package models

import "time"

// Password описывает модель логина и пароля от абстрактной учётной записи, а так же хранит мета-информацию.
type Password struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	URL       string    `json:"url"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiredAt time.Time `json:"expired_at"`
}
