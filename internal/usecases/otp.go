package usecases

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type otp struct {
	redisClient *redis.Client
	l           logger.Logger
}

func NewOtpUsecase(redisClient *redis.Client, l logger.Logger) OtpUsecase {
	return &otp{
		redisClient: redisClient,
		l:           l,
	}
}

func (o *otp) GenerateOTP() (string, error) {
	num, err := rand.Int(rand.Reader, big.NewInt(90000))
	if err != nil {
		return "", err
	}
	otpCode := fmt.Sprintf("%05d", num.Int64()+10000)
	return otpCode, nil
}

func (o *otp) SendOtpSms(ctx context.Context, phoneNumber string, otpCode string) error {
	if err := o.validateOtpRateLimit(ctx, phoneNumber); err != nil {
		return err
	}

	o.l.Info(fmt.Sprintf("Sending OTP %s to phone number %s", otpCode, phoneNumber))
	return nil
}

func (o *otp) SaveOTP(ctx context.Context, phoneNumber string, otpCode string) error {
	key := fmt.Sprintf("otp:code:%s", phoneNumber)
	if err := o.redisClient.Set(ctx, key, otpCode, 2*time.Minute).Err(); err != nil {
		return err
	}
	return nil
}

func (o *otp) VerifyOTP(ctx context.Context, phoneNumber string, otpCode string) error {
	key := fmt.Sprintf("otp:code:%s", phoneNumber)
	val, err := o.redisClient.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("invalid or expired otp")
	}
	if val != otpCode {
		return fmt.Errorf("invalid or expired otp")
	}
	// consume OTP
	_ = o.redisClient.Del(ctx, key).Err()
	return nil
}

func (o *otp) validateOtpRateLimit(ctx context.Context, phoneNumber string) error {
	key := fmt.Sprintf("otp:10m:%s", phoneNumber)

	count, err := o.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return err
	}

	if count == 1 {
		o.redisClient.Expire(ctx, key, 10*time.Minute)
	}

	if count > 3 {
		return fmt.Errorf("rate limit exceeded: max 3 OTPs per 10 minutes")
	}

	return nil
}
