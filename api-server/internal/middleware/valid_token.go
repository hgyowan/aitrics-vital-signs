package middleware

import (
	"aitrics-vital-signs/library/envs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ValidTokenMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")

		if token != envs.Token {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token is expired",
			})
			return
		}

		ctx.Next()
	}
}
