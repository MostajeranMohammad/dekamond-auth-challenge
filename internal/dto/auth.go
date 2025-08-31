package dto

type (
	LoginDTO struct {
		Phone string `json:"phone" validate:"required,min=8,max=20"`
	}

	VerifyLoginOTP struct {
		Phone string `json:"phone" validate:"required,min=8,max=20"`
		OTP   string `json:"otp" validate:"required,len=5,numeric"`
	}
)
