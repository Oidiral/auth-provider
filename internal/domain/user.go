package domain

import "time"

type User struct {
	ID           string     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	FirstName    string     `json:"firstName" db:"first_name"`
	LastName     string     `json:"lastName" db:"last_name"`
	Email        string     `json:"email" db:"email"`
	Phone        *string    `json:"phone,omitempty" db:"phone"`
	PasswordHash string     `json:"-" db:"password_hash"`
	IsVerified   bool       `json:"isVerified" db:"is_verified"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt    *time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}

// UpdateUserInput используется для частичного обновления пользователя
type UpdateUserInput struct {
	Email        *string `json:"email,omitempty"`
	Username     *string `json:"username,omitempty"`
	FirstName    *string `json:"firstName,omitempty"`
	LastName     *string `json:"lastName,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	PasswordHash *string `json:"-"`
	IsVerified   *bool   `json:"isVerified,omitempty"`
}
