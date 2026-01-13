package repositories

import (
	"context"

	"github.com/sisghe/inventory-management-system/backend/db"
	"github.com/sisghe/inventory-management-system/backend/models"
)

type ProductTypeRepository struct{}

func NewProductTypeRepository() *ProductTypeRepository {
	return &ProductTypeRepository{}
}

func (r *ProductTypeRepository) List(ctx context.Context) ([]models.ProductType, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, tipo
		FROM tipo_prodotto
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.ProductType, 0)
	for rows.Next() {
		var t models.ProductType
		if err := rows.Scan(&t.ID, &t.Tipo); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *ProductTypeRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	var ok bool
	err := db.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tipo_prodotto WHERE id=$1)`, id).Scan(&ok)
	return ok, err
}
