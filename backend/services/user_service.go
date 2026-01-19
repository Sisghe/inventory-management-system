package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sisghe/inventory-management-system/backend/models"
	"github.com/sisghe/inventory-management-system/backend/repositories"
	"github.com/sisghe/inventory-management-system/backend/utils"
)

var (
	ErrBadInput         = errors.New("bad input")
	ErrNotFound         = errors.New("not found")
	ErrUserExists       = errors.New("username already exists")
	ErrUsernameTooLong  = errors.New("username max length is 50")
	ErrNomeTooLong      = errors.New("nome max length is 100")
	ErrCognomeTooLong   = errors.New("cognome max length is 100")
	ErrInvalidBirthDate = errors.New("data_nascita is invalid")
	ErrPasswordTooLong  = errors.New("password is too long (max 72 characters)")
)

const (
	maxUsernameLen = 50  // vincolo DB: varchar(50)
	maxNameLen     = 100 // vincolo DB: varchar(100)
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

func normalizeEmailUsername(username string) (string, error) {
	u := strings.TrimSpace(strings.ToLower(username))
	if u == "" {
		return "", ErrBadInput
	}
	if len(u) > maxUsernameLen {
		return "", ErrUsernameTooLong
	}
	if err := utils.ValidateEmail(u); err != nil {
		return "", err
	}
	return u, nil
}

func normalizeOptionalName(p *string, field string) (*string, error) {
	if p == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*p)
	if v == "" {
		// su CREATE, consideriamo vuoto come "non presente"
		return nil, nil
	}
	if len(v) > maxNameLen {
		if field == "nome" {
			return nil, ErrNomeTooLong
		}
		return nil, ErrCognomeTooLong
	}
	return &v, nil
}

func validateBirthDate(d *time.Time) error {
	if d == nil {
		return nil
	}

	now := time.Now()
	// non accettiamo date future
	if d.After(now) {
		return ErrInvalidBirthDate
	}

	// non accettiamo date troppo vecchie (es. > 120 anni)
	min := now.AddDate(-120, 0, 0)
	if d.Before(min) {
		return ErrInvalidBirthDate
	}

	return nil
}

func (s *UserService) Create(
	ctx context.Context,
	username, password string,
	nome, cognome *string,
	dataNascita *time.Time,
) (*models.User, error) {

	password = strings.TrimSpace(password)
	if password == "" {
		return nil, ErrBadInput
	}
	// bcrypt tronca oltre 72 byte/caratteri: meglio bloccare esplicitamente
	if len(password) > 72 {
		return nil, ErrPasswordTooLong
	}

	uNorm, err := normalizeEmailUsername(username)
	if err != nil {
		return nil, err
	}

	if err := utils.ValidatePasswordAgID(password); err != nil {
		return nil, err
	}

	nomeNorm, err := normalizeOptionalName(nome, "nome")
	if err != nil {
		return nil, err
	}
	cognomeNorm, err := normalizeOptionalName(cognome, "cognome")
	if err != nil {
		return nil, err
	}

	if err := validateBirthDate(dataNascita); err != nil {
		return nil, err
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		Username:     uNorm,
		PasswordHash: hash,
		Nome:         nomeNorm,
		Cognome:      cognomeNorm,
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

func (s *UserService) Update(
	ctx context.Context,
	id int,
	username *string,
	password *string,
	nome, cognome *string,
	dataNascita *time.Time,
) (*models.User, error) {

	if id <= 0 {
		return nil, ErrBadInput
	}

	// username: se presente, deve essere email valida + normalizzata
	var usernameNorm *string
	if username != nil {
		u := strings.TrimSpace(*username)
		if u == "" {
			return nil, ErrBadInput
		}
		n, err := normalizeEmailUsername(u)
		if err != nil {
			return nil, err
		}
		usernameNorm = &n
	}

	// password: se presente, valida AgID + max 72
	var passwordHash *string
	if password != nil {
		p := strings.TrimSpace(*password)
		if p == "" {
			return nil, ErrBadInput
		}
		if len(p) > 72 {
			return nil, ErrPasswordTooLong
		}
		if err := utils.ValidatePasswordAgID(p); err != nil {
			return nil, err
		}
		h, err := utils.HashPassword(p)
		if err != nil {
			return nil, err
		}
		passwordHash = &h
	}

	// nome/cognome: in UPDATE, se mandati come "", li trattiamo come errore
	// (omissione = non modificare)
	var nomeNorm *string
	if nome != nil {
		v := strings.TrimSpace(*nome)
		if v == "" {
			return nil, errors.New("nome cannot be empty; omit field to keep unchanged")
		}
		if len(v) > maxNameLen {
			return nil, ErrNomeTooLong
		}
		nomeNorm = &v
	}

	var cognomeNorm *string
	if cognome != nil {
		v := strings.TrimSpace(*cognome)
		if v == "" {
			return nil, errors.New("cognome cannot be empty; omit field to keep unchanged")
		}
		if len(v) > maxNameLen {
			return nil, ErrCognomeTooLong
		}
		cognomeNorm = &v
	}

	if err := validateBirthDate(dataNascita); err != nil {
		return nil, err
	}

	updated, err := s.repo.Update(ctx, id, usernameNorm, passwordHash, nomeNorm, cognomeNorm, dataNascita)
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
	if id <= 0 {
		return ErrBadInput
	}
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
