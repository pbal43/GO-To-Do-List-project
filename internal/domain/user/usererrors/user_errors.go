package usererrors

import (
	"errors"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserEmptyInsert    = errors.New("empty insert")
	ErrUserIsAlreadyExist = errors.New("user is already exist")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserNotExist       = errors.New("user not exists")
	ErrNotValidCreds      = errors.New("the creds are invalid")
	ErrInvalidChar        = errors.New(`invalid character 'n' looking for beginning of value`)
	ErrInternalServer     = errors.New(`internal server error`)
	ErrGetAllUsersData    = errors.New("can't get all users data")
)
