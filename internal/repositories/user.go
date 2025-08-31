package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUserByPhone(ctx context.Context, phone string) (entities.User, error) {
	var u entities.User
	var id32 int32

	err := r.db.QueryRow(ctx,
		`SELECT id, phone, created_at FROM users WHERE phone = $1`,
		phone,
	).Scan(&id32, &u.Phone, &u.CreatedAt)
	if err != nil {
		return entities.User{}, err
	}

	u.Id = uint32(id32)
	return u, nil
}

func (r *userRepository) GetUserById(ctx context.Context, id uint32) (entities.User, error) {
	var u entities.User
	var id32 int32

	err := r.db.QueryRow(ctx,
		`SELECT id, phone, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&id32, &u.Phone, &u.CreatedAt)
	if err != nil {
		return entities.User{}, err
	}

	u.Id = uint32(id32)
	return u, nil
}

func (r *userRepository) CreateUser(ctx context.Context, user entities.User) (entities.User, error) {
	var u entities.User
	var id32 int32

	err := r.db.QueryRow(ctx,
		`INSERT INTO users (phone) VALUES ($1) RETURNING id, phone, created_at`,
		user.Phone,
	).Scan(&id32, &u.Phone, &u.CreatedAt)
	if err != nil {
		return entities.User{}, err
	}

	u.Id = uint32(id32)
	return u, nil
}

func (r *userRepository) GetAllUsers(ctx context.Context, skip, limit uint32, phoneSearchTerm *string, creationFrom, creationTo *time.Time) ([]entities.User, error) {
	var sb strings.Builder
	args := make([]any, 0, 5)
	conds := make([]string, 0, 3)
	idx := 1

	sb.WriteString("SELECT id, phone, created_at FROM users")

	if phoneSearchTerm != nil && *phoneSearchTerm != "" {
		conds = append(conds, fmt.Sprintf("phone ILIKE $%d", idx))
		args = append(args, "%"+*phoneSearchTerm+"%")
		idx++
	}
	if creationFrom != nil {
		conds = append(conds, fmt.Sprintf("created_at >= $%d", idx))
		args = append(args, *creationFrom)
		idx++
	}
	if creationTo != nil {
		conds = append(conds, fmt.Sprintf("created_at <= $%d", idx))
		args = append(args, *creationTo)
		idx++
	}

	if len(conds) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(conds, " AND "))
	}

	sb.WriteString(" ORDER BY id ASC")
	sb.WriteString(fmt.Sprintf(" OFFSET $%d LIMIT $%d", idx, idx+1))
	args = append(args, int64(skip), int64(limit))

	rows, err := r.db.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]entities.User, 0, limit)
	for rows.Next() {
		var u entities.User
		var id32 int32
		if err := rows.Scan(&id32, &u.Phone, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.Id = uint32(id32)
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
