package repositories

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/sisghe/inventory-management-system/backend/db"
	"github.com/sisghe/inventory-management-system/backend/models"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository { return &UserRepository{} }

// FindByUsername cerca l'utente per username e restituisce anche password_hash.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	u := &models.User{}
	err := db.Pool.QueryRow(ctx, `
		SELECT id, username, password_hash, nome, cognome, data_nascita
		FROM utente
		WHERE username = $1
	`, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Nome, &u.Cognome, &u.DataNascita)

	if err != nil {
		return nil, err
	}
	return u, nil
}

// List restituisce tutti gli utenti (senza password_hash in JSON).
func (r *UserRepository) List(ctx context.Context) ([]models.User, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, username, nome, cognome, data_nascita
		FROM utente
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Nome, &u.Cognome, &u.DataNascita); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

var ErrUsernameTaken = errors.New("username already exists")

func (r *UserRepository) Create(ctx context.Context, u *models.User) (*models.User, error) {
	created := &models.User{}
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO utente (username, password_hash, nome, cognome, data_nascita)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, nome, cognome, data_nascita
	`,
		u.Username, u.PasswordHash, u.Nome, u.Cognome, u.DataNascita,
	).Scan(&created.ID, &created.Username, &created.Nome, &created.Cognome, &created.DataNascita)

	if err != nil {
		// unique violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUsernameTaken
		}
		return nil, err
	}
	return created, nil
}

// Update aggiorna campi selezionati (dinamico). Se PasswordHash è vuoto, non lo aggiorna.
func (r *UserRepository) Update(ctx context.Context, id int, username *string, passwordHash *string, nome *string, cognome *string, dataNascita *time.Time) (*models.User, error) {
	set := make([]string, 0)
	args := make([]any, 0)
	i := 1

	if username != nil {
		set = append(set, "username = $"+itoa(i))
		args = append(args, *username)
		i++
	}
	if passwordHash != nil {
		set = append(set, "password_hash = $"+itoa(i))
		args = append(args, *passwordHash)
		i++
	}
	if nome != nil {
		set = append(set, "nome = $"+itoa(i))
		args = append(args, nome) // pointer ok (NULL se nil, ma qui non nil)
		i++
	}
	if cognome != nil {
		set = append(set, "cognome = $"+itoa(i))
		args = append(args, cognome)
		i++
	}
	if dataNascita != nil {
		set = append(set, "data_nascita = $"+itoa(i))
		args = append(args, dataNascita)
		i++
	}

	if len(set) == 0 {
		return r.GetByID(ctx, id) // niente da aggiornare
	}

	args = append(args, id)
	q := `
		UPDATE utente
		SET ` + strings.Join(set, ", ") + `
		WHERE id = $` + itoa(i) + `
		RETURNING id, username, nome, cognome, data_nascita
	`

	updated := &models.User{}
	err := db.Pool.QueryRow(ctx, q, args...).Scan(&updated.ID, &updated.Username, &updated.Nome, &updated.Cognome, &updated.DataNascita)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUsernameTaken
		}
		return nil, err
	}
	return updated, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	u := &models.User{}
	err := db.Pool.QueryRow(ctx, `
		SELECT id, username, nome, cognome, data_nascita
		FROM utente
		WHERE id = $1
	`, id).Scan(&u.ID, &u.Username, &u.Nome, &u.Cognome, &u.DataNascita)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	ct, err := db.Pool.Exec(ctx, `DELETE FROM utente WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func itoa(n int) string { return strconv.Itoa(n) }
