package usecases

import (
	"context"
	"errors"
	"strings"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/config"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/dto"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/entities"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/repositories"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/pkg/utils"
)

type authService struct {
	userRepository repositories.UserRepository
	jwtUsecase     JwtUsecase
	cfg            *config.Config
	otpUsecase     OtpUsecase
}

func NewAuthUsecase(
	userRepository repositories.UserRepository,
	jwtUsecase JwtUsecase,
	cfg *config.Config,
	otpUsecase OtpUsecase,
) AuthService {
	return &authService{
		userRepository: userRepository,
		jwtUsecase:     jwtUsecase,
		cfg:            cfg,
		otpUsecase:     otpUsecase,
	}
}

func (a *authService) LoginRequestOtp(ctx context.Context, req dto.LoginDTO) (err error) {
	if err := utils.ValidateStruct(req); err != nil {
		return err
	}
	// generate and save otp
	code, err := a.otpUsecase.GenerateOTP()
	if err != nil {
		return err
	}
	if err := a.otpUsecase.SaveOTP(ctx, req.Phone, code); err != nil {
		return err
	}
	return a.otpUsecase.SendOtpSms(ctx, req.Phone, code)
}

func (a *authService) VerifyLoginOTP(ctx context.Context, body dto.VerifyLoginOTP) (jwt string, err error) {
	if err := utils.ValidateStruct(body); err != nil {
		return "", err
	}
	if err := a.otpUsecase.VerifyOTP(ctx, body.Phone, body.OTP); err != nil {
		return "", err
	}
	// find or create user
	user, err := a.userRepository.GetUserByPhone(ctx, body.Phone)
	if err != nil {
		// try create if not found
		if strings.Contains(err.Error(), "no rows") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			user, err = a.userRepository.CreateUser(ctx, entities.User{Phone: body.Phone})
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	token, err := a.jwtUsecase.GenerateToken(entities.JwtPayload{UserId: user.Id})
	if err != nil {
		return "", err
	}
	return token, nil
}

func (a *authService) ValidateToken(ctx context.Context, token string) (entities.User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return entities.User{}, errors.New("missing token")
	}
	// Allow header with Bearer prefix
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	payload, err := a.jwtUsecase.ValidateToken(token)
	if err != nil {
		return entities.User{}, err
	}
	return a.userRepository.GetUserById(ctx, payload.UserId)
}
