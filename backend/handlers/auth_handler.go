package handlers

import (
	"encoding/hex"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sisghe/inventory-management-system/backend/repositories"
	"github.com/sisghe/inventory-management-system/backend/services"
	"github.com/sisghe/inventory-management-system/backend/utils"
)

type AuthHandler struct {
	auth          *services.AuthService
	emailVerify   *services.EmailVerificationService
	passwordReset *services.PasswordResetService
}

func NewAuthHandler(
	auth *services.AuthService,
	emailVerify *services.EmailVerificationService,
	passwordReset *services.PasswordResetService,
) *AuthHandler {
	return &AuthHandler{auth: auth, emailVerify: emailVerify, passwordReset: passwordReset}
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

	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	token, err := h.auth.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		case services.ErrEmailNotVerified:
			c.JSON(http.StatusForbidden, gin.H{"error": "email not verified. check your inbox to verify your account"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	}

	secure := os.Getenv("COOKIE_SECURE") == "true"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", token, 60*60, "/", "", secure, true)

	c.JSON(http.StatusOK, gin.H{"access_token": token})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", "", -1, "/", "", secure, true)
	c.Status(http.StatusNoContent)
}

func isValidHexToken(token string, expectedLen int) bool {
	if len(token) != expectedLen {
		return false
	}
	_, err := hex.DecodeString(token)
	return err == nil
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	if h.emailVerify == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "email verification not configured"})
		return
	}

	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}
	if !isValidHexToken(req.Token, 64) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	if err := h.emailVerify.Verify(c.Request.Context(), req.Token); err != nil {
		if err == repositories.ErrInvalidVerificationToken {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Status(http.StatusNoContent)
}

type forgotPasswordRequest struct {
	Username string `json:"username"` // email
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	if h.passwordReset == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password reset not configured"})
		return
	}

	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	// qui ha senso essere espliciti: username deve essere email valida
	if err := utils.ValidateEmail(req.Username); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// anti-enumeration: rispondiamo 204 comunque
	_ = h.passwordReset.Request(c.Request.Context(), req.Username)
	c.Status(http.StatusNoContent)
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	if h.passwordReset == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password reset not configured"})
		return
	}

	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	req.Password = strings.TrimSpace(req.Password)

	if req.Token == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token and password are required"})
		return
	}
	if !isValidHexToken(req.Token, 64) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}
	// bcrypt considera solo i primi 72 caratteri: meglio bloccare oltre
	if len(req.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password too long (max 72 characters)"})
		return
	}

	if err := h.passwordReset.Reset(c.Request.Context(), req.Token, req.Password); err != nil {
		switch err {
		case repositories.ErrInvalidResetToken:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
			return
		case services.ErrPasswordSameAsOld:
			c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be different from previous"})
			return
		default:
			// include errori AgID
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.Status(http.StatusNoContent)
}
