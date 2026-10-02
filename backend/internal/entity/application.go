package entity

import "time"

const (
	ApplicationStatusNew       = "new"
	ApplicationStatusAccepted  = "accepted"
	ApplicationStatusRejected  = "rejected"
	ApplicationStatusCompleted = "completed"
)

type Application struct {
	ID                 int64
	LessonID           int64
	OfferTitle         string
	TeacherProfileID   int64
	TeacherDisplayName string
	LessonStartsAt     time.Time
	LessonCapacity     int
	FullName           string
	Phone              string
	Email              string
	Comment            string
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
