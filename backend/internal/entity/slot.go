package entity

import "time"

type Slot struct {
	DirectionID int64     `json:"directionId"`
	Level       string    `json:"level"`
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	StartsAt    time.Time `json:"startsAt"`
	EndsAt      time.Time `json:"endsAt"`
	Format      string    `json:"format"`
	Kind        string    `json:"kind"`
	Capacity    int       `json:"capacity"`
	Occupied    int       `json:"occupied"`
	Status      string    `json:"status"`
	Published   bool      `json:"published"`
}
