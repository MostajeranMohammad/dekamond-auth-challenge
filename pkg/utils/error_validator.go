package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/pkg/logger"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HttpErrorHandler struct {
	logger     logger.Logger
	controller string
}

func NewHttpErrorHandler(logger logger.Logger, controller string) *HttpErrorHandler {
	return &HttpErrorHandler{
		logger:     logger,
		controller: controller,
	}
}

func (e *HttpErrorHandler) HandleErrorWithHttpStatusCode(err error, c *gin.Context, method string) {
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found"})
				return
			} else if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				c.JSON(http.StatusConflict, gin.H{"error": "This record already exists"})
				return
			} else if strings.Contains(err.Error(), "no rows in result set") {
				c.JSON(http.StatusNotFound, gin.H{"error": "The requested resource was not found"})
				return
			} else if strings.Contains(err.Error(), "violates foreign key constraint") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "This operation cannot be completed due to related data constraints"})
				return
			}

		}

		e.logger.Error(fmt.Sprintf("internal server error in controller %s in method %s with error message %v", e.controller, method, err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
}
