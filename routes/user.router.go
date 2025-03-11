package routes

import (
	"go-jwt-project/controllers"
	"go-jwt-project/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incoming_routes *gin.Engine) {
	incoming_routes.Use(middleware.Authenticate())
	incoming_routes.GET("/users", controllers.GetUsers())
	incoming_routes.GET("/users/:id", controllers.GetUser())
}
