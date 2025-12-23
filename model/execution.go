package model

type Execution struct {
	ID         uint64 `gorm:"primaryKey" json:"id"`
	WorkflowID uint64 `json:"workflow_id"`
	Status     string `json:"status"`
	// Agregar más campos según necesites
}
