package model

type WorkflowResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ExecutionResponse struct {
	ID       uint64           `json:"id"`
	Status   string           `json:"status"`
	Workflow WorkflowResponse `json:"workflow"`
}

type StepExecutionGroupResponse struct {
	ExecutionID uint64            `json:"execution_id"`
	Execution   ExecutionResponse `json:"execution"`
	Steps       []StepExecution   `json:"steps"`
}
