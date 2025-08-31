package routes

import (
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/controllers"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/guards"
	"github.com/gin-gonic/gin"
)

func RegisterAuthV1Router(ginEngine *gin.RouterGroup, authController controllers.AuthController, authGuard guards.AuthGuard) {
	authGroup := ginEngine.Group("/auth")

	authGroup.POST("/request-otp", authController.LoginOtp)
	authGroup.POST("/verify-otp", authController.VerifyLoginOTP)
}
