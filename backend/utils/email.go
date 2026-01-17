package utils

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidEmail = errors.New("username must be a valid email")

// Regex "ragionevole" per progetti reali (non perfetta RFC, ma evita casi palesemente errati)
// - deve esserci un solo @
// - dominio con almeno un punto
// - TLD solo lettere, min 2 caratteri
var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[A-Za-z]{2,24}$`)

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrInvalidEmail
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}
