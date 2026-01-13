package model

type Step struct {
	ID            uint64 `gorm:"primaryKey"`
	OrderIndex    int
	Name          string
	OperationType string
	Prompt        string `gorm:"type:mediumtext"`

	WorkflowID uint64
	Workflow   Workflow `gorm:"foreignKey:WorkflowID;references:ID"`

	AgentID *uint64
	Agent   Agent `gorm:"foreignKey:AgentID;references:ID"`
}
