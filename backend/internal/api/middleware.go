package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google-ai-proxy/internal/auth"
	"google-ai-proxy/internal/db"
)

// UserAuthMiddleware validates the user JWT and loads the user context.
func UserAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "please sign in first"})
			c.Abort()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims, err := auth.ValidateUserToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "login expired, please sign in again"})
			c.Abort()
			return
		}

		var user db.User
		if err := db.DB.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user does not exist"})
			c.Abort()
			return
		}

		if user.Status != "" && user.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{"error": "account has been disabled"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
