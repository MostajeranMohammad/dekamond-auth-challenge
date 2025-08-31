package usecases

import (
	"context"
	"time"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/entities"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/repositories"
)

type usersUsecase struct {
	usersRepo repositories.UserRepository
}

func NewUsersService(usersRepo repositories.UserRepository) UsersService {
	return &usersUsecase{
		usersRepo: usersRepo,
	}
}

func (u *usersUsecase) GetUser(ctx context.Context, id uint32) (entities.User, error) {
	return u.usersRepo.GetUserById(ctx, id)
}

func (u *usersUsecase) GetAllUsers(ctx context.Context, page, limit uint32, phoneSearchTerm *string, creationFrom, creationTo *time.Time) ([]entities.User, error) {
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}
	skip := (page - 1) * limit
	return u.usersRepo.GetAllUsers(ctx, skip, limit, phoneSearchTerm, creationFrom, creationTo)
}
