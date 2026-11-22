package domain

import "errors"

var ErrUsernameTaken = errors.New("username already exists")