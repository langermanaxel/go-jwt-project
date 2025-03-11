package middleware

import (
	"go-jwt-project/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		client_token := ctx.Request.Header.Get("token")
		if client_token == "" {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "No authorization header provided"})
			ctx.Abort()
			return
		}
		claims, err := helpers.ValidateToken(client_token)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			ctx.Abort()
			return
		}
		ctx.Set("email", claims.Email)
		ctx.Set("first_name", claims.First_Name)
		ctx.Set("last_name", claims.Last_Name)
		ctx.Set("uid", claims.Uid)
		ctx.Set("user_type", claims.User_Type)
		ctx.Next()
	}
}
