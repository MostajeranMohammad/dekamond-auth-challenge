package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/config"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/dto"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/entities"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/repositories/mockrepositories"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/usecases/mockusecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuthService_LoginRequestOtp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mockrepositories.NewMockUserRepository(ctrl)
	mockJwtUsecase := mockusecases.NewMockJwtUsecase(ctrl)
	mockOtpUsecase := mockusecases.NewMockOtpUsecase(ctrl)
	cfg := &config.Config{}

	service := NewAuthUsecase(mockUserRepo, mockJwtUsecase, cfg, mockOtpUsecase)

	tests := []struct {
		name       string
		req        dto.LoginDTO
		setupMock  func()
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "successful OTP request",
			req:  dto.LoginDTO{Phone: "+1234567890"},
			setupMock: func() {
				mockOtpUsecase.EXPECT().GenerateOTP().Return("12345", nil)
				mockOtpUsecase.EXPECT().SaveOTP(gomock.Any(), "+1234567890", "12345").Return(nil)
				mockOtpUsecase.EXPECT().SendOtpSms(gomock.Any(), "+1234567890", "12345").Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "invalid phone number (too short)",
			req:        dto.LoginDTO{Phone: "123"},
			setupMock:  func() {},
			wantErr:    true,
			wantErrMsg: "validation",
		},
		{
			name:       "invalid phone number (too long)",
			req:        dto.LoginDTO{Phone: strings.Repeat("1", 25)},
			setupMock:  func() {},
			wantErr:    true,
			wantErrMsg: "validation",
		},
		{
			name:       "empty phone number",
			req:        dto.LoginDTO{Phone: ""},
			setupMock:  func() {},
			wantErr:    true,
			wantErrMsg: "validation",
		},
		{
			name: "OTP generation error",
			req:  dto.LoginDTO{Phone: "+1234567890"},
			setupMock: func() {
				mockOtpUsecase.EXPECT().GenerateOTP().Return("", errors.New("generation failed"))
			},
			wantErr:    true,
			wantErrMsg: "generation failed",
		},
		{
			name: "OTP save error",
			req:  dto.LoginDTO{Phone: "+1234567890"},
			setupMock: func() {
				mockOtpUsecase.EXPECT().GenerateOTP().Return("12345", nil)
				mockOtpUsecase.EXPECT().SaveOTP(gomock.Any(), "+1234567890", "12345").Return(errors.New("save failed"))
			},
			wantErr:    true,
			wantErrMsg: "save failed",
		},
		{
			name: "SMS send error",
			req:  dto.LoginDTO{Phone: "+1234567890"},
			setupMock: func() {
				mockOtpUsecase.EXPECT().GenerateOTP().Return("12345", nil)
				mockOtpUsecase.EXPECT().SaveOTP(gomock.Any(), "+1234567890", "12345").Return(nil)
				mockOtpUsecase.EXPECT().SendOtpSms(gomock.Any(), "+1234567890", "12345").Return(errors.New("SMS failed"))
			},
			wantErr:    true,
			wantErrMsg: "SMS failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := service.LoginRequestOtp(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.wantErrMsg))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_VerifyLoginOTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mockrepositories.NewMockUserRepository(ctrl)
	mockJwtUsecase := mockusecases.NewMockJwtUsecase(ctrl)
	mockOtpUsecase := mockusecases.NewMockOtpUsecase(ctrl)
	cfg := &config.Config{}

	service := NewAuthUsecase(mockUserRepo, mockJwtUsecase, cfg, mockOtpUsecase)

	now := time.Now()
	existingUser := entities.User{
		Id:        123,
		Phone:     "+1234567890",
		CreatedAt: now,
	}

	tests := []struct {
		name       string
		body       dto.VerifyLoginOTP
		setupMock  func()
		wantJWT    string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "successful verification - existing user",
			body: dto.VerifyLoginOTP{Phone: "+1234567890", OTP: "12345"},
			setupMock: func() {
				mockOtpUsecase.EXPECT().VerifyOTP(gomock.Any(), "+1234567890", "12345").Return(nil)
				mockUserRepo.EXPECT().GetUserByPhone(gomock.Any(), "+1234567890").Return(existingUser, nil)
				mockJwtUsecase.EXPECT().GenerateToken(entities.JwtPayload{UserId: 123}).Return("jwt-token-123", nil)
			},
			wantJWT: "jwt-token-123",
			wantErr: false,
		},
		{
			name: "successful verification - new user creation",
			body: dto.VerifyLoginOTP{Phone: "+0987654321", OTP: "54321"},
			setupMock: func() {
				mockOtpUsecase.EXPECT().VerifyOTP(gomock.Any(), "+0987654321", "54321").Return(nil)
				mockUserRepo.EXPECT().GetUserByPhone(gomock.Any(), "+0987654321").Return(entities.User{}, errors.New("no rows"))
				newUser := entities.User{Id: 456, Phone: "+0987654321", CreatedAt: now}
				mockUserRepo.EXPECT().CreateUser(gomock.Any(), entities.User{Phone: "+0987654321"}).Return(newUser, nil)
				mockJwtUsecase.EXPECT().GenerateToken(entities.JwtPayload{UserId: 456}).Return("jwt-token-456", nil)
			},
			wantJWT: "jwt-token-456",
			wantErr: false,
		},
		{
			name: "successful verification - new user creation (not found error)",
			body: dto.VerifyLoginOTP{Phone: "+0987654321", OTP: "54321"},
			setupMock: func() {
				mockOtpUsecase.EXPECT().VerifyOTP(gomock.Any(), "+0987654321", "54321").Return(nil)
				mockUserRepo.EXPECT().GetUserByPhone(gomock.Any(), "+0987654321").Return(entities.User{}, errors.New("user not found"))
				newUser := entities.User{Id: 456, Phone: "+0987654321", CreatedAt: now}
				mockUserRepo.EXPECT().CreateUser(gomock.Any(), entities.User{Phone: "+0987654321"}).Return(newUser, nil)
				mockJwtUsecase.EXPECT().GenerateToken(entities.JwtPayload{UserId: 456}).Return("jwt-token-456", nil)
			},
			wantJWT: "jwt-token-456",
			wantErr: false,
		},
		{
			name:       "invalid phone number",
			body:       dto.VerifyLoginOTP{Phone: "123", OTP: "12345"},
			setupMock:  func() {},
			wantJWT:    "",
			wantErr:    true,
			wantErrMsg: "validation",
		},
		{
			name:       "invalid OTP format",
			body:       dto.VerifyLoginOTP{Phone: "+1234567890", OTP: "123"},
			setupMock:  func() {},
			wantJWT:    "",
			wantErr:    true,
			wantErrMsg: "validation",
		},
		{
			name:       "non-numeric OTP",
			body:       dto.VerifyLoginOTP{Phone: "+1234567890", OTP: "abcde"},
			setupMock:  func() {},
			wantJWT:    "",
			wantErr:    true,
			wantErrMsg: "validation",
		},
		{
			name: "OTP verification failed",
			body: dto.VerifyLoginOTP{Phone: "+1234567890", OTP: "12345"},
			setupMock: func() {
				mockOtpUsecase.EXPECT().VerifyOTP(gomock.Any(), "+1234567890", "12345").Return(errors.New("invalid OTP"))
			},
			wantJWT:    "",
			wantErr:    true,
			wantErrMsg: "invalid OTP",
		},
		{
			name: "user repository error (not user creation case)",
			body: dto.VerifyLoginOTP{Phone: "+1234567890", OTP: "12345"},
			setupMock: func() {
				mockOtpUsecase.EXPECT().VerifyOTP(gomock.Any(), "+1234567890", "12345").Return(nil)
				mockUserRepo.EXPECT().GetUserByPhone(gomock.Any(), "+1234567890").Return(entities.User{}, errors.New("database connection error"))
			},
			wantJWT:    "",
			wantErr:    true,
			wantErrMsg: "database connection error",
		},
		{
			name: "user creation failed",
			body: dto.VerifyLoginOTP{Phone: "+0987654321", OTP: "54321"},
			setupMock: func() {
				mockOtpUsecase.EXPECT().VerifyOTP(gomock.Any(), "+0987654321", "54321").Return(nil)
				mockUserRepo.EXPECT().GetUserByPhone(gomock.Any(), "+0987654321").Return(entities.User{}, errors.New("no rows"))
				mockUserRepo.EXPECT().CreateUser(gomock.Any(), entities.User{Phone: "+0987654321"}).Return(entities.User{}, errors.New("creation failed"))
			},
			wantJWT:    "",
			wantErr:    true,
			wantErrMsg: "creation failed",
		},
		{
			name: "JWT generation failed",
			body: dto.VerifyLoginOTP{Phone: "+1234567890", OTP: "12345"},
			setupMock: func() {
				mockOtpUsecase.EXPECT().VerifyOTP(gomock.Any(), "+1234567890", "12345").Return(nil)
				mockUserRepo.EXPECT().GetUserByPhone(gomock.Any(), "+1234567890").Return(existingUser, nil)
				mockJwtUsecase.EXPECT().GenerateToken(entities.JwtPayload{UserId: 123}).Return("", errors.New("JWT generation failed"))
			},
			wantJWT:    "",
			wantErr:    true,
			wantErrMsg: "JWT generation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			jwt, err := service.VerifyLoginOTP(context.Background(), tt.body)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, jwt)
				if tt.wantErrMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.wantErrMsg))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantJWT, jwt)
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mockrepositories.NewMockUserRepository(ctrl)
	mockJwtUsecase := mockusecases.NewMockJwtUsecase(ctrl)
	mockOtpUsecase := mockusecases.NewMockOtpUsecase(ctrl)
	cfg := &config.Config{}

	service := NewAuthUsecase(mockUserRepo, mockJwtUsecase, cfg, mockOtpUsecase)

	now := time.Now()
	testUser := entities.User{
		Id:        123,
		Phone:     "+1234567890",
		CreatedAt: now,
	}

	tests := []struct {
		name       string
		token      string
		setupMock  func()
		wantUser   entities.User
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:  "successful token validation",
			token: "valid-jwt-token",
			setupMock: func() {
				mockJwtUsecase.EXPECT().ValidateToken("valid-jwt-token").Return(entities.JwtPayload{UserId: 123}, nil)
				mockUserRepo.EXPECT().GetUserById(gomock.Any(), uint32(123)).Return(testUser, nil)
			},
			wantUser: testUser,
			wantErr:  false,
		},
		{
			name:  "successful token validation with Bearer prefix",
			token: "Bearer valid-jwt-token",
			setupMock: func() {
				mockJwtUsecase.EXPECT().ValidateToken("valid-jwt-token").Return(entities.JwtPayload{UserId: 123}, nil)
				mockUserRepo.EXPECT().GetUserById(gomock.Any(), uint32(123)).Return(testUser, nil)
			},
			wantUser: testUser,
			wantErr:  false,
		},
		{
			name:  "successful token validation with bearer prefix (lowercase)",
			token: "bearer valid-jwt-token",
			setupMock: func() {
				mockJwtUsecase.EXPECT().ValidateToken("valid-jwt-token").Return(entities.JwtPayload{UserId: 123}, nil)
				mockUserRepo.EXPECT().GetUserById(gomock.Any(), uint32(123)).Return(testUser, nil)
			},
			wantUser: testUser,
			wantErr:  false,
		},
		{
			name:  "token with extra spaces",
			token: "  Bearer   valid-jwt-token  ",
			setupMock: func() {
				mockJwtUsecase.EXPECT().ValidateToken("valid-jwt-token").Return(entities.JwtPayload{UserId: 123}, nil)
				mockUserRepo.EXPECT().GetUserById(gomock.Any(), uint32(123)).Return(testUser, nil)
			},
			wantUser: testUser,
			wantErr:  false,
		},
		{
			name:       "empty token",
			token:      "",
			setupMock:  func() {},
			wantUser:   entities.User{},
			wantErr:    true,
			wantErrMsg: "missing token",
		},
		{
			name:       "only spaces token",
			token:      "   ",
			setupMock:  func() {},
			wantUser:   entities.User{},
			wantErr:    true,
			wantErrMsg: "missing token",
		},
		{
			name:  "only Bearer without token",
			token: "Bearer",
			setupMock: func() {
				// "Bearer" without space doesn't match "bearer " pattern, so it gets validated as-is
				mockJwtUsecase.EXPECT().ValidateToken("Bearer").Return(entities.JwtPayload{}, errors.New("invalid token"))
			},
			wantUser:   entities.User{},
			wantErr:    true,
			wantErrMsg: "invalid token",
		},
		{
			name:  "Bearer with only spaces",
			token: "Bearer   ",
			setupMock: func() {
				// "Bearer   " gets trimmed to "Bearer", which doesn't match "bearer " pattern
				mockJwtUsecase.EXPECT().ValidateToken("Bearer").Return(entities.JwtPayload{}, errors.New("invalid token"))
			},
			wantUser:   entities.User{},
			wantErr:    true,
			wantErrMsg: "invalid token",
		},
		{
			name:  "JWT validation failed",
			token: "invalid-jwt-token",
			setupMock: func() {
				mockJwtUsecase.EXPECT().ValidateToken("invalid-jwt-token").Return(entities.JwtPayload{}, errors.New("invalid token"))
			},
			wantUser:   entities.User{},
			wantErr:    true,
			wantErrMsg: "invalid token",
		},
		{
			name:  "user not found",
			token: "valid-jwt-token",
			setupMock: func() {
				mockJwtUsecase.EXPECT().ValidateToken("valid-jwt-token").Return(entities.JwtPayload{UserId: 999}, nil)
				mockUserRepo.EXPECT().GetUserById(gomock.Any(), uint32(999)).Return(entities.User{}, errors.New("user not found"))
			},
			wantUser:   entities.User{},
			wantErr:    true,
			wantErrMsg: "user not found",
		},
		{
			name:  "database error during user retrieval",
			token: "valid-jwt-token",
			setupMock: func() {
				mockJwtUsecase.EXPECT().ValidateToken("valid-jwt-token").Return(entities.JwtPayload{UserId: 123}, nil)
				mockUserRepo.EXPECT().GetUserById(gomock.Any(), uint32(123)).Return(entities.User{}, errors.New("database error"))
			},
			wantUser:   entities.User{},
			wantErr:    true,
			wantErrMsg: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			user, err := service.ValidateToken(context.Background(), tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, entities.User{}, user)
				if tt.wantErrMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.wantErrMsg))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, user)
			}
		})
	}
}

func TestAuthService_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mockrepositories.NewMockUserRepository(ctrl)
	mockJwtUsecase := mockusecases.NewMockJwtUsecase(ctrl)
	mockOtpUsecase := mockusecases.NewMockOtpUsecase(ctrl)
	cfg := &config.Config{}

	service := NewAuthUsecase(mockUserRepo, mockJwtUsecase, cfg, mockOtpUsecase)

	phone := "+1234567890"
	otp := "12345"
	now := time.Now()

	// Full flow integration test
	t.Run("complete authentication flow", func(t *testing.T) {
		// Step 1: Request OTP
		mockOtpUsecase.EXPECT().GenerateOTP().Return(otp, nil)
		mockOtpUsecase.EXPECT().SaveOTP(gomock.Any(), phone, otp).Return(nil)
		mockOtpUsecase.EXPECT().SendOtpSms(gomock.Any(), phone, otp).Return(nil)

		err := service.LoginRequestOtp(context.Background(), dto.LoginDTO{Phone: phone})
		require.NoError(t, err)

		// Step 2: Verify OTP and create new user
		mockOtpUsecase.EXPECT().VerifyOTP(gomock.Any(), phone, otp).Return(nil)
		mockUserRepo.EXPECT().GetUserByPhone(gomock.Any(), phone).Return(entities.User{}, errors.New("no rows"))

		newUser := entities.User{Id: 123, Phone: phone, CreatedAt: now}
		mockUserRepo.EXPECT().CreateUser(gomock.Any(), entities.User{Phone: phone}).Return(newUser, nil)

		jwtToken := "jwt-token-123"
		mockJwtUsecase.EXPECT().GenerateToken(entities.JwtPayload{UserId: 123}).Return(jwtToken, nil)

		token, err := service.VerifyLoginOTP(context.Background(), dto.VerifyLoginOTP{
			Phone: phone,
			OTP:   otp,
		})
		require.NoError(t, err)
		assert.Equal(t, jwtToken, token)

		// Step 3: Validate the token
		mockJwtUsecase.EXPECT().ValidateToken(jwtToken).Return(entities.JwtPayload{UserId: 123}, nil)
		mockUserRepo.EXPECT().GetUserById(gomock.Any(), uint32(123)).Return(newUser, nil)

		user, err := service.ValidateToken(context.Background(), token)
		require.NoError(t, err)
		assert.Equal(t, newUser, user)
	})

	t.Run("existing user login flow", func(t *testing.T) {
		existingUser := entities.User{Id: 456, Phone: phone, CreatedAt: now.Add(-time.Hour)}

		// Step 1: Request OTP
		mockOtpUsecase.EXPECT().GenerateOTP().Return(otp, nil)
		mockOtpUsecase.EXPECT().SaveOTP(gomock.Any(), phone, otp).Return(nil)
		mockOtpUsecase.EXPECT().SendOtpSms(gomock.Any(), phone, otp).Return(nil)

		err := service.LoginRequestOtp(context.Background(), dto.LoginDTO{Phone: phone})
		require.NoError(t, err)

		// Step 2: Verify OTP for existing user
		mockOtpUsecase.EXPECT().VerifyOTP(gomock.Any(), phone, otp).Return(nil)
		mockUserRepo.EXPECT().GetUserByPhone(gomock.Any(), phone).Return(existingUser, nil)

		jwtToken := "jwt-token-456"
		mockJwtUsecase.EXPECT().GenerateToken(entities.JwtPayload{UserId: 456}).Return(jwtToken, nil)

		token, err := service.VerifyLoginOTP(context.Background(), dto.VerifyLoginOTP{
			Phone: phone,
			OTP:   otp,
		})
		require.NoError(t, err)
		assert.Equal(t, jwtToken, token)

		// Step 3: Validate the token with Bearer prefix
		mockJwtUsecase.EXPECT().ValidateToken(jwtToken).Return(entities.JwtPayload{UserId: 456}, nil)
		mockUserRepo.EXPECT().GetUserById(gomock.Any(), uint32(456)).Return(existingUser, nil)

		user, err := service.ValidateToken(context.Background(), "Bearer "+token)
		require.NoError(t, err)
		assert.Equal(t, existingUser, user)
	})
}
