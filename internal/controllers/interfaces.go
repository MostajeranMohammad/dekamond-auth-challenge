package controllers

import "github.com/gin-gonic/gin"

type (
	AuthController interface {
		LoginOtp(c *gin.Context)
		VerifyLoginOTP(c *gin.Context)
	}

	UsersController interface {
		GetUser(c *gin.Context)
		GetAllUsers(c *gin.Context)
	}
)
