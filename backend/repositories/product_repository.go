package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/sisghe/inventory-management-system/backend/db"
	"github.com/sisghe/inventory-management-system/backend/models"
)

type ProductRepository struct{}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{}
}

var ErrForeignKey = errors.New("invalid foreign key")

func (r *ProductRepository) List(ctx context.Context) ([]models.Product, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, nome_oggetto, descrizione, data_inserimento, tipo_prodotto_id
		FROM prodotto
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Product, 0)
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.NomeOggetto, &p.Descrizione, &p.DataInserimento, &p.TipoProdottoID); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProductRepository) Create(ctx context.Context, p *models.Product) (*models.Product, error) {
	created := &models.Product{}
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO prodotto (nome_oggetto, descrizione, tipo_prodotto_id)
		VALUES ($1, $2, $3)
		RETURNING id, nome_oggetto, descrizione, data_inserimento, tipo_prodotto_id
	`, p.NomeOggetto, p.Descrizione, p.TipoProdottoID).
		Scan(&created.ID, &created.NomeOggetto, &created.Descrizione, &created.DataInserimento, &created.TipoProdottoID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, ErrForeignKey
		}
		return nil, err
	}
	return created, nil
}

func (r *ProductRepository) Update(ctx context.Context, id int, nome string, descrizione *string, tipoID *int) (*models.Product, error) {
	updated := &models.Product{}
	err := db.Pool.QueryRow(ctx, `
		UPDATE prodotto
		SET nome_oggetto = $1,
		    descrizione = $2,
		    tipo_prodotto_id = $3
		WHERE id = $4
		RETURNING id, nome_oggetto, descrizione, data_inserimento, tipo_prodotto_id
	`, nome, descrizione, tipoID, id).
		Scan(&updated.ID, &updated.NomeOggetto, &updated.Descrizione, &updated.DataInserimento, &updated.TipoProdottoID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, ErrForeignKey
		}
		return nil, err
	}
	return updated, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int) error {
	ct, err := db.Pool.Exec(ctx, `DELETE FROM prodotto WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
