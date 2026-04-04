package dto

import "time"

type CotentReview struct {
	ID               uint64 `gorm:"primaryKey" json:"id"`
	ExecutionID      uint64 `json:"execution_id"`
	Title            string `json:"title"`
	ShortDescription string `json:"short_description"`
	Message          string `json:"message"`
	Status           string `json:"status"`
	Category         string `json:"category"`
	SubCategory      string `json:"sub_category"`
	ImageURL         string `json:"image_url"`
	ImagePrompt      string `json:"image_prompt"`

	CreatedAt time.Time `json:"created"`
	UpdatedAt time.Time `json:"last_updated"`
}
