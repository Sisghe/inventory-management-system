package utils

import "testing"

func TestValidatePasswordAgID_OK(t *testing.T) {
	if err := ValidatePasswordAgID("Abcdefg!"); err != nil {
		t.Fatalf("expected valid password, got error: %v", err)
	}
}

func TestValidatePasswordAgID_TooShort(t *testing.T) {
	if err := ValidatePasswordAgID("Abc!123"); err == nil {
		t.Fatalf("expected error for short password")
	}
}

func TestValidatePasswordAgID_NoUppercase(t *testing.T) {
	if err := ValidatePasswordAgID("abcdefg!"); err == nil {
		t.Fatalf("expected error for missing uppercase letter")
	}
}

func TestValidatePasswordAgID_NoSpecial(t *testing.T) {
	if err := ValidatePasswordAgID("Abcdefgh"); err == nil {
		t.Fatalf("expected error for missing special character")
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	pw := "Abcdefg!1"

	hash, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" || hash == pw {
		t.Fatalf("expected a bcrypt hash, got: %q", hash)
	}

	if !CheckPassword(pw, hash) {
		t.Fatalf("expected CheckPassword to be true for correct password")
	}
	if CheckPassword("WrongPass!1", hash) {
		t.Fatalf("expected CheckPassword to be false for wrong password")
	}
}
