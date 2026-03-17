package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		token := ctx.GetHeader("authorization")
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "token isn't found",
			})
			return
		}
		ctx.Set("token", token)
		ctx.Next()
	}
}
