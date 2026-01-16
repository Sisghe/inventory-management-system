package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sisghe/inventory-management-system/backend/services"
)

type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	// trim per evitare username/password con soli spazi
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	token, err := h.auth.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	//  Set cookie HttpOnly (utile per middleware.ts di Next, che non legge localStorage)
	// In dev: COOKIE_SECURE=false (http). In prod: true (https).
	secure := os.Getenv("COOKIE_SECURE") == "true"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"access_token",
		token,
		60*60, // 1 ora (allinea al TTL del JWT se è 1h)
		"/",
		"",
		secure,
		true, // HttpOnly
	)

	// Manteniamo anche la risposta JSON per compatibilità
	c.JSON(http.StatusOK, gin.H{"access_token": token})
}

//  Logout: invalida il cookie (necessario se il token è HttpOnly)
func (h *AuthHandler) Logout(c *gin.Context) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", "", -1, "/", "", secure, true)
	c.Status(http.StatusNoContent)
}
