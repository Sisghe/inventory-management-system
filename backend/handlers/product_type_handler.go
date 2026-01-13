package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sisghe/inventory-management-system/backend/repositories"
)

type ProductTypeHandler struct {
	types *repositories.ProductTypeRepository
}

func NewProductTypeHandler(types *repositories.ProductTypeRepository) *ProductTypeHandler {
	return &ProductTypeHandler{types: types}
}

func (h *ProductTypeHandler) List(c *gin.Context) {
	out, err := h.types.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, out)
}
