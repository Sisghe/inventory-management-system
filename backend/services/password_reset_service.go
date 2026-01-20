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

	// anti-enumeration: se email non valida o utente non trovato, non riveliamo nulla
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

	ttlHuman := utils.HumanizeDurationMinutes(s.ttl)


	// Email HTML professionale (bottone + fallback link)
	body := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; max-width: 560px; margin: 0 auto; line-height: 1.5; color:#111;">
  <h2 style="margin: 0 0 12px;">Reimposta la tua password</h2>

  <p style="margin: 0 0 12px;">Ciao,</p>
  <p style="margin: 0 0 18px;">
    abbiamo ricevuto una richiesta di reimpostazione della password. Puoi procedere cliccando sul pulsante qui sotto.
  </p>

  <p style="margin: 22px 0;">
    <a href="%s"
       style="background:#0B5ED7;color:#fff;text-decoration:none;padding:12px 18px;border-radius:6px;display:inline-block;">
      Reimposta password
    </a>
  </p>

  <p style="margin: 0 0 10px; color:#555; font-size: 13px;">
    Il link scade tra %s.
  </p>

  <p style="margin: 18px 0 0; color:#555; font-size: 13px;">
    Se il pulsante non funziona, copia e incolla questo link nel browser:<br/>
    <a href="%s" style="color:#0B5ED7;">%s</a>
  </p>

  <hr style="border:none;border-top:1px solid #eee;margin:24px 0;" />

  <p style="margin:0; color:#777; font-size: 12px;">
    Se non hai richiesto tu il reset della password, puoi ignorare questa email.
  </p>
</div>
`, link, ttlHuman, link, link)

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
