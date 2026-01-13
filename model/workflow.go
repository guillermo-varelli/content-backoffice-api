package model

type Workflow struct {
	ID           uint64 `gorm:"primaryKey"`
	Name         string
    Description  string
	Steps        []Step `gorm:"foreignKey:WorkflowID;constraint:OnDelete:CASCADE;"`
}
