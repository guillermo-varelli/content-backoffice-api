package model

type Step struct {
    ID            uint64 `gorm:"primaryKey"`
    OrderIndex    int
    Name          string
    OperationType string
    Prompt        string `gorm:"type:mediumtext"`
    WorkflowID    uint64
    AgentID       *uint64
}
