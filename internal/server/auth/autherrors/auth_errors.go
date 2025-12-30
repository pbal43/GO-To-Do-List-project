package autherrors

import "errors"

var (
	ErrInvalidAccessToken        = errors.New("invalid access token")
	ErrInvalidRefreshToken       = errors.New("invalid access token")
	ErrMissingAccessToken        = errors.New("missing access token")
	ErrMissingRefreshToken       = errors.New("missing refresh token")
	ErrFailToParseNewAccessToken = errors.New("failed to parse new access token")
)
