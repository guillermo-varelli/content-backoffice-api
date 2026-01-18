package model

import "time"

type StepExecution struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	StepID      uint64     `json:"step_id"`
	Step        Step       `gorm:"foreignKey:StepID" json:"step,omitempty"`
	Status      string     `json:"status"`
	ExecutionID uint64     `json:"execution_id"`
	Execution   *Execution `gorm:"foreignKey:ExecutionID" json:"execution"`
	Output      string     `gorm:"type:longtext" json:"output"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"` // Fecha de creación
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"` // Fecha de actualización

}
