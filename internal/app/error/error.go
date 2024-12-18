// error -- описывает ошибки аутентификации
package error

import "errors"

// ErrLoginAlreadyTaken -- такой логин уже существует
var ErrLoginAlreadyTaken = errors.New("login already taken")

// ErrInvalidLoginOrPassword -- не валидный логин или парол�
var ErrInvalidLoginOrPassword = errors.New("invalid login or password")
