package security

import (
	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher struct{}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

func (h *PasswordHasher) Hash(password string) (ud.PasswordHash, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ud.PasswordHash{}, err
	}
	return ud.PasswordHash{
		Algo:  ud.PasswordHashBcrypt,
		Value: string(hash),
	}, nil
}

func (h *PasswordHasher) Compare(password string, hash ud.PasswordHash) bool {
	if hash.Algo != ud.PasswordHashBcrypt {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash.Value), []byte(password)) == nil
}
