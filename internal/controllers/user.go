package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/entities"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/usecases"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/pkg/logger"
	"github.com/gin-gonic/gin"
)

type usersController struct {
	usersService usecases.UsersService
}

func NewUsersController(l logger.Logger, usersService usecases.UsersService) UsersController {
	return &usersController{
		usersService: usersService,
	}
}

// @Summary Get user by id
// @Tags Users
// @Produce json
// @Success 200
// @Failure 400
// @Failure 404
// @Router /api/v1/users/profile [get]
// @Security		BearerAuth
func (uc *usersController) GetUser(c *gin.Context) {
	user, ok := c.MustGet("user").(entities.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user from context"})
		return
	}

	user, err := uc.usersService.GetUser(c.Request.Context(), user.Id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// @Summary List users
// @Tags Users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size" default(10)
// @Param search query string false "Search by phone"
// @Param created_from query string false "Created from (RFC3339)"
// @Param created_to query string false "Created to (RFC3339)"
// @Success 200
// @Router /api/v1/users [get]
// @Security		BearerAuth
func (uc *usersController) GetAllUsers(c *gin.Context) {
	page64, _ := strconv.ParseUint(c.DefaultQuery("page", "1"), 10, 32)
	limit64, _ := strconv.ParseUint(c.DefaultQuery("limit", "10"), 10, 32)
	page := uint32(page64)
	limit := uint32(limit64)

	search := c.Query("search")
	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	var fromPtr, toPtr *time.Time
	if v := c.Query("created_from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			fromPtr = &t
		}
	}
	if v := c.Query("created_to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			toPtr = &t
		}
	}

	users, err := uc.usersService.GetAllUsers(c.Request.Context(), page, limit, searchPtr, fromPtr, toPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}
