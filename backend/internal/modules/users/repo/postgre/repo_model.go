package postgre

import "time"

type userModel struct {
	ID                string    `db:"id"`
	Login             string    `db:"login"`
	Email             string    `db:"email"`
	Role              string    `db:"role"`
	PasswordHash      string    `db:"password_hash"`
	PasswordHashAlgo  string    `db:"password_hash_algo"`
	PasswordChangedAt time.Time `db:"password_changed_at"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

func (userModel) TableName() string {
	return "users"
}
