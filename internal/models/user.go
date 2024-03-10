package models

import "time"

type User struct {
	ID       int
	Login    string
	Password string
}

type Machine struct {
	IPAddress  string
	MACAddress string
	PublicKey  string
}

type UserMachine struct {
	User    User
	Machine Machine
}

type Session struct {
	ID                    int64
	UserID                int
	AccessToken           string
	RefreshToken          string
	IPAddress             string
	MACAddress            string
	RefreshTokenExpiredAt int64
	PublicKey             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
