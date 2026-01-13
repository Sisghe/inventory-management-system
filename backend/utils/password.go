package utils

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// ValidatePasswordAgID controlla le regole richieste:
// - almeno 8 caratteri
// - almeno 1 lettera maiuscola
// - almeno 1 carattere speciale
func ValidatePasswordAgID(pw string) error {
	if len(pw) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	hasUpper := false
	hasSpecial := false

	for _, r := range pw {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		// "speciale" = non lettera e non numero
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// HashPassword genera un hash bcrypt da salvare nel DB.
func HashPassword(pw string) (string, error) {
	// costo standard: bcrypt.DefaultCost
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword confronta password in chiaro con hash bcrypt salvato nel DB.
func CheckPassword(pw string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}
