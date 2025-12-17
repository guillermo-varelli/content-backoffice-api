package model

import "time"

type N struct {
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
    Created          time.Time
    LastUpdated      time.Time
}
