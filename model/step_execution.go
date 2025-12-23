package model

type StepExecution struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	StepID      uint64    `json:"step_id"`
	Step        Step      `gorm:"foreignKey:StepID" json:"step,omitempty"`
	Status      string    `json:"status"`
	ExecutionID uint64    `json:"execution_id"`
	Execution   Execution `gorm:"foreignKey:ExecutionID" json:"execution,omitempty"`
	Output      string    `gorm:"type:longtext" json:"output"`
}
