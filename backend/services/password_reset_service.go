package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sisghe/inventory-management-system/backend/repositories"
	"github.com/sisghe/inventory-management-system/backend/utils"
)

var ErrPasswordSameAsOld = errors.New("new password must be different from previous")

type PasswordResetService struct {
	repo     *repositories.PasswordResetRepository
	users    *repositories.UserRepository
	mailer   utils.Mailer
	frontend string
	ttl      time.Duration
}

func NewPasswordResetService(repo *repositories.PasswordResetRepository, users *repositories.UserRepository, mailer utils.Mailer) *PasswordResetService {
	frontend := os.Getenv("FRONTEND_URL")
	if frontend == "" {
		frontend = "http://localhost:3000"
	}

	ttlMin := 30
	if v := os.Getenv("PASSWORD_RESET_TTL_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttlMin = n
		}
	}

	return &PasswordResetService{
		repo:     repo,
		users:    users,
		mailer:   mailer,
		frontend: frontend,
		ttl:      time.Duration(ttlMin) * time.Minute,
	}
}

func (s *PasswordResetService) Request(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	// se per qualche motivo arrivasse qui email non valida, non facciamo nulla (anti-enumeration)
	if err := utils.ValidateEmail(email); err != nil {
		return nil
	}

	u, err := s.users.FindByUsername(ctx, email)
	if err != nil {
		return nil
	}

	token, err := utils.GenerateRandomToken(32)
	if err != nil {
		return err
	}
	tokenHash := utils.HashToken(token)

	expiresAt := time.Now().Add(s.ttl)
	if err := s.repo.Create(ctx, u.ID, tokenHash, expiresAt); err != nil {
		return err
	}

	link := fmt.Sprintf("%s/reset-password?token=%s", s.frontend, url.QueryEscape(token))
	subject := "Recupero password"
	body := "Ciao!\n\nPer reimpostare la password clicca qui:\n\n" +
		link + "\n\n" +
		"Il link scade tra " + s.ttl.String() + ".\n" +
		"Se non hai richiesto tu il reset, ignora questa email.\n"

	return s.mailer.Send(email, subject, body)
}

func (s *PasswordResetService) Reset(ctx context.Context, token string, newPassword string) error {
	token = strings.TrimSpace(token)
	newPassword = strings.TrimSpace(newPassword)

	if token == "" || newPassword == "" {
		return errors.New("token and password are required")
	}
	if len(newPassword) > 72 {
		return errors.New("password too long (max 72 characters)")
	}

	tokenHash := utils.HashToken(token)

	userID, err := s.repo.Consume(ctx, tokenHash)
	if err != nil {
		return err
	}

	oldHash, err := s.users.GetPasswordHashByID(ctx, userID)
	if err != nil {
		return err
	}

	if utils.CheckPassword(newPassword, oldHash) {
		return ErrPasswordSameAsOld
	}

	// regole AgID (specifica)
	if err := utils.ValidatePasswordAgID(newPassword); err != nil {
		return err
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.users.UpdatePasswordHash(ctx, userID, newHash)
}
