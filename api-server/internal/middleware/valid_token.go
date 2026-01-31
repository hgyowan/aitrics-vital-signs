package middleware

import (
	"aitrics-vital-signs/library/envs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ValidTokenMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header missing",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		token := parts[1]

		if token != envs.Token {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token is expired",
			})
			return
		}

		ctx.Next()
	}
}
