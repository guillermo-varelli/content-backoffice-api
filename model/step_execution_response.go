package model

type StepExecutionGroupResponse struct {
	ExecutionID uint64          `json:"execution_id"`
	Execution   Execution       `json:"execution"`
	Steps       []StepExecution `json:"steps"`
}
