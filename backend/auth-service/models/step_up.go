package models

// VerifyPasswordRequest is the body of POST /auth/verify-password.
type VerifyPasswordRequest struct {
	Password string `json:"password" binding:"required" validate:"required"`
	// Purpose must be whitelisted (only "org_levels.change" today).
	Purpose string `json:"purpose" binding:"required" validate:"required"`
}

// VerifyPasswordResponse carries the one-time step-up token. The client sends
// it as X-Step-Up-Token to the protected endpoint within ExpiresIn seconds.
type VerifyPasswordResponse struct {
	StepUpToken string `json:"step_up_token"`
	ExpiresIn   int    `json:"expires_in"`
}
