package entity

import "time"

type Role string

const (
	RoleAdmin Role = "admin"
)

func (role Role) Valid() bool {
	return role == RoleAdmin
}

type User struct {
	ID        int64
	Username  string
	Roles     []Role
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Credentials struct {
	User
	PasswordHash string
}

type Session struct {
	Token     string
	ExpiresAt time.Time
}
