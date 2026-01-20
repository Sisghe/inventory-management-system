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

	ttlHuman := utils.HumanizeDurationMinutes(s.ttl)


	// Email HTML professionale (link “nascosto” dietro bottone)
	body := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; max-width: 560px; margin: 0 auto; line-height: 1.5; color:#111;">
  <h2 style="margin: 0 0 12px;">Verifica la tua email</h2>

  <p style="margin: 0 0 12px;">Ciao,</p>
  <p style="margin: 0 0 18px;">
    per completare la registrazione, conferma il tuo indirizzo email cliccando sul pulsante qui sotto.
  </p>

  <p style="margin: 22px 0;">
    <a href="%s"
       style="background:#0B5ED7;color:#fff;text-decoration:none;padding:12px 18px;border-radius:6px;display:inline-block;">
      Verifica email
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
    Se non hai richiesto questa operazione, puoi ignorare questa email.
  </p>
</div>
`, link, ttlHuman, link, link)

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

// humanizeDurationMinutes converte una durata in una stringa semplice e professionale.
func humanizeDurationMinutes(d time.Duration) string {
	min := int(d.Round(time.Minute) / time.Minute)
	if min <= 1 {
		return "1 minuto"
	}
	return fmt.Sprintf("%d minuti", min)
}
