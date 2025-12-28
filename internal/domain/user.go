package domain

import "time"

type User struct {
	ID           string     `db:"id"`
	Username     string     `db:"username"`
	FirstName    string     `db:"first_name"`
	LastName     string     `db:"last_name"`
	Email        string     `db:"email"`
	Phone        *string    `db:"phone"`
	PasswordHash string     `db:"password_hash"`
	IsVerified   bool       `db:"is_verified"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

// UpdateUserInput используется для частичного обновления пользователя
type UpdateUserInput struct {
	Email        *string
	Username     *string
	FirstName    *string
	LastName     *string
	Phone        *string
	PasswordHash *string
	IsVerified   *bool
}
