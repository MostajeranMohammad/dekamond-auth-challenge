package usecases

import (
	"time"

	"errors"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/entities"
	"github.com/golang-jwt/jwt/v5"
)

type jwtUsecase struct {
	jwtSecretKey string
}

func NewJwtUsecase(jwtSecretKey string) JwtUsecase {
	return &jwtUsecase{
		jwtSecretKey: jwtSecretKey,
	}
}

func (j *jwtUsecase) GenerateToken(payload entities.JwtPayload) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": payload.UserId,
		"exp":    time.Now().Add(time.Hour * 1).Unix(),
	})
	tokenString, err := token.SignedString([]byte(j.jwtSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j *jwtUsecase) ValidateToken(tokenString string) (entities.JwtPayload, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.jwtSecretKey), nil
	})
	if err != nil {
		return entities.JwtPayload{}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId := claims["userId"].(float64)
		refresh, ok := claims["refresh"].(bool)
		if ok && refresh {
			return entities.JwtPayload{}, errors.New("invalid token")
		}
		return entities.JwtPayload{
			UserId: uint32(userId),
		}, nil
	} else {
		return entities.JwtPayload{}, err
	}
}

func (j *jwtUsecase) GenerateRefreshToken(payload entities.JwtPayload) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":  payload.UserId,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
		"refresh": true,
	})
	refreshTokenString, err := refreshToken.SignedString([]byte(j.jwtSecretKey))
	if err != nil {
		return "", err
	}

	return refreshTokenString, nil
}
