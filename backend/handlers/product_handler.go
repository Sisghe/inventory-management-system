package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/sisghe/inventory-management-system/backend/services"
)

type ProductHandler struct {
	products *services.ProductService
}

func NewProductHandler(products *services.ProductService) *ProductHandler {
	return &ProductHandler{products: products}
}

func (h *ProductHandler) List(c *gin.Context) {
	out, err := h.products.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, out)
}

type createProductRequest struct {
	NomeOggetto    string  `json:"nome_oggetto"`
	Descrizione    *string `json:"descrizione"`
	TipoProdottoID *int    `json:"tipo_prodotto_id"`
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	req.NomeOggetto = strings.TrimSpace(req.NomeOggetto)

	// validazioni base (handler)
	if req.NomeOggetto == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nome_oggetto is required"})
		return
	}
	if req.TipoProdottoID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tipo_prodotto_id is required"})
		return
	}
	if *req.TipoProdottoID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tipo_prodotto_id must be > 0"})
		return
	}

	created, err := h.products.Create(c.Request.Context(), req.NomeOggetto, req.Descrizione, req.TipoProdottoID)
	if err != nil {
		switch err {
		case services.ErrProductBadInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad input"})
		case services.ErrInvalidProductType:
			c.JSON(http.StatusBadRequest, gin.H{"error": "tipo_prodotto_id is invalid"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	req.NomeOggetto = strings.TrimSpace(req.NomeOggetto)

	if req.NomeOggetto == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nome_oggetto is required"})
		return
	}
	if req.TipoProdottoID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tipo_prodotto_id is required"})
		return
	}
	if *req.TipoProdottoID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tipo_prodotto_id must be > 0"})
		return
	}

	updated, err := h.products.Update(c.Request.Context(), id, req.NomeOggetto, req.Descrizione, req.TipoProdottoID)
	if err != nil {
		switch err {
		case services.ErrProductBadInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad input"})
		case services.ErrInvalidProductType:
			c.JSON(http.StatusBadRequest, gin.H{"error": "tipo_prodotto_id is invalid"})
		case services.ErrProductNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.products.Delete(c.Request.Context(), id); err != nil {
		switch err {
		case services.ErrProductNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		case services.ErrProductBadInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
