package model

type Workflow struct {
	ID          uint64 `gorm:"primaryKey"`
	Name        string `json:"Name"` // Campo para el nombre del workflow
	Description string
	Steps       []Step `gorm:"foreignKey:WorkflowID;constraint:OnDelete:CASCADE;"`
	Enabled     bool   `gorm:"enabled"`
}
