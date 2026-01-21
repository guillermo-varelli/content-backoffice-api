package model

import "time"

type Execution struct {
	ID         uint64    `json:"id"`
	WorkflowID uint64    `json:"workflow_id"`
	Workflow   Workflow  `json:"workflow"` // Relación con Workflow
	Status     string    `json:"status"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"` // Fecha de creación
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"` // Fecha de actualización

}
