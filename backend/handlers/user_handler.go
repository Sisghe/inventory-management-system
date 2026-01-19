package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sisghe/inventory-management-system/backend/services"
)

type UserHandler struct {
	users       *services.UserService
	emailVerify *services.EmailVerificationService
}

func NewUserHandler(users *services.UserService, emailVerify *services.EmailVerificationService) *UserHandler {
	return &UserHandler{users: users, emailVerify: emailVerify}
}

func (h *UserHandler) List(c *gin.Context) {
	out, err := h.users.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, out)
}

type createUserRequest struct {
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	Nome        *string `json:"nome"`
	Cognome     *string `json:"cognome"`
	DataNascita *string `json:"data_nascita"` // "YYYY-MM-DD"
}

func (h *UserHandler) Create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	var dn *time.Time
	if req.DataNascita != nil && strings.TrimSpace(*req.DataNascita) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.DataNascita))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "data_nascita must be YYYY-MM-DD"})
			return
		}
		dn = &t
	}

	created, err := h.users.Create(c.Request.Context(), req.Username, req.Password, req.Nome, req.Cognome, dn)
	if err != nil {
		switch err {
		case services.ErrBadInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		case services.ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		default:
			// include: email non valida, password AgID, lunghezze, data_nascita invalida...
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	if h.emailVerify == nil {
		_ = h.users.Delete(c.Request.Context(), created.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "email verification not configured"})
		return
	}

	if err := h.emailVerify.Send(c.Request.Context(), created.ID, created.Username); err != nil {
		log.Println("failed to send verification email:", err)
		_ = h.users.Delete(c.Request.Context(), created.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send verification email"})
		return
	}

	c.JSON(http.StatusCreated, created)
}

type updateUserRequest struct {
	Username    *string `json:"username"`
	Password    *string `json:"password"`
	Nome        *string `json:"nome"`
	Cognome     *string `json:"cognome"`
	DataNascita *string `json:"data_nascita"` // "YYYY-MM-DD"
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	var dn *time.Time
	if req.DataNascita != nil {
		v := strings.TrimSpace(*req.DataNascita)
		if v == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "data_nascita cannot be empty; omit field to keep unchanged"})
			return
		}
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "data_nascita must be YYYY-MM-DD"})
			return
		}
		dn = &t
	}

	updated, err := h.users.Update(c.Request.Context(), id, req.Username, req.Password, req.Nome, req.Cognome, dn)
	if err != nil {
		switch err {
		case services.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case services.ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.users.Delete(c.Request.Context(), id); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Status(http.StatusNoContent)
}
