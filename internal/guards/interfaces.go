package guards

import (
	"github.com/gin-gonic/gin"
)

type (
	AuthGuard interface {
		JwtGuard(c *gin.Context)
	}
)
