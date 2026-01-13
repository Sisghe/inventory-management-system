package models

import "time"

type User struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Nome         *string    `json:"nome,omitempty"`
	Cognome      *string    `json:"cognome,omitempty"`
	DataNascita  *time.Time `json:"data_nascita,omitempty"`
}
