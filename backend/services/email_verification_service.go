package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/sisghe/inventory-management-system/backend/repositories"
	"github.com/sisghe/inventory-management-system/backend/utils"
)

var ErrEmailNotVerified = errors.New("email not verified")

type EmailVerificationService struct {
	repo      *repositories.EmailVerificationRepository
	usersRepo *repositories.UserRepository
	mailer    utils.Mailer
	frontend  string
	ttl       time.Duration
}

func NewEmailVerificationService(repo *repositories.EmailVerificationRepository, usersRepo *repositories.UserRepository, mailer utils.Mailer) *EmailVerificationService {
	frontend := os.Getenv("FRONTEND_URL")
	if frontend == "" {
		frontend = "http://localhost:3000"
	}

	ttlMin := 30
	if v := os.Getenv("EMAIL_VERIFY_TTL_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttlMin = n
		}
	}

	return &EmailVerificationService{
		repo:      repo,
		usersRepo: usersRepo,
		mailer:    mailer,
		frontend:  frontend,
		ttl:       time.Duration(ttlMin) * time.Minute,
	}
}

func (s *EmailVerificationService) Send(ctx context.Context, userID int, email string) error {
	token, err := utils.GenerateRandomToken(32) // 64 hex chars
	if err != nil {
		return err
	}
	tokenHash := utils.HashToken(token)

	expiresAt := time.Now().Add(s.ttl)
	if err := s.repo.Create(ctx, userID, tokenHash, expiresAt); err != nil {
		return err
	}

	link := fmt.Sprintf("%s/verify-email?token=%s", s.frontend, url.QueryEscape(token))
	subject := "Verifica la tua email"
	body := "Ciao!\n\nPer completare la registrazione, verifica la tua email cliccando qui:\n\n" +
		link + "\n\n" +
		"Il link scade tra " + s.ttl.String() + ".\n"

	return s.mailer.Send(email, subject, body)
}

func (s *EmailVerificationService) Verify(ctx context.Context, token string) error {
	tokenHash := utils.HashToken(token)
	userID, err := s.repo.Consume(ctx, tokenHash)
	if err != nil {
		return err
	}
	_, err = s.usersRepo.MarkEmailVerified(ctx, userID)
	return err
}
