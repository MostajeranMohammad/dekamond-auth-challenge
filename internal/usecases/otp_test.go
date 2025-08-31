package usecases

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockLogger is a mock for logger.Logger interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(message string, args ...interface{}) {
	m.Called(message, args)
}

func (m *MockLogger) Error(message interface{}, args ...interface{}) {
	m.Called(message, args)
}

func (m *MockLogger) Debug(message interface{}, args ...interface{}) {
	m.Called(message, args)
}

func (m *MockLogger) Warn(message string, args ...interface{}) {
	m.Called(message, args)
}

func (m *MockLogger) Fatal(message interface{}, args ...interface{}) {
	m.Called(message, args)
}

func TestOtpUsecase_GenerateOTP(t *testing.T) {
	// We can test GenerateOTP directly since it doesn't depend on external services

	// Create a mock logger (not actually used in GenerateOTP but needed for struct)
	mockLogger := &MockLogger{}

	o := &otp{
		redisClient: nil, // Not used in GenerateOTP
		l:           mockLogger,
	}

	t.Run("successful OTP generation", func(t *testing.T) {
		otpCode, err := o.GenerateOTP()

		assert.NoError(t, err)
		assert.Len(t, otpCode, 5)
		assert.Regexp(t, `^\d{5}$`, otpCode) // Should be 5 digits

		// OTP should be between 10000 and 99999
		assert.GreaterOrEqual(t, otpCode, "10000")
		assert.LessOrEqual(t, otpCode, "99999")
	})

	t.Run("OTP generation produces different codes", func(t *testing.T) {
		codes := make(map[string]bool)

		// Generate multiple OTPs and ensure they're different
		for i := 0; i < 10; i++ {
			otpCode, err := o.GenerateOTP()
			assert.NoError(t, err)
			codes[otpCode] = true
		}

		// Most should be unique (allowing for rare duplicates due to randomness)
		assert.GreaterOrEqual(t, len(codes), 8)
	})
}

// For the other tests that require Redis interaction, we'll create integration-style tests
// that can be run with a real Redis instance, or we'll test the logic separately

func TestOtpUsecase_GenerateOTPEdgeCases(t *testing.T) {
	mockLogger := &MockLogger{}

	o := &otp{
		redisClient: nil,
		l:           mockLogger,
	}

	// Test that all generated OTPs are valid
	for i := 0; i < 100; i++ {
		otpCode, err := o.GenerateOTP()
		require.NoError(t, err)
		require.Len(t, otpCode, 5)
		require.Regexp(t, `^\d{5}$`, otpCode)

		// Parse as integer to ensure it's in valid range
		require.GreaterOrEqual(t, otpCode, "10000")
		require.LessOrEqual(t, otpCode, "99999")
	}
}

// TestOtpUsecaseWithRealRedis tests the OTP usecase with a real Redis connection
// This test requires a Redis instance running on localhost:6379
// To run: go test -tags=integration ./internal/usecases/...
func TestOtpUsecaseWithRealRedis(t *testing.T) {
	// Skip this test in CI or when Redis is not available
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use test database
	})

	// Test if Redis is available
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis is not available, skipping integration test")
	}

	// Clean up test data
	defer func() {
		redisClient.FlushDB(ctx)
		redisClient.Close()
	}()

	mockLogger := &MockLogger{}
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Maybe()

	o := NewOtpUsecase(redisClient, mockLogger)

	phoneNumber := "+1234567890"

	t.Run("complete OTP flow with real Redis", func(t *testing.T) {
		// Generate OTP
		otpCode, err := o.GenerateOTP()
		require.NoError(t, err)
		require.Len(t, otpCode, 5)

		// Save OTP
		err = o.SaveOTP(ctx, phoneNumber, otpCode)
		require.NoError(t, err)

		// Verify correct OTP
		err = o.VerifyOTP(ctx, phoneNumber, otpCode)
		require.NoError(t, err)

		// Try to verify again (should fail because OTP is consumed)
		err = o.VerifyOTP(ctx, phoneNumber, otpCode)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired otp")
	})

	t.Run("verify wrong OTP", func(t *testing.T) {
		// Generate and save OTP
		otpCode, err := o.GenerateOTP()
		require.NoError(t, err)

		err = o.SaveOTP(ctx, phoneNumber+"_wrong", otpCode)
		require.NoError(t, err)

		// Try to verify with wrong OTP
		err = o.VerifyOTP(ctx, phoneNumber+"_wrong", "00000")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired otp")
	})

	t.Run("rate limiting", func(t *testing.T) {
		testPhone := "+1111111111"

		// First 3 requests should succeed
		for i := 0; i < 3; i++ {
			err := o.SendOtpSms(ctx, testPhone, "12345")
			assert.NoError(t, err, "Request %d should succeed", i+1)
		}

		// 4th request should fail due to rate limiting
		err := o.SendOtpSms(ctx, testPhone, "12345")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")
	})
}
