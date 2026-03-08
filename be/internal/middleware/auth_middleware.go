package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/reshap/trading-bot/internal/service"
)

// AuthMiddleware creates a middleware that validates JWT tokens
func AuthMiddleware(srvc *service.Services) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get token from header
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Missing authorization token",
			})
			ctx.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(tokenString) > 7 && strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = tokenString[7:]
		}

		claims, err := srvc.ValidateToken(tokenString)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
			})
			ctx.Abort()
			return
		}

		// Store claims in context for use in handlers
		ctx.Set("username", claims.Username)
		ctx.Set("name", claims.Name)
		ctx.Set("claims", claims)

		ctx.Next()
	}
}
