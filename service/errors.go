package service

import "errors"

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrWrongPassword      = errors.New("old password is incorrect")
	ErrUserNotFound       = errors.New("user not found")
)
