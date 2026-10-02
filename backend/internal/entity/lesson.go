package entity

import "time"

const (
	LessonDeliveryOnline  = "online"
	LessonDeliveryOffline = "offline"

	LessonTypeIndividual = "individual"
	LessonTypeGroup      = "group"

	LessonStatusPlanned   = "planned"
	LessonStatusCompleted = "completed"
	LessonStatusCancelled = "cancelled"
)

type Lesson struct {
	ID                 int64
	TeacherOfferID     int64
	TeacherProfileID   int64
	TeacherDisplayName string
	OfferTitle         string
	PriceRubles        *int
	Description        string
	StartsAt           time.Time
	EndsAt             time.Time
	DeliveryFormat     string
	LessonType         string
	Capacity           int
	Status             string
	EnrollmentOpen     bool
	GroupGoal          string
	GroupLevel         string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
