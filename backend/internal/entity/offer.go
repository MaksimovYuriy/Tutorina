package entity

import "time"

const (
	OfferFormatOnline  = "online"
	OfferFormatOffline = "offline"
	OfferFormatBoth    = "both"
)

type Offer struct {
	ID                     int64
	Title                  string
	Description            string
	Goal                   string
	DefaultDurationMinutes int
	Format                 string
	PriceRubles            *int
	IsPublished            bool
	Teachers               []TeacherOffer
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type TeacherOffer struct {
	ID                 int64
	OfferID            int64
	TeacherProfileID   int64
	TeacherDisplayName string
	DurationMinutes    *int
	PriceRubles        *int
	IsPublished        bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
