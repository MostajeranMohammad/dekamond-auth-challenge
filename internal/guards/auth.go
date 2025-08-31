package guards

import (
	"net/http"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/usecases"
	"github.com/gin-gonic/gin"
)

type authGuard struct {
	authService usecases.AuthService
}

func NewAuthGuard(authService usecases.AuthService) AuthGuard {
	return &authGuard{authService: authService}
}
func (ag *authGuard) JwtGuard(c *gin.Context) {
	user, err := ag.authService.ValidateToken(c.Request.Context(), c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{})
		c.Abort()
		return
	}

	c.Set("user", user)
	c.Next()
}
