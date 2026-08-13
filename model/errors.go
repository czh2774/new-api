package model

import "errors"

// Common errors
var (
	ErrDatabase = errors.New("database error")
)

// Option update errors
var (
	ErrOptionCASInvalid  = errors.New("option compare-and-swap is invalid")
	ErrOptionCASNoop     = errors.New("option compare-and-swap is a no-op")
	ErrOptionCASConflict = errors.New("option changed concurrently or is not persisted")
)

// User auth errors
var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserEmptyCredentials = errors.New("empty credentials")
	ErrEmailAlreadyTaken    = errors.New("email already taken")
	ErrEmailNotFound        = errors.New("email not found")
	ErrEmailAmbiguous       = errors.New("email matches multiple users")
)

// Token auth errors
var (
	ErrTokenNotProvided                 = errors.New("token not provided")
	ErrTokenInvalid                     = errors.New("token invalid")
	ErrTokenModelLimitsInvalid          = errors.New("token model limits update is invalid")
	ErrTokenModelLimitsNoop             = errors.New("token model limits update is a no-op")
	ErrTokenModelLimitsConflict         = errors.New("token model limits changed concurrently")
	ErrTokenModelLimitsCacheUnsupported = errors.New("token model limits update is unavailable while Redis token caching is enabled")
)

// Channel routing CAS errors
var (
	ErrChannelRoutingCASInvalid          = errors.New("channel routing update is invalid")
	ErrChannelRoutingCASNoop             = errors.New("channel routing update is a no-op")
	ErrChannelRoutingCASConflict         = errors.New("channel routing changed concurrently")
	ErrChannelRoutingCASCacheUnsupported = errors.New("channel routing update is unavailable while memory caching is enabled")
)

// Redemption errors
var ErrRedeemFailed = errors.New("redeem.failed")

// 2FA errors
var ErrTwoFANotEnabled = errors.New("2fa not enabled")
var ErrTwoFAAlreadyEnabled = errors.New("2fa already enabled")
