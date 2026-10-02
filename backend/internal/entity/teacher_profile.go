package entity

import "time"

type TeacherProfile struct {
	ID          int64
	UserID      *int64
	DisplayName string
	Education   string
	Experience  string
	Approach    string
	PhotoURL    string
	IsPublished bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
