package entity

import "time"

type Role string

const (
	RoleTeacher Role = "teacher"
	RoleAdmin   Role = "admin"
)

func (role Role) Valid() bool {
	return role == RoleTeacher || role == RoleAdmin
}

type User struct {
	ID        int64
	Email     string
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
