package domain

import "errors"

var (
	ErrUsernameTaken   = errors.New("username already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
)
