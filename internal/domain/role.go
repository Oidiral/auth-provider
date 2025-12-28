package domain

import "time"

type Role struct {
	ID          int       `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type UserRole struct {
	UserID    string    `db:"user_id"`
	RoleID    int       `db:"role_id"`
	CreatedAt time.Time `db:"created_at"`
}
