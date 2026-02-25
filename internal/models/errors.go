package models

import "errors"

var (
	BcryptError            = errors.New(`models: bcrypt hashing failed`)
	EmailAlreadyExist      = errors.New(`models: email already exist`)
	TokenValidationError   = errors.New("models: registration token validation failed")
	TokenMissMatchError    = errors.New("models: registration token mismatch")
	TokenExpiredError      = errors.New("models: registration token is expired")
	CreationError          = errors.New("models: error creating record")
	ErrInvalidCredentials  = errors.New("models: invalid credentials")
	ErrUserNotFound        = errors.New("models: user not found")
	ErrChurchNotFound      = errors.New("models: church not found")
	ErrInvitationNotFound  = errors.New("models: invitation not found")
	ErrInvitationExpired   = errors.New("models: invitation has expired")
	ErrInvitationAccepted  = errors.New("models: invitation already accepted")
	ErrRecordNotFound      = errors.New("models: record not found")
)
