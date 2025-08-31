package usecases

import (
	"context"
	"time"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/dto"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/entities"
)

type (
	AuthService interface {
		LoginRequestOtp(ctx context.Context, req dto.LoginDTO) (err error)
		VerifyLoginOTP(ctx context.Context, body dto.VerifyLoginOTP) (jwt string, err error)
		ValidateToken(ctx context.Context, token string) (entities.User, error)
	}

	OtpUsecase interface {
		SendOtpSms(ctx context.Context, phone string, otp string) error
		GenerateOTP() (string, error)
		SaveOTP(ctx context.Context, phone string, otp string) error
		VerifyOTP(ctx context.Context, phone string, otp string) error
	}

	JwtUsecase interface {
		GenerateToken(payload entities.JwtPayload) (jwt string, err error)
		ValidateToken(token string) (entities.JwtPayload, error)
		GenerateRefreshToken(payload entities.JwtPayload) (refreshToken string, err error)
	}

	UsersService interface {
		GetUser(ctx context.Context, id uint32) (entities.User, error)
		GetAllUsers(ctx context.Context, page, limit uint32, phoneSearchTerm *string, creationFrom, creationTo *time.Time) ([]entities.User, error)
	}
)
