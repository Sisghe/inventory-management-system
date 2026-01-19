package services

import (
	"context"
	"errors"
	"strings"

	"github.com/sisghe/inventory-management-system/backend/repositories"
	"github.com/sisghe/inventory-management-system/backend/utils"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	users *repositories.UserRepository
}

func NewAuthService(users *repositories.UserRepository) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	// normalizzazione
	username = strings.ToLower(strings.TrimSpace(username))
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	// username deve essere email (ma non vogliamo leakare dettagli in login)
	if err := utils.ValidateEmail(username); err != nil {
		return "", ErrInvalidCredentials
	}

	u, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if !utils.CheckPassword(password, u.PasswordHash) {
		return "", ErrInvalidCredentials
	}

	// blocco login finché non verifica email
	if u.EmailVerifiedAt == nil {
		return "", ErrEmailNotVerified
	}

	token, err := utils.GenerateToken(u.ID, u.Username)
	if err != nil {
		return "", err
	}

	return token, nil
}
