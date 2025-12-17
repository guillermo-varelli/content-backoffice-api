package model

type Workflow struct {
    ID   uint64 `gorm:"primaryKey"`
    Name string
    Steps []Step
}
