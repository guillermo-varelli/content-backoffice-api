package model

type Execution struct {
	ID         uint64   `json:"id"`
	WorkflowID uint64   `json:"workflow_id"`
	Workflow   Workflow `json:"workflow"` // Relación con Workflow
	Status     string   `json:"status"`
}
