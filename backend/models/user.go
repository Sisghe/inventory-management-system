package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Nome         string    `json:"nome"`
	Cognome      string    `json:"cognome"`
	DataNascita  time.Time `json:"data_nascita"`
}
