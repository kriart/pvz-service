package user

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email уже зарегистрирован")
	ErrInvalidCredentials = errors.New("неправильный email или пароль")
)
