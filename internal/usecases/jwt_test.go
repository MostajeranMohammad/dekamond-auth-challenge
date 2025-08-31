package usecases

import (
	"testing"
	"time"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/entities"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwtUsecase_GenerateToken(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		payload   entities.JwtPayload
		wantErr   bool
	}{
		{
			name:      "successful token generation",
			secretKey: "test-secret-key",
			payload:   entities.JwtPayload{UserId: 123},
			wantErr:   false,
		},
		{
			name:      "successful token generation with different user id",
			secretKey: "test-secret-key",
			payload:   entities.JwtPayload{UserId: 456},
			wantErr:   false,
		},
		{
			name:      "empty secret key should still work",
			secretKey: "",
			payload:   entities.JwtPayload{UserId: 789},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := NewJwtUsecase(tt.secretKey)

			token, err := j.GenerateToken(tt.payload)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify the token can be parsed
				parsedToken, parseErr := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte(tt.secretKey), nil
				})
				assert.NoError(t, parseErr)
				assert.True(t, parsedToken.Valid)

				// Verify claims
				if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
					assert.Equal(t, float64(tt.payload.UserId), claims["userId"])
					assert.NotZero(t, claims["exp"])

					// Verify expiration is approximately 1 hour from now
					exp := int64(claims["exp"].(float64))
					expectedExp := time.Now().Add(time.Hour).Unix()
					assert.InDelta(t, expectedExp, exp, 60) // Allow 1 minute tolerance
				}
			}
		})
	}
}

func TestJwtUsecase_ValidateToken(t *testing.T) {
	secretKey := "test-secret-key"
	j := NewJwtUsecase(secretKey)

	tests := []struct {
		name        string
		tokenString string
		wantPayload entities.JwtPayload
		wantErr     bool
		setupToken  func() string
	}{
		{
			name: "valid token",
			setupToken: func() string {
				token, _ := j.GenerateToken(entities.JwtPayload{UserId: 123})
				return token
			},
			wantPayload: entities.JwtPayload{UserId: 123},
			wantErr:     false,
		},
		{
			name:        "invalid token format",
			tokenString: "invalid-token",
			wantErr:     true,
		},
		{
			name:        "empty token",
			tokenString: "",
			wantErr:     true,
		},
		{
			name: "expired token",
			setupToken: func() string {
				// Create an expired token
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"userId": float64(123),
					"exp":    time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
				})
				tokenString, _ := token.SignedString([]byte(secretKey))
				return tokenString
			},
			wantErr: true,
		},
		{
			name: "refresh token should be rejected",
			setupToken: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"userId":  float64(123),
					"exp":     time.Now().Add(time.Hour).Unix(),
					"refresh": true,
				})
				tokenString, _ := token.SignedString([]byte(secretKey))
				return tokenString
			},
			wantErr: true,
		},
		{
			name: "token with wrong secret",
			setupToken: func() string {
				wrongSecretJwt := NewJwtUsecase("wrong-secret")
				token, _ := wrongSecretJwt.GenerateToken(entities.JwtPayload{UserId: 123})
				return token
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := tt.tokenString
			if tt.setupToken != nil {
				tokenString = tt.setupToken()
			}

			payload, err := j.ValidateToken(tokenString)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Zero(t, payload)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPayload, payload)
			}
		})
	}
}

func TestJwtUsecase_GenerateRefreshToken(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		payload   entities.JwtPayload
		wantErr   bool
	}{
		{
			name:      "successful refresh token generation",
			secretKey: "test-secret-key",
			payload:   entities.JwtPayload{UserId: 123},
			wantErr:   false,
		},
		{
			name:      "successful refresh token generation with different user id",
			secretKey: "test-secret-key",
			payload:   entities.JwtPayload{UserId: 456},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := NewJwtUsecase(tt.secretKey)

			token, err := j.GenerateRefreshToken(tt.payload)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify the token can be parsed
				parsedToken, parseErr := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte(tt.secretKey), nil
				})
				assert.NoError(t, parseErr)
				assert.True(t, parsedToken.Valid)

				// Verify claims
				if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
					assert.Equal(t, float64(tt.payload.UserId), claims["userId"])
					assert.Equal(t, true, claims["refresh"])
					assert.NotZero(t, claims["exp"])

					// Verify expiration is approximately 7 days from now
					exp := int64(claims["exp"].(float64))
					expectedExp := time.Now().Add(time.Hour * 24 * 7).Unix()
					assert.InDelta(t, expectedExp, exp, 60) // Allow 1 minute tolerance
				}
			}
		})
	}
}

func TestJwtUsecase_Integration(t *testing.T) {
	secretKey := "integration-test-secret"
	j := NewJwtUsecase(secretKey)

	// Test normal token flow
	payload := entities.JwtPayload{UserId: 999}

	// Generate token
	token, err := j.GenerateToken(payload)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Validate token
	validatedPayload, err := j.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, payload, validatedPayload)

	// Generate refresh token
	refreshToken, err := j.GenerateRefreshToken(payload)
	require.NoError(t, err)
	require.NotEmpty(t, refreshToken)

	// Refresh token should not be valid for normal validation
	_, err = j.ValidateToken(refreshToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")

	// Normal token and refresh token should be different
	assert.NotEqual(t, token, refreshToken)
}
