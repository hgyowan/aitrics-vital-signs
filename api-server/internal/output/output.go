package output

import (
	pkgError "aitrics-vital-signs/library/error"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Output struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func Send(ctx *gin.Context, respData interface{}) {

	ctx.JSON(http.StatusOK, Output{
		Code: 1,
		Data: respData,
	})

	ctx.Abort()
}

func AppendErrorContext(ctx *gin.Context, err error, respData interface{}) {
	if err != nil {
		_ = ctx.Error(err)
		castedErr, ok := pkgError.CastBusinessError(err)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, Output{
				Code: int(pkgError.None),
				Data: nil,
			})
			return
		}

		if castedErr.Status.Data != nil {
			respData = castedErr.Status.Data
		}

		ctx.JSON(castedErr.Status.HttpStatusCode, Output{
			Code: castedErr.Status.Code,
			Data: respData,
		})
	}

	ctx.Abort()
}
