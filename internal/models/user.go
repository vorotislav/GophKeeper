package models

import "time"

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Machine struct {
	IPAddress string `json:"ip_address"`
}

type UserMachine struct {
	User    User    `json:"user"`
	Machine Machine `json:"machine"`
}

type Session struct {
	ID                    int64 `json:"id"`
	UserID                int
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	IPAddress             string
	RefreshTokenExpiredAt int64
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
