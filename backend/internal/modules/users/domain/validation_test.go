package users_domain

import (
	"errors"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want error
	}{
		{name: "empty", in: "", want: ErrEmptyPassword},
		{name: "short", in: "1234567", want: ErrPasswordTooShort},
		{name: "valid", in: "12345678", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.in)
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
		})
	}
}

func TestEmailAndLoginValidation(t *testing.T) {
	if email, err := NewEmail("USER@Example.COM "); err != nil || string(*email) != "user@example.com" {
		t.Fatalf("expected normalized email, got %v err=%v", email, err)
	}
	if _, err := NewEmail("bad"); !errors.Is(err, ErrInvalidEmailFormat) {
		t.Fatalf("expected invalid email, got %v", err)
	}
	if login, err := NewLogin(" user_1 "); err != nil || string(*login) != "user_1" {
		t.Fatalf("expected trimmed login, got %v err=%v", login, err)
	}
	if _, err := NewLogin("юзер"); !errors.Is(err, ErrInvalidLoginFormat) {
		t.Fatalf("expected invalid login, got %v", err)
	}
}
