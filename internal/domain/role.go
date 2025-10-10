package domain

import "time"

type Role struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type UserRole struct {
	UserID    string    `json:"userId" db:"user_id"`
	RoleID    string    `json:"roleId" db:"role_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
