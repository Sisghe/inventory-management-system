package services

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sisghe/inventory-management-system/backend/models"
	"github.com/sisghe/inventory-management-system/backend/repositories"
	"github.com/sisghe/inventory-management-system/backend/utils"
)

var (
	ErrBadInput   = errors.New("bad input")
	ErrNotFound   = errors.New("not found")
	ErrUserExists = errors.New("username already exists")
)

type UserService struct {
	repo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) List(ctx context.Context) ([]models.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) Create(ctx context.Context, username, password string, nome, cognome *string, dataNascita *time.Time) (*models.User, error) {
	if username == "" || password == "" {
		return nil, ErrBadInput
	}
	if err := utils.ValidatePasswordAgID(password); err != nil {
		return nil, err
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		Username:     username,
		PasswordHash: hash,
		Nome:         nome,
		Cognome:      cognome,
		DataNascita:  dataNascita,
	}

	created, err := s.repo.Create(ctx, u)
	if err != nil {
		if err == repositories.ErrUsernameTaken {
			return nil, ErrUserExists
		}
		return nil, err
	}
	return created, nil
}

func (s *UserService) Update(ctx context.Context, id int, username *string, password *string, nome, cognome *string, dataNascita *time.Time) (*models.User, error) {
	var passwordHash *string
	if password != nil {
		if *password == "" {
			return nil, ErrBadInput
		}
		if err := utils.ValidatePasswordAgID(*password); err != nil {
			return nil, err
		}
		h, err := utils.HashPassword(*password)
		if err != nil {
			return nil, err
		}
		passwordHash = &h
	}

	updated, err := s.repo.Update(ctx, id, username, passwordHash, nome, cognome, dataNascita)
	if err != nil {
		if err == repositories.ErrUsernameTaken {
			return nil, ErrUserExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return updated, nil
}

func (s *UserService) Delete(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
