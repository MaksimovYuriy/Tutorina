package entity

import "time"

type Session struct {
	Token     string
	ExpiresAt time.Time
}
