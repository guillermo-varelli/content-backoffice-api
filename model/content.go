package model

import "time"

type Content struct {
	ID               uint64 `gorm:"primaryKey"`
	ExecutionID      uint64
	Title            string
	ShortDescription string
	Message          string
	Status           string
	Type             string
	SubType          string
	Category         string
	SubCategory      string
	ImageURL         string
	ImagePrompt      string
	Slug             string
	Created          time.Time
	LastUpdated      time.Time
}

// ⚠️ IMPORTANTE si la tabla es literalmente "n"
func (Content) TableName() string {
	return "content"
}
