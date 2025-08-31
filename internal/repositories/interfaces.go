package repositories

import (
	"context"
	"time"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/entities"
)

type (
	UserRepository interface {
		GetUserByPhone(ctx context.Context, phone string) (entities.User, error)
		GetUserById(ctx context.Context, id uint32) (entities.User, error)
		CreateUser(ctx context.Context, user entities.User) (entities.User, error)
		GetAllUsers(ctx context.Context, skip, limit uint32, phoneSearchTerm *string, creationFrom, creationTo *time.Time) ([]entities.User, error)
	}
)
