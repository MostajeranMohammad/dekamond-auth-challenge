package controllers

import (
	"net/http"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/dto"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/usecases"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/pkg/logger"
	"github.com/gin-gonic/gin"
)

type authController struct {
	logger      logger.Logger
	authService usecases.AuthService
}

func NewAuthController(logger logger.Logger, authService usecases.AuthService) AuthController {
	return &authController{
		logger:      logger,
		authService: authService,
	}
}

// @Summary		Request login OTP
// @Description	Generates and sends OTP for the given phone number
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			loginDTO	body	dto.LoginDTO	true	"Login DTO"
// @Success		200
// @Failure		400
// @Failure		401
// @Router			/api/v1/auth/request-otp [post]
func (ac *authController) LoginOtp(c *gin.Context) {
	var body dto.LoginDTO
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ac.authService.LoginRequestOtp(c.Request.Context(), body)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "otp sms sent successfully."})
}

// @Summary		Verify login OTP
// @Description	Verifies OTP, creates user if needed, and returns JWT
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			body	body	dto.VerifyLoginOTP	true	"Verify Login Otp"
// @Success		200
// @Failure		401
// @Router			/api/v1/auth/verify-otp [post]
func (ac *authController) VerifyLoginOTP(c *gin.Context) {
	var body dto.VerifyLoginOTP
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jwt, err := ac.authService.VerifyLoginOTP(c.Request.Context(), body)
	if err != nil {
		ac.logger.Error("Failed to verifying signUp otp", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jwt": jwt,
	})
}
