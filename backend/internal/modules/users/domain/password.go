package users_domain

import "strings"

const PasswordMinLen = 8

func ValidatePassword(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return ErrEmptyPassword
	}
	if len(raw) < PasswordMinLen {
		return ErrPasswordTooShort
	}
	return nil
}
