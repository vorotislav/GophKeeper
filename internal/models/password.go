package models

import "time"

// Password описывает модель логина и пароля от абстрактной учётной записи, а так же хранит мета-информацию.
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
