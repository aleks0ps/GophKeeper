// Package util -- пакет реализует функции для херширования пароля пользователя и его проверки
package util

import (
	"golang.org/x/crypto/bcrypt"
)

func Hash(s string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash -- функция проверят пароль пользователья при аутентификации
func CheckPasswordHash(hash string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}
