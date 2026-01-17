package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sisghe/inventory-management-system/backend/db"
)

var ErrInvalidResetToken = errors.New("invalid or expired reset token")

type PasswordResetRepository struct{}

func NewPasswordResetRepository() *PasswordResetRepository {
	return &PasswordResetRepository{}
}

func (r *PasswordResetRepository) Create(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO public.password_reset_token (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *PasswordResetRepository) Consume(ctx context.Context, tokenHash string) (int, error) {
	var userID int
	err := db.Pool.QueryRow(ctx, `
		UPDATE public.password_reset_token
		SET used_at = NOW()
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND expires_at > NOW()
		RETURNING user_id
	`, tokenHash).Scan(&userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidResetToken
		}
		return 0, err
	}
	return userID, nil
}
