package services

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/sisghe/inventory-management-system/backend/models"
	"github.com/sisghe/inventory-management-system/backend/repositories"
)

var (
	ErrProductBadInput    = errors.New("bad input")
	ErrProductNotFound    = errors.New("not found")
	ErrInvalidProductType = errors.New("invalid product type")
)

type ProductService struct {
	products *repositories.ProductRepository
	types    *repositories.ProductTypeRepository
}

func NewProductService(products *repositories.ProductRepository, types *repositories.ProductTypeRepository) *ProductService {
	return &ProductService{products: products, types: types}
}

func (s *ProductService) List(ctx context.Context) ([]models.Product, error) {
	return s.products.List(ctx)
}

func (s *ProductService) Create(ctx context.Context, nome string, descrizione *string, tipoID *int) (*models.Product, error) {
	if nome == "" {
		return nil, ErrProductBadInput
	}
	// tipoID può essere nil (DB lo permette), ma nel progetto lo vogliamo obbligatorio:
	if tipoID == nil {
		return nil, ErrInvalidProductType
	}
	ok, err := s.types.ExistsByID(ctx, *tipoID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidProductType
	}

	p := &models.Product{
		NomeOggetto:    nome,
		Descrizione:    descrizione,
		TipoProdottoID: tipoID,
	}

	created, err := s.products.Create(ctx, p)
	if err != nil {
		if err == repositories.ErrForeignKey {
			return nil, ErrInvalidProductType
		}
		return nil, err
	}
	return created, nil
}

func (s *ProductService) Update(ctx context.Context, id int, nome string, descrizione *string, tipoID *int) (*models.Product, error) {
	if id <= 0 || nome == "" {
		return nil, ErrProductBadInput
	}
	if tipoID == nil {
		return nil, ErrInvalidProductType
	}
	ok, err := s.types.ExistsByID(ctx, *tipoID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidProductType
	}

	updated, err := s.products.Update(ctx, id, nome, descrizione, tipoID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		if err == repositories.ErrForeignKey {
			return nil, ErrInvalidProductType
		}
		return nil, err
	}
	return updated, nil
}

func (s *ProductService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrProductBadInput
	}
	err := s.products.Delete(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrProductNotFound
	}
	return err
}
