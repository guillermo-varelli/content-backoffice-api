package dto

import "example.com/workflowapi/model"



type StepResponseDto struct {
	ID            uint64               `json:"id"`
	OrderIndex    int                  `json:"orderIndex"`
	Name          string               `json:"name"`
	OperationType string               `json:"operationType"`
	Prompt        string               `json:"prompt"`
	Workflow      WorkflowResponseDto  `json:"workflow"`
	Agent         AgentResponseDto     `json:"agent"`
}

func ToStepResponse(s model.Step) StepResponseDto {
	return StepResponseDto{
		ID:            s.ID,
		OrderIndex:    s.OrderIndex,
		Name:          s.Name,
		OperationType: s.OperationType,
		Prompt:        s.Prompt,
		Workflow: WorkflowResponseDto{
			ID:   s.Workflow.ID,
			Name: s.Workflow.Name,
			Description: s.Workflow.Description,

		},
		Agent: AgentResponseDto{
			ID:       s.Agent.ID,
			Provider: s.Agent.Provider,
		},
	}
}