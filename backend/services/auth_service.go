package services

import (
	"context"
	"errors"

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
	u, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if !utils.CheckPassword(password, u.PasswordHash) {
		return "", ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(u.ID, u.Username)
	if err != nil {
		return "", err
	}

	return token, nil
}
