package dto

type ExecutionResponseDto struct {
	ID       uint64              `json:"id"`
	Status   string              `json:"status"`
	Workflow WorkflowResponseDto `json:"workflow"`
}
