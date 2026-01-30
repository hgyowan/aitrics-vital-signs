package middleware

import (
	pkgLogger "aitrics-vital-signs/library/logger"

	"github.com/gin-gonic/gin"
)

func GinBusinessErrLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx != nil {
			ctx.Next()
			err := ctx.Errors.Last()
			if err != nil {
				pkgLogger.ZapLogger.Logger.Error(err.Error())
			}
		}
	}
}
