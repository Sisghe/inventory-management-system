package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sisghe/inventory-management-system/backend/utils"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 1) Preferisci Authorization: Bearer <token> (compat con Thunder Client)
		auth := c.GetHeader("Authorization")
		if auth != "" {
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
				return
			}
		}

		// 2) Fallback: cookie HttpOnly (per Next middleware e fetch con credentials: include)
		if token == "" {
			if cookieToken, err := c.Cookie("access_token"); err == nil {
				token = cookieToken
			}
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
