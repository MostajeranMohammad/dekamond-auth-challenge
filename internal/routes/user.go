package routes

import (
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/controllers"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/guards"
	"github.com/gin-gonic/gin"
)

func RegisterUserV1Router(ginEngine *gin.RouterGroup, usersController controllers.UsersController, authGuard guards.AuthGuard) {
	usersGroup := ginEngine.Group("/users")

	usersGroup.GET("/profile", authGuard.JwtGuard, usersController.GetUser)
	usersGroup.GET("/", authGuard.JwtGuard, usersController.GetAllUsers)
}
