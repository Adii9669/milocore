package requests

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type VerifyOtpRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp"   validate:"required,len=6"`
}
