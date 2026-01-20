package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken_OK(t *testing.T) {
	t.Setenv("JWT_SECRET", "testsecret")

	token, err := GenerateToken(21, "test@a.com")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if claims.UserID != 21 || claims.Username != "test@a.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		t.Fatalf("expected token not expired")
	}
}

func TestGenerateToken_EmptySecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	_, err := GenerateToken(1, "x")
	if err == nil {
		t.Fatalf("expected error when JWT_SECRET is empty")
	}
}

func TestValidateToken_EmptySecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	_, err := ValidateToken("whatever")
	if err == nil {
		t.Fatalf("expected error when JWT_SECRET is empty")
	}
}

func TestValidateToken_TamperedToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "testsecret")

	token, err := GenerateToken(1, "x")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	// manomette il token (cambia l'ultimo carattere)
	tampered := token[:len(token)-1]
	if token[len(token)-1] == 'a' {
		tampered += "b"
	} else {
		tampered += "a"
	}

	_, err = ValidateToken(tampered)
	if err == nil {
		t.Fatalf("expected error for tampered token")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "testsecret")
	secret := []byte("testsecret")

	claims := Claims{
		UserID:   1,
		Username: "x",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // scaduto
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := tok.SignedString(secret)
	if err != nil {
		t.Fatalf("failed signing token: %v", err)
	}

	_, err = ValidateToken(tokenStr)
	if err == nil {
		t.Fatalf("expected error for expired token")
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	t.Setenv("JWT_SECRET", "testsecret")
	secret := []byte("testsecret")

	claims := Claims{
		UserID:   1,
		Username: "x",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	}

	// firma con metodo diverso da HS256 (ValidateToken deve rifiutare)
	tok := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	tokenStr, err := tok.SignedString(secret)
	if err != nil {
		t.Fatalf("failed signing token: %v", err)
	}

	_, err = ValidateToken(tokenStr)
	if err == nil {
		t.Fatalf("expected error for unexpected signing method")
	}
}
